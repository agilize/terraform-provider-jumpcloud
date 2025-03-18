#!/bin/bash

# Script to publish the provider to GitHub Container Registry locally

set -e

# Determine the version
VERSION=$(grep 'version =' internal/provider/version.go | cut -d'"' -f2)
if [ -z "$VERSION" ]; then
  VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.1.0")
  VERSION=${VERSION#v}
fi

echo "Publishing provider version $VERSION to GitHub Container Registry"

# Compile for supported platforms
echo "Compiling binaries..."
mkdir -p dist

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o dist/terraform-provider-jumpcloud_v${VERSION}
# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o dist/terraform-provider-jumpcloud_v${VERSION}_linux_arm64
# MacOS AMD64
GOOS=darwin GOARCH=amd64 go build -o dist/terraform-provider-jumpcloud_v${VERSION}_darwin_amd64
# MacOS ARM64
GOOS=darwin GOARCH=arm64 go build -o dist/terraform-provider-jumpcloud_v${VERSION}_darwin_arm64
# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o dist/terraform-provider-jumpcloud_v${VERSION}_windows_amd64.exe

# Create ZIPs
echo "Creating ZIP files..."
cd dist
for file in terraform-provider-jumpcloud_v${VERSION}*; do
  platform=${file#terraform-provider-jumpcloud_v${VERSION}}
  platform=${platform%.*}
  if [ -z "$platform" ]; then
    platform="_linux_amd64"
  fi
  zip -j terraform-provider-jumpcloud_${VERSION}${platform}.zip $file
done
cd ..

# Generate checksums
echo "Generating checksums..."
cd dist
sha256sum *.zip > SHA256SUMS
cd ..

# Ensure Docker BuildX is available
echo "Setting up Docker BuildX..."
docker buildx inspect --bootstrap

# Build and publish Docker image
echo "Building and publishing Docker image..."
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/${GITHUB_USER:-ferreirafa}/terraform-provider-jumpcloud:v${VERSION} \
  -t ghcr.io/${GITHUB_USER:-ferreirafa}/terraform-provider-jumpcloud:latest \
  --build-arg VERSION=${VERSION} \
  --push \
  .

echo "Logging in to GitHub Container Registry..."
echo "Please provide your GitHub personal access token with packages:write permission"
echo "Or press Enter to skip this step if you're already authenticated."
read -s GITHUB_TOKEN

if [ -n "$GITHUB_TOKEN" ]; then
  echo $GITHUB_TOKEN | docker login ghcr.io -u ${GITHUB_USER:-ferreirafa} --password-stdin
fi

echo "Provider published successfully!"
echo "To use it, add to your ~/.terraformrc:"
echo "
provider_installation {
  network_mirror {
    url = \"https://ghcr.io/${GITHUB_USER:-ferreirafa}/terraform-provider-jumpcloud\"
    include = [\"ghcr.io/${GITHUB_USER:-ferreirafa}/jumpcloud\"]
  }
  direct {
    exclude = [\"ghcr.io/${GITHUB_USER:-ferreirafa}/jumpcloud\"]
  }
}
"

echo ""
echo "This container includes provider binaries for ALL supported platforms:"
echo "- Linux AMD64 & ARM64"
echo "- MacOS AMD64 & ARM64"
echo "- Windows AMD64" 