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

# Create Dockerfile with platform args
echo "Creating Dockerfile..."
cat > Dockerfile << EOF
FROM --platform=\${BUILDPLATFORM} alpine:3.17

ARG TARGETPLATFORM
ARG BUILDPLATFORM

LABEL org.opencontainers.image.source=https://github.com/agilize/terraform-provider-jumpcloud
LABEL org.opencontainers.image.description="JumpCloud Terraform Provider v${VERSION} - Platform-specific build for \${TARGETPLATFORM}"
LABEL org.opencontainers.image.licenses=MIT
LABEL io.jumpcloud.terraform.platforms="linux_amd64,linux_arm64,darwin_amd64,darwin_arm64,windows_amd64"
LABEL io.jumpcloud.terraform.version="${VERSION}"
LABEL io.jumpcloud.terraform.targetplatform="\${TARGETPLATFORM}"
LABEL io.jumpcloud.terraform.registry="registry.terraform.io/agilize/jumpcloud"

WORKDIR /terraform-provider

# Copy platform-specific binaries and checksums
COPY dist/*.zip /terraform-provider/
COPY dist/SHA256SUMS /terraform-provider/

# Add usage information
COPY README.md /terraform-provider/
COPY LICENSE /terraform-provider/

# Create platform identifier file
RUN echo "\${TARGETPLATFORM}" > /terraform-provider/PLATFORM

# Default command
CMD ["sh", "-c", "echo 'Terraform Provider for JumpCloud - Platform: '\$(cat /terraform-provider/PLATFORM)"]
EOF

# Ensure Docker BuildX is available
echo "Setting up Docker BuildX..."
docker buildx inspect --bootstrap

# Build and publish Docker image
echo "Building and publishing Docker image for all platforms..."
echo "This may take several minutes..."
docker buildx build --platform linux/amd64,linux/arm64,darwin/amd64,darwin/arm64,windows/amd64 \
  -t ghcr.io/${GITHUB_USER:-agilize}/terraform-provider-jumpcloud:v${VERSION} \
  -t ghcr.io/${GITHUB_USER:-agilize}/terraform-provider-jumpcloud:latest \
  --build-arg VERSION=${VERSION} \
  --push \
  .

echo "Logging in to GitHub Container Registry..."
echo "Please provide your GitHub personal access token with packages:write permission"
echo "Or press Enter to skip this step if you're already authenticated."
read -s GITHUB_TOKEN

if [ -n "$GITHUB_TOKEN" ]; then
  echo $GITHUB_TOKEN | docker login ghcr.io -u ${GITHUB_USER:-agilize} --password-stdin
fi

echo "Provider published successfully!"
echo ""
echo "=== Uso com GitHub Container Registry (Método Alternativo) ==="
echo "To use it, add to your ~/.terraformrc:"
echo "
provider_installation {
  network_mirror {
    url = \"https://ghcr.io/${GITHUB_USER:-agilize}/terraform-provider-jumpcloud\"
    include = [\"ghcr.io/${GITHUB_USER:-agilize}/jumpcloud\"]
  }
  direct {
    exclude = [\"ghcr.io/${GITHUB_USER:-agilize}/jumpcloud\"]
  }
}
"

echo ""
echo "Then in your Terraform file:"
echo "
terraform {
  required_providers {
    jumpcloud = {
      source  = \"ghcr.io/${GITHUB_USER:-agilize}/jumpcloud\"
      version = \"~> ${VERSION}\"
    }
  }
}
"

echo ""
echo "=== Uso com Terraform Registry (Método Recomendado) ==="
echo "Assim que o provider estiver publicado no Terraform Registry, você poderá usá-lo com:"
echo "
terraform {
  required_providers {
    jumpcloud = {
      source  = \"registry.terraform.io/agilize/jumpcloud\"
      version = \"~> ${VERSION}\"
    }
  }
}
"

echo ""
echo "Para publicar no Terraform Registry, visite: https://registry.terraform.io/publish/provider"
echo ""
echo "Lembre-se de adicionar sua chave GPG pública ao Terraform Registry antes de publicar."

echo ""
echo "This provider includes binaries for ALL supported platforms:"
echo "- Linux AMD64 & ARM64"
echo "- MacOS AMD64 & ARM64"
echo "- Windows AMD64"
echo ""
echo "The GitHub Container Registry should now show support for all 5 platforms." 