#!/bin/bash

# Benchmark Script
# Runs benchmarks and optionally compares with previous runs

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

BENCH_DIR=".benchmarks"
CURRENT_BENCH="$BENCH_DIR/current.txt"
PREVIOUS_BENCH="$BENCH_DIR/previous.txt"

# Parse arguments
COMPARE=false
SAVE=false
PACKAGE="./..."
BENCHTIME="1s"

while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--compare)
            COMPARE=true
            shift
            ;;
        -s|--save)
            SAVE=true
            shift
            ;;
        -p|--package)
            PACKAGE="$2"
            shift 2
            ;;
        -t|--benchtime)
            BENCHTIME="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -c, --compare      Compare with previous benchmark"
            echo "  -s, --save         Save results as baseline"
            echo "  -p, --package PKG  Benchmark specific package (default: ./...)"
            echo "  -t, --benchtime T  Run each benchmark for T time (default: 1s)"
            echo "  -h, --help         Show this help"
            echo ""
            echo "Examples:"
            echo "  $0                              # Run benchmarks"
            echo "  $0 --save                       # Run and save as baseline"
            echo "  $0 --compare                    # Run and compare with baseline"
            echo "  $0 --package ./internal/gzcli   # Benchmark specific package"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Create benchmark directory
mkdir -p "$BENCH_DIR"

echo -e "${BLUE}╔══════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              gzcli Benchmark Runner                  ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════╝${NC}"
echo ""

# Run benchmarks
echo -e "${BLUE}Running benchmarks on $PACKAGE...${NC}"
echo -e "${BLUE}Benchmark time: $BENCHTIME${NC}"
echo ""

go test -bench=. -benchmem -benchtime="$BENCHTIME" "$PACKAGE" | tee "$CURRENT_BENCH"

echo ""

# Save if requested
if [ "$SAVE" = true ]; then
    cp "$CURRENT_BENCH" "$PREVIOUS_BENCH"
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    cp "$CURRENT_BENCH" "$BENCH_DIR/bench_$TIMESTAMP.txt"
    echo -e "${GREEN}✓ Benchmark results saved as baseline${NC}"
    echo -e "${BLUE}  Baseline: $PREVIOUS_BENCH${NC}"
    echo -e "${BLUE}  Archive: $BENCH_DIR/bench_$TIMESTAMP.txt${NC}"
    echo ""
fi

# Compare if requested and previous exists
if [ "$COMPARE" = true ]; then
    if [ ! -f "$PREVIOUS_BENCH" ]; then
        echo -e "${YELLOW}⚠ No previous benchmark found${NC}"
        echo -e "${BLUE}Run with --save to create a baseline first${NC}"
        exit 0
    fi

    echo -e "${BLUE}Comparing with previous benchmark...${NC}"
    echo ""

    # Check if benchstat is available
    if command -v benchstat &> /dev/null; then
        benchstat "$PREVIOUS_BENCH" "$CURRENT_BENCH"
    else
        echo -e "${YELLOW}⚠ benchstat not installed${NC}"
        echo -e "${BLUE}Install with: go install golang.org/x/perf/cmd/benchstat@latest${NC}"
        echo ""
        echo -e "${BLUE}Basic comparison:${NC}"
        echo ""
        echo -e "${BLUE}Previous:${NC}"
        grep "Benchmark" "$PREVIOUS_BENCH" | head -n 10
        echo ""
        echo -e "${BLUE}Current:${NC}"
        grep "Benchmark" "$CURRENT_BENCH" | head -n 10
    fi
fi

# Show summary
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Summary${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════${NC}"

# Count benchmarks
BENCH_COUNT=$(grep -c "^Benchmark" "$CURRENT_BENCH" || echo "0")
echo -e "Total benchmarks run: ${GREEN}$BENCH_COUNT${NC}"

# Extract some key metrics if available
if grep -q "ns/op" "$CURRENT_BENCH"; then
    echo ""
    echo -e "${BLUE}Top 5 slowest operations:${NC}"
    grep "ns/op" "$CURRENT_BENCH" | sort -k3 -rn | head -n 5 | while read -r line; do
        echo "  $line"
    done
fi

echo ""
echo -e "${GREEN}✓ Benchmark complete${NC}"
echo ""
echo -e "${BLUE}Benchmark results saved to: $CURRENT_BENCH${NC}"

if [ "$SAVE" != true ] && [ "$COMPARE" != true ]; then
    echo ""
    echo -e "${BLUE}Tip: Use --save to save this as baseline for future comparisons${NC}"
    echo -e "${BLUE}     Use --compare to compare with the baseline${NC}"
fi

echo ""
