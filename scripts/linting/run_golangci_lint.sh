#!/bin/bash

echo "Running golangci-lint with project configuration"
echo "==============================================="

# Define o caminho para o golangci-lint
GOLANGCI_LINT="$HOME/go/bin/golangci-lint"

# Check if golangci-lint is installed
if [ ! -f "$GOLANGCI_LINT" ]; then
    echo "Error: golangci-lint is not installed at $GOLANGCI_LINT"
    echo "Please install it: https://golangci-lint.run/usage/install/"
    exit 1
fi

# Run golangci-lint
echo "Running linting checks..."
"$GOLANGCI_LINT" run --timeout=5m --config=.golangci.yml --verbose --out-format=colored-line-number "$@"

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