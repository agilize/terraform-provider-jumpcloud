package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
)

// JumpCloudClient is an interface for interaction with the JumpCloud API
type JumpCloudClient interface {
	DoRequest(method, path string, body interface{}) ([]byte, error)
	GetApiKey() string
	GetOrgID() string
}

// GetJumpCloudClient extracts the JumpCloud client from the meta interface
func GetJumpCloudClient(meta interface{}) (JumpCloudClient, diag.Diagnostics) {
	client, ok := meta.(*apiclient.Client)
	if !ok {
		return nil, diag.Errorf("error asserting meta as *apiclient.Client")
	}
	return client, nil
}
