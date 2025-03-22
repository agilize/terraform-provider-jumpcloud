package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: jumpcloud.New,
	})
}
