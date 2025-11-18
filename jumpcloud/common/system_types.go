package common

import (
	"encoding/json"
)

// System represents a system in JumpCloud
// Note: Systems API uses v1 which returns both "_id" and "id" fields
type System struct {
	ID                             string             `json:"_id,omitempty"`
	DisplayName                    string             `json:"displayName"`
	HostName                       string             `json:"hostname,omitempty"`
	OS                             string             `json:"os,omitempty"`
	SystemType                     string             `json:"systemType,omitempty"` // linux, windows, mac, etc.
	Version                        string             `json:"version,omitempty"`
	Architecture                   string             `json:"arch,omitempty"`
	RemoteIP                       string             `json:"remoteIP,omitempty"`
	LastContact                    string             `json:"lastContact,omitempty"`
	AgentVersion                   string             `json:"agentVersion,omitempty"`
	AllowMFA                       bool               `json:"allowMFA,omitempty"`
	AllowSshPassAuth               bool               `json:"allowSshPassAuth,omitempty"`
	AllowSshRootLogin              bool               `json:"allowSshRootLogin,omitempty"`
	AllowSshPasswordAuthentication bool               `json:"allowSshPasswordAuthentication,omitempty"`
	AllowMultiFactorAuthentication bool               `json:"allowMultiFactorAuthentication,omitempty"`
	AllowAutoUpdate                bool               `json:"allowAutoUpdate,omitempty"`
	SystemTimeZone                 int                `json:"systemTimezone,omitempty"` // Timezone offset in minutes
	SystemInsights                 SystemInsightsInfo `json:"systemInsights,omitempty"`
	Created                        string             `json:"created,omitempty"`
	LastUpdated                    string             `json:"lastUpdated,omitempty"`
	Organization                   string             `json:"organization,omitempty"`
	NetworkInterfaces              []NetworkInterface `json:"networkInterfaces,omitempty"`
	Tags                           []string           `json:"tags,omitempty"`
	Attributes                     SystemAttributes   `json:"attributes,omitempty"`
	SshRootEnabled                 bool               `json:"sshRootEnabled,omitempty"`
	AutoDeploymentEnabled          bool               `json:"autoDeploymentEnabled,omitempty"`
	Description                    string             `json:"description,omitempty"`
}

// SystemAttributes is a custom type that can handle both array and map formats
// The API returns [] when empty and a map when there are attributes
type SystemAttributes map[string]interface{}

// UnmarshalJSON implements custom unmarshaling to handle both array and map
func (sa *SystemAttributes) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as array first (empty attributes case)
	var arr []interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		// It's an array, initialize as empty map
		*sa = make(map[string]interface{})
		return nil
	}

	// Try to unmarshal as map
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*sa = m
	return nil
}

// SystemInsightsInfo represents system insights information
type SystemInsightsInfo struct {
	State string `json:"state,omitempty"` // "enabled" or "disabled"
}

// NetworkInterface represents a network interface of a system
type NetworkInterface struct {
	Name     string `json:"name,omitempty"`
	Family   string `json:"family,omitempty"`
	MAC      string `json:"address,omitempty"`
	IP       string `json:"ipAddress,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}
