package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"registry.terraform.io/agilize/jumpcloud/internal/provider"
)

// main is the entry point for the provider
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New,
	})
}
