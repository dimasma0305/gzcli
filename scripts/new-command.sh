#!/bin/bash

# Script to scaffold a new command
# Usage: ./scripts/new-command.sh <command-name>

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

if [ $# -eq 0 ]; then
    echo -e "${RED}Error: Command name required${NC}"
    echo "Usage: $0 <command-name>"
    echo "Example: $0 status"
    exit 1
fi

COMMAND_NAME=$1
COMMAND_FILE="cmd/${COMMAND_NAME}.go"
TEST_FILE="cmd/${COMMAND_NAME}_test.go"

# Check if command already exists
if [ -f "$COMMAND_FILE" ]; then
    echo -e "${RED}Error: Command '$COMMAND_NAME' already exists${NC}"
    exit 1
fi

echo -e "${BLUE}Creating new command: $COMMAND_NAME${NC}"
echo ""

# Create command file
echo -e "${BLUE}Creating $COMMAND_FILE...${NC}"
cat > "$COMMAND_FILE" << EOF
package cmd

import (
	"github.com/spf13/cobra"
)

var ${COMMAND_NAME}Cmd = &cobra.Command{
	Use:   "${COMMAND_NAME}",
	Short: "Short description of ${COMMAND_NAME}",
	Long: \`Long description of ${COMMAND_NAME}.

Add more detailed description here.

Example:
  gzcli ${COMMAND_NAME}
  gzcli ${COMMAND_NAME} --flag value\`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement ${COMMAND_NAME} command
		cmd.Println("${COMMAND_NAME} command executed")
	},
}

func init() {
	rootCmd.AddCommand(${COMMAND_NAME}Cmd)

	// TODO: Add flags here
	// ${COMMAND_NAME}Cmd.Flags().StringP("flag", "f", "", "Flag description")
}
EOF

echo -e "${GREEN}✓${NC} Created $COMMAND_FILE"

# Create test file
echo -e "${BLUE}Creating $TEST_FILE...${NC}"
cat > "$TEST_FILE" << 'EOF'
package cmd

import (
	"bytes"
	"testing"
)

func Test__Command(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "basic execution",
			args:    []string{},
			wantErr: false,
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(append([]string{"__CMD_NAME__"}, tt.args...))

			// Execute
			err := rootCmd.Execute()

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Reset
			rootCmd.SetArgs([]string{})
		})
	}
}
EOF

# Replace placeholders in test file
sed -i "s/__Command/${COMMAND_NAME^}Command/g" "$TEST_FILE"
sed -i "s/__CMD_NAME__/${COMMAND_NAME}/g" "$TEST_FILE"

echo -e "${GREEN}✓${NC} Created $TEST_FILE"

echo ""
echo -e "${GREEN}✓ Command scaffolding complete!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Implement the command logic in $COMMAND_FILE"
echo "  2. Add tests in $TEST_FILE"
echo "  3. Run tests: go test ./cmd -run Test${COMMAND_NAME^}"
echo "  4. Update documentation if needed"
echo ""
echo -e "${BLUE}The command has been added to rootCmd and is ready to use:${NC}"
echo "  gzcli ${COMMAND_NAME}"
echo ""
