.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: vet
## vet: run go vet
vet:
	@go vet ./...

.PHONY: test
## test: run tests
test: vet
	@go test -cover ./...

.PHONY: lint
## lint: run golangci-lint
# Install: https://golangci-lint.run/usage/install/
lint:
	@golangci-lint run ./... --out-format colored-line-number
