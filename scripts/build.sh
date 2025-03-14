#!/bin/bash

# Build the provider
echo "Building provider..."
go build -o terraform-provider-jumpcloud

# Get OS and architecture
OS=$(go env GOOS)
ARCH=$(go env GOARCH)
VERSION=$(grep "VERSION =" Makefile | cut -d '=' -f2 | tr -d ' ')

# Install the provider
echo "Installing provider to ~/.terraform.d/plugins/registry.terraform.io/ferreirafav/jumpcloud/${VERSION}/${OS}_${ARCH}/"
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/ferreirafav/jumpcloud/${VERSION}/${OS}_${ARCH}/
cp terraform-provider-jumpcloud ~/.terraform.d/plugins/registry.terraform.io/ferreirafav/jumpcloud/${VERSION}/${OS}_${ARCH}/

echo "Build and installation complete!" 