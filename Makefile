include .envrc

build:
	@echo Building binary...
	set CGO_ENABLED=0&& set goos=windows&& go build -username ${KDS_USERNAME} -password ${KDS_PASSWORD} -o ./bin/kamarRefresh.exe ./cmd/api
	@echo Done

run:
	@echo Running KAMAR Refresh project...
	go run ./cmd/api