FROM alpine:3.17

LABEL org.opencontainers.image.source=https://github.com/agilize/terraform-provider-jumpcloud
LABEL org.opencontainers.image.description="JumpCloud Terraform Provider - Contains binaries for Linux, macOS, and Windows (AMD64/ARM64)"
LABEL org.opencontainers.image.licenses=MIT
LABEL io.jumpcloud.terraform.platforms="linux_amd64,linux_arm64,darwin_amd64,darwin_arm64,windows_amd64"

ARG VERSION
WORKDIR /terraform-provider

# Copy binaries and checksums
COPY dist/*.zip /terraform-provider/
COPY dist/SHA256SUMS /terraform-provider/

# Add usage information
COPY README.md /terraform-provider/
COPY LICENSE /terraform-provider/

# Default command
CMD ["sh", "-c", "echo 'Terraform Provider for JumpCloud - Use with Terraform CLI'"] 