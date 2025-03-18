FROM alpine:3.17

LABEL org.opencontainers.image.source=https://github.com/ferreirafa/terraform-provider-jumpcloud
LABEL org.opencontainers.image.description="JumpCloud Terraform Provider"
LABEL org.opencontainers.image.licenses=MIT

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