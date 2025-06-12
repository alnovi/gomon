.SILENT:
.DEFAULT-GOAL:= help

## help: справка
.PHONY: help
help:
	@echo 'COMMON'
	@echo ''
	@echo 'Usage:'
	@echo '  make <command>'
	@echo ''
	@echo 'The commands are:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## lint: статический анализ
.PHONY: lint
lint:
	go tool golangci-lint run ./...

## lint-fix: статический анализ (авто-исправление)
.PHONY: lint-fix
lint-fix:
	go tool golangci-lint run ./... --fix --timeout 650s
