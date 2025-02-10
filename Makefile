include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

# NOTE: both help and confirm will not work on Windows.

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# This terminal command will return false if the user responds with anything other than y. Make will stop execution if any rule returns false - this rule failing will prevent sequential rules from executing.
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build/linux: remove previously built binaries and build a linux binary
.PHONY: build/linux
build/linux: remove build/frontend
	@echo Building listener service Linux binary...
	GOOS=linux CGO_ENABLED=0 go build -o=./bin/listenerService ./cmd/api
	@echo Done

## build/windows: remove previously built binaries and build a windows binary
.PHONY: build/windows
build/windows: remove build/frontend
	@echo Building listener service Windows binary...
	set CGO_ENABLED=0&& set goos=windows&& go build -o ./bin/listenerService.exe ./cmd/api
	@echo Done

## build/frontend: remove previously built frontend files and build new files to cmd/web/ui, to be packaged with go binary
.PHONY: build/frontend
build/frontend:
	@echo Building Svelte frontend for admin dashboard...
	rm -rf ./cmd/api/ui
	cd ./admin-frontend && npm run build

#==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/frontend: run vite development server for admin dashboard
.PHONY: run/frontend
run/frontend:
	cd ./admin-frontend && npm run dev

## install/frontend: install npm packages for admin dashboard
.PHONY: install/frontend
install/frontend:
	cd ./admin-frontend && npm install

## run/api: run the listener api service in dev mode
.PHONY: run/api
run/api:
	@echo Running KAMAR Refresh API...
	go run ./cmd/api

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## clean: clear the cache
.PHONY: clean
clean:
	@echo 'Cleaning go cache...'
	go clean

## remove: run go clean and remove built binaries
.PHONY: remove
remove: clean confirm
	@echo 'Removing built binaries...'
	rm -r ./bin/*