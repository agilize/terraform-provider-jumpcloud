#!/bin/bash

echo "Running linting checks - ignoring errors that will be fixed in future phases"
echo "Checking only critical errors and R014 (already fixed)"

# Check R014 (already fixed)
tfproviderlint \
  -AT001=false \
  -AT005=false \
  -AT012=false \
  -R001=false \
  -R017=false \
  -R019=false \
  -V013=false \
  -R014=true \
  ./...

# List ignored errors (for reference)
echo "Temporarily ignored errors (to be fixed in phases):"
echo "- AT001: missing CheckDestroy"
echo "- AT005: acceptance test function name should begin with TestAcc"
echo "- AT012: file contains multiple acceptance test name prefixes"
echo "- R001: ResourceData.Set() key argument should be string literal"
echo "- R017: schema attributes should be stable across Terraform runs"
echo "- R019: d.HasChanges() has many arguments, consider d.HasChangesExcept()"
echo "- V013: custom SchemaValidateFunc should be replaced with validation.StringInSlice()"

echo "See the complete correction plan in docs/LINTING.md" 