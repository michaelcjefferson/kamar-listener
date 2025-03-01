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
# GOOS=linux CGO_ENABLED=0 go build -o=./bin/listenerService ./cmd/api
.PHONY: build/linux
build/linux: remove build/frontend
	@echo Building listener service Linux binary...
	go build -tags "linux sqlite_fts5" -o=./bin/listenerService ./cmd/api
	@echo Done

## build/windows: remove previously built binaries and build a windows binary
.PHONY: build/windows
build/windows: remove build/frontend
	@echo Building listener service Windows binary...
	set CGO_ENABLED=0&& set goos=windows&& go build -tags "sqlite_fts5" -o ./bin/listenerService.exe ./cmd/api
	@echo Done

# TODO: the following two commands need to first check for the existance of the folder they're attempting to copy into, and then create it if necessary
## build/templates: build templ files in /ui
.PHONY: build/templates
build/templates:
	@echo 'Generating templ templates...'
	rm -rf ./cmd/api/assets
	cp -r ./ui/assets ./cmd/api
	cd ./ui && templ generate
	@echo 'Done'

## build/frontend: remove previously built frontend files and build new files to cmd/web/ui, to be packaged with go binary
.PHONY: build/frontend
build/frontend:
	@echo Building Svelte frontend for admin dashboard...
	rm -rf ./cmd/api/ui
	cd ./admin-frontend && npm run build
	cp -r ./admin-frontend/static-pages/* ./cmd/api/ui

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
run/api: build/templates
	@echo Running KAMAR Refresh API...
	go run -tags "sqlite_fts5" ./cmd/api

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

# go mod vendor stores the source code for any dependencies in a vendor folder in the project it's run in, preventing any issues that might arise if source code for a dependency is removed from its primary repository and/or any proxy hosts.
# go mod vendor doesn't verify that the checksums of dependencies in the module cache match those in the go.sum file, so it's important to run go mod verify regularly as well, hence the go mod verify rule above go mod vendor.
## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

## clean: clear the cache
.PHONY: clean
clean:
	@echo 'Cleaning go cache...'
	go clean

## remove: run go clean and remove built binaries
.PHONY: remove
remove: confirm clean
	@echo 'Removing built binaries...'
	@if [ -d "./bin" ] && [ "$(ls -A ./bin)" ]; then \
		rm -rf ./bin/*; \
		echo "Binaries removed."; \
	else \
		echo "No binaries to remove."; \
	fi