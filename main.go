package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud"
)

// main is the entry point for the provider.
// It initializes and serves the JumpCloud Terraform provider.
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: jumpcloud.New,
	})
}
