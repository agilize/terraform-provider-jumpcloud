# Using the JumpCloud Provider via GitHub Container Registry (GHCR)

This document explains how to configure and use the JumpCloud Terraform Provider hosted on GitHub Container Registry.

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

The provider container includes:

1. Provider binaries for all supported platforms (Linux, macOS, Windows)
2. Checksum file (SHA256SUMS)
3. Documentation and licenses

## Troubleshooting

If you encounter issues using the provider via GHCR:

1. Verify that your ~/.terraformrc configuration is correct
2. Run `terraform init -upgrade` to force downloading the latest version
3. Check Terraform logs with `TF_LOG=debug terraform init`
4. Make sure you have permission to access the container (public packages should not require authentication)

## Local Development

For developers contributing to the provider, you can publish a local version using:

```bash
./scripts/publish-local.sh
```

This script compiles the provider, creates a Docker image, and publishes it to GHCR under your username. 