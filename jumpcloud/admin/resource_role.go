package admin

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ResourceRole returns a placeholder schema.Resource for JumpCloud admin roles
// This is a placeholder and will be implemented in the future
func ResourceRole() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}
