#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running linters...${NC}"
echo ""

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}golangci-lint is not installed${NC}"
    echo "Install with: make tools"
    exit 1
fi

# Run go fmt
echo -e "${BLUE}→ Running gofmt...${NC}"
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo -e "${RED}The following files are not formatted:${NC}"
    echo "$UNFORMATTED"
    echo ""
    echo "Run 'make fmt' to format them"
    exit 1
fi
echo -e "${GREEN}✓ All files are formatted${NC}"

# Run go vet
echo -e "${BLUE}→ Running go vet...${NC}"
go vet ./...
echo -e "${GREEN}✓ go vet passed${NC}"

# Run golangci-lint
echo -e "${BLUE}→ Running golangci-lint...${NC}"
golangci-lint run ./...
echo -e "${GREEN}✓ golangci-lint passed${NC}"

echo ""
echo -e "${GREEN}✓ All linters passed!${NC}"

