#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Setting up gzcli development environment...${NC}"
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}Go is not installed. Please install Go 1.23 or later.${NC}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}✓${NC} Go ${GO_VERSION} is installed"

# Download dependencies
echo -e "${BLUE}Downloading dependencies...${NC}"
go mod download
echo -e "${GREEN}✓${NC} Dependencies downloaded"

# Install development tools
echo -e "${BLUE}Installing development tools...${NC}"

tools=(
    "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    "golang.org/x/tools/cmd/goimports@latest"
    "github.com/goreleaser/goreleaser@latest"
    "github.com/air-verse/air@latest"
)

for tool in "${tools[@]}"; do
    echo "  Installing ${tool}..."
    go install "${tool}"
done

echo -e "${GREEN}✓${NC} Development tools installed"

# Install git hooks
echo -e "${BLUE}Installing git hooks...${NC}"
if [ -f scripts/install-hooks.sh ]; then
    bash scripts/install-hooks.sh
fi

# Build the project
echo -e "${BLUE}Building the project...${NC}"
make build
echo -e "${GREEN}✓${NC} Build successful"

echo ""
echo -e "${GREEN}✓ Development environment setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Run 'make test' to run tests"
echo "  2. Run 'make dev' to start development mode"
echo "  3. Run 'make help' to see all available commands"
echo ""

