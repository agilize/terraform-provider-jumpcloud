#!/bin/bash

echo "Running golangci-lint with project configuration"
echo "==============================================="

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: golangci-lint is not installed or not in PATH"
    echo "Please install it: https://golangci-lint.run/usage/install/"
    exit 1
fi

# Run golangci-lint
echo "Running linting checks..."
golangci-lint run --timeout=5m --config=.golangci.yml --verbose --out-format=colored-line-number "$@"

exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "✅ Linting passed successfully!"
else
    echo "❌ Linting found issues that need to be fixed."
    echo "🔍 See the errors above for details."
    echo
    echo "To disable certain linters or specific checks, edit the .golangci.yml file."
fi

exit $exit_code 