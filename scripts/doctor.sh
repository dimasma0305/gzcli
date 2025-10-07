#!/bin/bash

# Development Environment Doctor
# Diagnoses common issues in the development environment

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

ERRORS=0
WARNINGS=0

echo -e "${BLUE}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║     gzcli Development Environment Doctor            ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to check command existence
check_command() {
    local cmd=$1
    local required=$2
    local install_hint=$3

    if command -v "$cmd" &> /dev/null; then
        local version=$($cmd version 2>&1 | head -n 1 || echo "unknown")
        echo -e "${GREEN}✓${NC} $cmd is installed ($version)"
        return 0
    else
        if [ "$required" = "true" ]; then
            echo -e "${RED}✗${NC} $cmd is NOT installed (required)"
            echo -e "  ${YELLOW}Install with: $install_hint${NC}"
            ((ERRORS++))
        else
            echo -e "${YELLOW}⚠${NC} $cmd is NOT installed (optional)"
            echo -e "  ${BLUE}Install with: $install_hint${NC}"
            ((WARNINGS++))
        fi
        return 1
    fi
}

# Check Go installation
echo -e "${BLUE}[1/10] Checking Go installation...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}✓${NC} Go is installed (version $GO_VERSION)"

    # Check Go version (should be 1.23 or later)
    MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
    MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

    if [ "$MAJOR" -ge 1 ] && [ "$MINOR" -ge 23 ]; then
        echo -e "${GREEN}✓${NC} Go version is compatible"
    else
        echo -e "${YELLOW}⚠${NC} Go version might be too old (1.23+ recommended)"
        ((WARNINGS++))
    fi
else
    echo -e "${RED}✗${NC} Go is NOT installed"
    echo -e "  ${YELLOW}Install from: https://golang.org/dl/${NC}"
    ((ERRORS++))
fi
echo ""

# Check Git
echo -e "${BLUE}[2/10] Checking Git...${NC}"
check_command "git" "true" "https://git-scm.com/downloads"
echo ""

# Check Make
echo -e "${BLUE}[3/10] Checking Make...${NC}"
check_command "make" "false" "apt install make (Debian/Ubuntu) or brew install make (macOS)"
echo ""

# Check Development Tools
echo -e "${BLUE}[4/10] Checking development tools...${NC}"
check_command "golangci-lint" "false" "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
check_command "goimports" "false" "go install golang.org/x/tools/cmd/goimports@latest"
check_command "goreleaser" "false" "go install github.com/goreleaser/goreleaser@latest"
check_command "air" "false" "go install github.com/air-verse/air@latest"
echo ""

# Check GOPATH
echo -e "${BLUE}[5/10] Checking GOPATH...${NC}"
if [ -n "$GOPATH" ]; then
    echo -e "${GREEN}✓${NC} GOPATH is set: $GOPATH"
else
    DEFAULT_GOPATH=$(go env GOPATH)
    echo -e "${YELLOW}⚠${NC} GOPATH not set, using default: $DEFAULT_GOPATH"
    ((WARNINGS++))
fi
echo ""

# Check if binary is in PATH
echo -e "${BLUE}[6/10] Checking PATH configuration...${NC}"
GOBIN=$(go env GOPATH)/bin
if echo "$PATH" | grep -q "$GOBIN"; then
    echo -e "${GREEN}✓${NC} \$GOPATH/bin is in PATH"
else
    echo -e "${YELLOW}⚠${NC} \$GOPATH/bin is NOT in PATH"
    echo -e "  ${BLUE}Add to your shell profile: export PATH=\$PATH:\$GOPATH/bin${NC}"
    ((WARNINGS++))
fi
echo ""

# Check dependencies
echo -e "${BLUE}[7/10] Checking Go dependencies...${NC}"
if [ -f go.mod ]; then
    echo -e "${GREEN}✓${NC} go.mod exists"

    if go mod verify &> /dev/null; then
        echo -e "${GREEN}✓${NC} Dependencies verified"
    else
        echo -e "${YELLOW}⚠${NC} Dependencies need updating"
        echo -e "  ${BLUE}Run: go mod download${NC}"
        ((WARNINGS++))
    fi
else
    echo -e "${RED}✗${NC} go.mod not found (are you in the project root?)"
    ((ERRORS++))
fi
echo ""

# Check network connectivity (optional)
echo -e "${BLUE}[8/10] Checking network connectivity (optional)...${NC}"
if command -v curl &> /dev/null || command -v wget &> /dev/null; then
    echo -e "${GREEN}✓${NC} Network tools available (curl/wget)"
else
    echo -e "${BLUE}ℹ${NC} No network tools found (optional, but helpful for API testing)"
fi
echo ""

# Check git hooks
echo -e "${BLUE}[9/10] Checking git hooks...${NC}"
if [ -f .git/hooks/pre-commit ]; then
    echo -e "${GREEN}✓${NC} Git pre-commit hook is installed"
else
    echo -e "${YELLOW}⚠${NC} Git hooks not installed"
    echo -e "  ${BLUE}Install with: make hooks${NC}"
    ((WARNINGS++))
fi
echo ""

# Check build
echo -e "${BLUE}[10/10] Checking if project builds...${NC}"
if go build -o /tmp/gzcli-test . &> /tmp/gzcli-build.log; then
    echo -e "${GREEN}✓${NC} Project builds successfully"
    rm -f /tmp/gzcli-test
else
    echo -e "${RED}✗${NC} Project build failed"
    echo -e "  ${YELLOW}See: /tmp/gzcli-build.log${NC}"
    ((ERRORS++))
fi
echo ""

# Port checks (optional)
echo -e "${BLUE}[Bonus] Checking common ports...${NC}"
if command -v lsof &> /dev/null || command -v netstat &> /dev/null; then
    PORTS_TO_CHECK=(8080 8081)
    for port in "${PORTS_TO_CHECK[@]}"; do
        if command -v lsof &> /dev/null; then
            if lsof -i ":$port" &> /dev/null; then
                echo -e "${YELLOW}⚠${NC} Port $port is in use"
            else
                echo -e "${GREEN}✓${NC} Port $port is available"
            fi
        fi
    done
else
    echo -e "${BLUE}ℹ${NC} Port check tools not available (lsof/netstat)"
fi
echo ""

# Summary
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ Everything looks good! No issues found.${NC}"
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ Found $WARNINGS warning(s). Development should work, but some features may be limited.${NC}"
    exit 0
else
    echo -e "${RED}✗ Found $ERRORS error(s) and $WARNINGS warning(s).${NC}"
    echo -e "${YELLOW}Please fix the errors before continuing development.${NC}"
    exit 1
fi
