# Using the JumpCloud Provider via GitHub Container Registry (GHCR)

This document explains how to configure and use the JumpCloud Terraform Provider hosted on GitHub Container Registry.

## Supported Platforms

The provider is available for all the following platforms:

- **Linux**: AMD64, ARM64
- **macOS**: AMD64, ARM64
- **Windows**: AMD64

Each platform has a dedicated container image variant, and the GitHub Container Registry automatically serves the correct variant based on your system. You'll see all 5 platform combinations listed under the "OS / Arch" section in the GitHub Container Registry interface.

## Terraform Configuration

### 1. Configure the ~/.terraformrc file

For Terraform to fetch the provider from the GitHub Container Registry, you need to add a `provider_installation` configuration to your Terraform configuration file. This file is located at:

- Linux/Mac: `~/.terraformrc`
- Windows: `%APPDATA%\terraform.rc`

Add the following content:

```hcl
provider_installation {
  network_mirror {
    url = "https://ghcr.io/ferreirafa/terraform-provider-jumpcloud"
    include = ["ghcr.io/ferreirafa/jumpcloud"]
  }
  direct {
    exclude = ["ghcr.io/ferreirafa/jumpcloud"]
  }
}
```

### 2. Configure your Terraform file

In your Terraform configuration file (usually `main.tf`), declare the provider as follows:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "ghcr.io/ferreirafa/jumpcloud"
      version = "0.1.0" # Replace with the desired version
    }
  }
}

provider "jumpcloud" {
  api_key = "your-api-key" # We recommend using environment variables
}
```

## Checking Available Versions

You can check the available versions of the provider by visiting:
https://github.com/ferreirafa/terraform-provider-jumpcloud/pkgs/container/terraform-provider-jumpcloud

## Available Tags

The provider is published with the following tags:

- `vX.Y.Z` - Specific versions (e.g., `v0.1.0`)
- `latest` - Always points to the most recent stable version
- `beta` - Latest beta version (from the `develop` branch)

## Using Beta Versions

To use a beta version of the provider, update the configuration:

```hcl
terraform {
  required_providers {
    jumpcloud = {
      source  = "ghcr.io/ferreirafa/jumpcloud"
      version = "0.1.0-beta" # Specific beta version
    }
  }
}
```

## JumpCloud Authentication

The JumpCloud provider requires an API key for authentication. You can provide it in two ways:

1. **Environment variable (recommended):**
   ```bash
   export JUMPCLOUD_API_KEY="your-api-key-here"
   ```

2. **Direct configuration in the provider:**
   ```hcl
   provider "jumpcloud" {
     api_key = "your-api-key-here"
   }
   ```

## Container Structure

Each platform-specific container includes:

1. Provider binaries for all supported platforms
2. Checksum file (SHA256SUMS)
3. Documentation and licenses
4. Platform identifier file

## Viewing Architecture Checksums

To view the SHA256 checksums for all supported architectures, you can run:

```bash
./scripts/list-architectures.sh [VERSION]
```

This script will display a table of all supported architectures and their corresponding SHA256 checksums.

## Troubleshooting

If you encounter issues using the provider via GHCR:

1. Verify that your ~/.terraformrc configuration is correct
2. Run `terraform init -upgrade` to force downloading the latest version
3. Check Terraform logs with `TF_LOG=debug terraform init`
4. Make sure you have permission to access the container (public packages should not require authentication)
5. If using macOS or Windows, ensure your Docker installation supports multi-platform images

## Local Development

For developers contributing to the provider, you can publish a local version using:

```bash
./scripts/publish-local.sh
```

This script compiles the provider, creates Docker images for all supported platforms, and publishes them to GHCR under your username. 