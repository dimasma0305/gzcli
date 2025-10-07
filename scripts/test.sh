#!/bin/bash

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Running comprehensive test suite...${NC}"
echo ""

# Run unit tests
echo -e "${BLUE}→ Running unit tests...${NC}"
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
echo -e "${BLUE}→ Generating coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')

echo ""
echo -e "${GREEN}✓ Tests completed!${NC}"
echo -e "  Coverage: ${COVERAGE}"
echo -e "  Report: coverage.html"
echo ""

# Check coverage threshold (optional)
THRESHOLD=50
COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
if (( $(echo "$COVERAGE_NUM < $THRESHOLD" | bc -l) )); then
    echo -e "${RED}⚠ Coverage is below ${THRESHOLD}%${NC}"
    exit 1
fi

echo -e "${GREEN}✓ All tests passed!${NC}"

