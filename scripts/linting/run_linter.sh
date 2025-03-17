#!/bin/bash

# Script to run tfproviderlint with custom settings
# Allows temporarily ignoring certain types of errors while they are fixed in phases

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running custom lint checks...${NC}"

# Path to tfproviderlint binary
LINTER_BIN="$HOME/go/bin/tfproviderlint"

# Check if binary exists
if [ ! -f "$LINTER_BIN" ]; then
    echo -e "${RED}tfproviderlint not found. Installing...${NC}"
    go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
fi

# Check R014 (errors we've already fixed)
echo -e "${YELLOW}Checking rule R014 (interface{} must be named 'meta')...${NC}"
$LINTER_BIN -R014=true ./...

# If you want to check other specific errors, uncomment and run
# For example, to check only R001:
# echo -e "${YELLOW}Checking rule R001...${NC}"
# $LINTER_BIN -R001=true -AT=false -R=false -S=false -V=false ./...

# List of errors we're temporarily ignoring
echo -e "\n${YELLOW}Temporarily ignored errors (to be fixed in phases):${NC}"
echo -e "- ${YELLOW}AT001${NC}: missing CheckDestroy"
echo -e "- ${YELLOW}AT005${NC}: acceptance test function name should begin with TestAcc"
echo -e "- ${YELLOW}AT012${NC}: file contains multiple acceptance test name prefixes"
echo -e "- ${YELLOW}R001${NC}: ResourceData.Set() key argument should be string literal"
echo -e "- ${YELLOW}R017${NC}: schema attributes should be stable across Terraform runs"
echo -e "- ${YELLOW}R019${NC}: d.HasChanges() has many arguments, consider d.HasChangesExcept()"
echo -e "- ${YELLOW}V013${NC}: custom SchemaValidateFunc should be replaced with validation.StringInSlice()"

echo -e "\n${GREEN}To check a specific error, run:${NC}"
echo -e "  $LINTER_BIN -AT=false -R=false -S=false -V=false -<RULE>=true ./..."
echo -e "  Example: $LINTER_BIN -AT=false -R=false -S=false -V=false -R001=true ./..."

echo -e "\n${GREEN}To check all errors, run:${NC}"
echo -e "  $LINTER_BIN ./..."

# Example of next steps for correction
echo -e "\n${YELLOW}Suggested plan for phased correction:${NC}"
echo -e "1. Fix R001 (ResourceData.Set with string literal)"
echo -e "2. Fix R019 (HasChanges → HasChangesExcept)"
echo -e "3. Fix V013 (SchemaValidateFunc → validation.StringInSlice)"
echo -e "4. Fix R017 (Schema attributes should be stable)"
echo -e "5. Fix AT* (acceptance test issues)" 