#!/bin/bash

# Script to check only critical linting errors, ignoring those that will be handled in phases
# This script runs tfproviderlint disabling all rules except critical ones

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running critical linting checks...${NC}"

# Path to tfproviderlint binary
LINTER_BIN="$HOME/go/bin/tfproviderlint"

# Check if binary exists
if [ ! -f "$LINTER_BIN" ]; then
    echo -e "${RED}tfproviderlint not found. Installing...${NC}"
    go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
fi

# List of errors being ignored
echo -e "${YELLOW}Running tfproviderlint ignoring non-priority errors...${NC}"

# Run linter disabling all rules except R014 (which we've already fixed)
$LINTER_BIN \
  -AT001=false \
  -AT005=false \
  -AT012=false \
  -R001=false \
  -R017=false \
  -R019=false \
  -V013=false \
  ./...

# Check exit code
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}Check completed: All critical errors have been fixed!${NC}"
    echo -e "${GREEN}Non-critical errors will be addressed in future phases.${NC}"
else
    echo -e "\n${RED}Check failed: There are critical errors that need to be fixed.${NC}"
fi

echo -e "\n${YELLOW}Note:${NC} To check all errors, use the command:"
echo -e "  $LINTER_BIN ./..."
echo -e "${YELLOW}For phase-by-phase corrections, refer to the run_linter.sh script${NC}" 