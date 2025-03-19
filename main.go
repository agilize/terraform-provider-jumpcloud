package main

import (
	"github.com/agilize/terraform-provider-jumpcloud/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// main is the entry point for the provider
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New,
	})
}
