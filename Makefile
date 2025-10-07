.PHONY: all build install clean dev fmt lint vet test test-coverage test-race release help

# Variables
BINARY_NAME=gzcli
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"
GOPATH?=$(shell go env GOPATH)

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

all: build

## help: Display this help message
help:
	@echo "${BLUE}gzcli Makefile${NC}"
	@echo ""
	@echo "${GREEN}Available targets:${NC}"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${BLUE}%-18s${NC} %s\n", $$1, $$2 } /^##@/ { printf "\n${GREEN}%s${NC}\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

## build: Build the binary
build:
	@echo "${BLUE}Building ${BINARY_NAME}...${NC}"
	@go build ${LDFLAGS} -o ${BINARY_NAME} .
	@echo "${GREEN}Build complete: ${BINARY_NAME}${NC}"

## install: Install binary to $GOPATH/bin
install:
	@echo "${BLUE}Installing ${BINARY_NAME} to ${GOPATH}/bin...${NC}"
	@go install ${LDFLAGS} .
	@echo "${GREEN}Installation complete${NC}"

## clean: Clean build artifacts
clean:
	@echo "${BLUE}Cleaning build artifacts...${NC}"
	@rm -f ${BINARY_NAME}
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@go clean
	@echo "${GREEN}Clean complete${NC}"

##@ Development

## dev: Run in development mode with auto-reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "${RED}air not installed. Install with: go install github.com/air-verse/air@latest${NC}"; \
		echo "${BLUE}Running without auto-reload...${NC}"; \
		go run ${LDFLAGS} . ; \
	fi

## fmt: Format code with gofmt and goimports
fmt:
	@echo "${BLUE}Formatting code...${NC}"
	@gofmt -s -w .
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "${RED}goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest${NC}"; \
	fi
	@echo "${GREEN}Format complete${NC}"

## lint: Run golangci-lint
lint:
	@echo "${BLUE}Running linters...${NC}"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
		echo "${GREEN}Lint complete${NC}"; \
	else \
		echo "${RED}golangci-lint not installed${NC}"; \
		echo "Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.5.0"; \
		exit 1; \
	fi

## vet: Run go vet
vet:
	@echo "${BLUE}Running go vet...${NC}"
	@go vet ./...
	@echo "${GREEN}Vet complete${NC}"

##@ Testing

## test: Run tests
test:
	@echo "${BLUE}Running tests...${NC}"
	@go test -v ./...
	@echo "${GREEN}Tests complete${NC}"

## test-unit: Run unit tests only (short mode)
test-unit:
	@echo "${BLUE}Running unit tests...${NC}"
	@go test -short -v ./...
	@echo "${GREEN}Unit tests complete${NC}"

## test-integration: Run integration tests
test-integration:
	@echo "${BLUE}Running integration tests...${NC}"
	@go test -run Integration -v ./...
	@echo "${GREEN}Integration tests complete${NC}"

## test-coverage: Generate test coverage report
test-coverage:
	@echo "${BLUE}Generating coverage report...${NC}"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report generated: coverage.html${NC}"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

## test-race: Run tests with race detector
test-race:
	@echo "${BLUE}Running tests with race detector...${NC}"
	@go test -race -v ./...
	@echo "${GREEN}Race tests complete${NC}"

## test-watcher: Run watcher-specific tests
test-watcher:
	@echo "${BLUE}Running watcher tests...${NC}"
	@go test -v ./internal/gzcli/watcher/...
	@echo "${GREEN}Watcher tests complete${NC}"

## test-challenge: Run challenge-specific tests
test-challenge:
	@echo "${BLUE}Running challenge tests...${NC}"
	@go test -v ./internal/gzcli/challenge/...
	@echo "${GREEN}Challenge tests complete${NC}"

## test-api: Run API client tests
test-api:
	@echo "${BLUE}Running API client tests...${NC}"
	@go test -v ./internal/gzcli/gzapi/...
	@echo "${GREEN}API client tests complete${NC}"

