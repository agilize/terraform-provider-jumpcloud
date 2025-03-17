#!/bin/bash

# Script to run golangci-lint with the correct settings
# This helps ensure consistent execution between local development and CI

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running golangci-lint checks...${NC}"

# Path to the golangci-lint binary
LINTER_BIN="$HOME/go/bin/golangci-lint"

# Check if binary exists
if [ ! -f "$LINTER_BIN" ]; then
    echo -e "${RED}golangci-lint not found. Installing...${NC}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.0
    LINTER_BIN="$HOME/go/bin/golangci-lint"
fi

# Get the version
VERSION=$($LINTER_BIN --version | head -n 1)
echo -e "${GREEN}Using $VERSION${NC}"

# Run the lint check with the same settings as CI
echo -e "${YELLOW}Running lint checks with configuration from .golangci.yml...${NC}"
$LINTER_BIN run \
  --timeout=5m \
  --disable=goanalysis_metalinter \
  --config=.golangci.yml

# Check exit code
EXIT_CODE=$?
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "\n${GREEN}Lint checks passed successfully!${NC}"
else
    echo -e "\n${RED}Lint checks failed with exit code $EXIT_CODE.${NC}"
    echo -e "${YELLOW}Please fix the issues above before committing.${NC}"
fi

exit $EXIT_CODE 