#!/bin/bash

echo "Running full linting checks - this will show all errors"
echo "Refer to docs/LINTING.md for the planned phases for fixing these errors"

# Run tfproviderlint with all checks enabled
tfproviderlint ./...

echo "The errors shown above will be fixed in phases according to docs/LINTING.md" 