## test-watch: Watch tests and re-run on changes (requires air)
test-watch:
	@if command -v air > /dev/null; then \
		air -c .air.test.toml; \
	else \
		echo "${RED}air not installed. Install with: go install github.com/air-verse/air@latest${NC}"; \
		echo "${BLUE}Falling back to single test run...${NC}"; \
		make test; \
	fi

## bench: Run benchmarks
bench:
	@echo "${BLUE}Running benchmarks...${NC}"
	@go test -bench=. -benchmem ./...
	@echo "${GREEN}Benchmarks complete${NC}"

##@ Release

## release: Build for multiple platforms using goreleaser
release:
	@echo "${BLUE}Building release with goreleaser...${NC}"
	@if command -v goreleaser > /dev/null; then \
		goreleaser release --snapshot --clean; \
		echo "${GREEN}Release build complete: dist/${NC}"; \
	else \
		echo "${RED}goreleaser not installed${NC}"; \
		echo "Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

##@ Dependencies

## deps: Download dependencies
deps:
	@echo "${BLUE}Downloading dependencies...${NC}"
	@go mod download
	@echo "${GREEN}Dependencies downloaded${NC}"

## deps-update: Update dependencies
deps-update:
	@echo "${BLUE}Updating dependencies...${NC}"
	@go get -u ./...
	@go mod tidy
	@echo "${GREEN}Dependencies updated${NC}"

## tools: Install development tools
tools:
	@echo "${BLUE}Installing development tools...${NC}"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/goreleaser/goreleaser@latest
	@go install github.com/air-verse/air@latest
	@echo "${GREEN}Tools installed${NC}"

##@ Git Hooks

## hooks: Install git hooks
hooks:
	@echo "${BLUE}Installing git hooks...${NC}"
	@if [ -f scripts/install-hooks.sh ]; then \
		bash scripts/install-hooks.sh; \
	else \
		echo "${RED}scripts/install-hooks.sh not found${NC}"; \
	fi

##@ Miscellaneous

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "${GREEN}All checks passed!${NC}"

## ci: Run CI checks (vet, lint, test, test-race)
ci: vet lint test test-race
	@echo "${GREEN}CI checks complete!${NC}"

## doctor: Diagnose development environment issues
doctor:
	@bash scripts/doctor.sh

## setup-complete: Complete setup with verification
setup-complete:
	@bash scripts/setup.sh
	@echo "${BLUE}Running verification...${NC}"
	@make doctor

## quick-test: Run fast smoke tests
quick-test:
	@echo "${BLUE}Running quick smoke tests...${NC}"
	@go test -short -run "^Test.*_Unit$$" ./...
	@echo "${GREEN}Quick tests complete${NC}"

## coverage-browse: Open coverage report in browser
coverage-browse: test-coverage
	@echo "${BLUE}Opening coverage report in browser...${NC}"
	@if command -v xdg-open > /dev/null; then \
		xdg-open coverage.html; \
	elif command -v open > /dev/null; then \
		open coverage.html; \
	elif command -v start > /dev/null; then \
		start coverage.html; \
	else \
		echo "${YELLOW}Could not detect browser opener. Please open coverage.html manually.${NC}"; \
	fi

##@ Testing Environment

## test-env-init: Initialize .test folder structure for development
test-env-init:
	@echo "${BLUE}Initializing test environment...${NC}"
	@mkdir -p .test
	@cd .test && ../gzcli init
	@echo "${GREEN}Test environment initialized${NC}"
	@echo "${BLUE}Next steps:${NC}"
	@echo "  1. Edit .test/.gzctf/conf.yaml with your GZCTF platform settings"
	@echo "  2. Configure the URL to point to your running GZCTF instance"

## test-env-clean: Clean all test environment data
test-env-clean:
	@echo "${RED}This will remove all test data. Continue? [y/N]${NC}" && read ans && [ $${ans:-N} = y ]
	@echo "${BLUE}Cleaning test environment...${NC}"
	@rm -rf .test/
	@echo "${GREEN}Test environment cleaned${NC}"
