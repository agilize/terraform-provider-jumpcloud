#!/bin/bash

# Script to list all supported architectures and their SHA256 checksums
# Usage: ./scripts/list-architectures.sh [VERSION]

set -e

# Determine the version if not provided
if [ -z "$1" ]; then
  VERSION=$(grep 'version =' internal/provider/version.go | cut -d'"' -f2)
  if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
    VERSION=${VERSION#v}
  fi
else
  VERSION=$1
fi

echo "JumpCloud Terraform Provider v${VERSION}"
echo "========================================="
echo "Supported Architectures:"
echo

# Check if we're looking at a local build or a published version
if [ -d "dist" ] && [ -f "dist/SHA256SUMS" ]; then
  # Local build
  echo "Using local build from dist directory"
  CHECKSUMS_FILE="dist/SHA256SUMS"
elif [ -n "$GITHUB_REPOSITORY" ]; then
  # GitHub Actions environment
  echo "Using GitHub Actions release artifacts"
  CHECKSUMS_FILE="release/SHA256SUMS"
else
  # Try to download from GitHub
  TEMP_DIR=$(mktemp -d)
  trap 'rm -rf "$TEMP_DIR"' EXIT
  
  echo "Downloading SHA256SUMS from GitHub release v${VERSION}"
  OWNER=${GITHUB_REPOSITORY_OWNER:-ferreirafa}
  
  if ! curl -sSL -o "$TEMP_DIR/SHA256SUMS" "https://github.com/${OWNER}/terraform-provider-jumpcloud/releases/download/v${VERSION}/SHA256SUMS"; then
    echo "Error: Unable to download SHA256SUMS for version v${VERSION}"
    echo "Please check if the release exists or try specifying a different version."
    exit 1
  fi
  
  CHECKSUMS_FILE="$TEMP_DIR/SHA256SUMS"
fi

# List architecture information
echo "| Platform      | Architecture | SHA256 Checksum                                                    |"
echo "|---------------|--------------|-------------------------------------------------------------------|"

grep -i ".zip" "$CHECKSUMS_FILE" | while read -r line; do
  CHECKSUM=$(echo "$line" | awk '{print $1}')
  FILENAME=$(echo "$line" | awk '{print $2}')
  
  # Extract platform and architecture from filename
  if [[ $FILENAME =~ _([^_]+)_([^\.]+)\.zip$ ]]; then
    PLATFORM="${BASH_REMATCH[1]}"
    ARCH="${BASH_REMATCH[2]}"
  elif [[ $FILENAME =~ _([^\.]+)\.zip$ ]]; then
    # Default linux_amd64 case where platform might be omitted
    PLATFORM="linux"
    ARCH="amd64"
  else
    PLATFORM="unknown"
    ARCH="unknown"
  fi
  
  echo "| $PLATFORM | $ARCH | $CHECKSUM |"
done

echo
echo "To use a specific architecture, configure Terraform to use the provider from GitHub Container Registry."
echo "See docs/ghcr-usage.md for detailed instructions." 