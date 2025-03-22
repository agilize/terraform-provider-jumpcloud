package admin

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceRoleBinding returns a placeholder schema.Resource for JumpCloud admin role bindings
// This is a placeholder and will be implemented in the future
func ResourceRoleBinding() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
