#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing git hooks...${NC}"

# Check if pre-commit framework is installed
if command -v pre-commit &> /dev/null; then
    echo -e "${BLUE}Installing pre-commit hooks...${NC}"
    pre-commit install
    pre-commit install --hook-type commit-msg
    echo -e "${GREEN}✓ Pre-commit framework hooks installed!${NC}"
    echo ""
    echo "Pre-commit will run:"
    echo "  - Trailing whitespace removal"
    echo "  - EOF fixes"
    echo "  - YAML/JSON validation"
    echo "  - Go formatting (gofmt, goimports)"
    echo "  - Go vet"
    echo "  - Go linting (golangci-lint)"
    echo "  - Unit tests (short mode)"
    echo ""
    echo "To skip hooks temporarily, use: git commit --no-verify"
    echo "To run hooks manually, use: pre-commit run --all-files"
else
    echo -e "${YELLOW}⚠ pre-commit framework not found. Installing basic git hooks...${NC}"
    echo ""

    # Create hooks directory if it doesn't exist
    mkdir -p .git/hooks

    # Pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

echo "Running pre-commit checks..."

# Format code
echo "→ Formatting code..."
make fmt

# Run linters
echo "→ Running linters..."
if ! make lint; then
    echo "❌ Linting failed. Please fix the issues before committing."
    exit 1
fi

# Run tests
echo "→ Running tests..."
if ! make test-unit; then
    echo "❌ Tests failed. Please fix the tests before committing."
    exit 1
fi

echo "✅ Pre-commit checks passed!"
EOF

    # Make hooks executable
    chmod +x .git/hooks/pre-commit

    echo -e "${GREEN}✓ Basic git hooks installed successfully!${NC}"
    echo ""
    echo "Pre-commit hook will run:"
    echo "  - Code formatting"
    echo "  - Linters"
    echo "  - Unit tests"
    echo ""
    echo "To skip hooks temporarily, use: git commit --no-verify"
    echo ""
    echo -e "${BLUE}Tip: Install pre-commit framework for better hook management:${NC}"
    echo "  pip install pre-commit"
    echo "  pre-commit install"
fi
