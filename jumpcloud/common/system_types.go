package common

// System represents a system in JumpCloud
type System struct {
	ID                             string                 `json:"_id,omitempty"`
	DisplayName                    string                 `json:"displayName"`
	HostName                       string                 `json:"hostname,omitempty"`
	OS                             string                 `json:"os,omitempty"`
	SystemType                     string                 `json:"systemType,omitempty"` // linux, windows, mac, etc.
	Version                        string                 `json:"version,omitempty"`
	Architecture                   string                 `json:"arch,omitempty"`
	RemoteIP                       string                 `json:"remoteIP,omitempty"`
	LastContact                    string                 `json:"lastContact,omitempty"`
	AgentVersion                   string                 `json:"agentVersion,omitempty"`
	AllowMFA                       bool                   `json:"allowMFA,omitempty"`
	AllowSshPassAuth               bool                   `json:"allowSshPasswordAuthentication,omitempty"`
	AllowSshRootLogin              bool                   `json:"allowSshRootLogin,omitempty"`
	AllowSshPasswordAuthentication bool                   `json:"allowSshPasswordAuthentication,omitempty"`
	AllowMultiFactorAuthentication bool                   `json:"allowMultiFactorAuthentication,omitempty"`
	AllowAutoUpdate                bool                   `json:"allowAutoUpdate,omitempty"`
	SystemTimeZone                 string                 `json:"systemTimeZone,omitempty"`
	SystemInsights                 bool                   `json:"systemInsights,omitempty"`
	Created                        string                 `json:"created,omitempty"`
	LastUpdated                    string                 `json:"lastUpdated,omitempty"`
	Organization                   string                 `json:"organization,omitempty"`
	NetworkInterfaces              []NetworkInterface     `json:"networkInterfaces,omitempty"`
	Tags                           []string               `json:"tags,omitempty"`
	Attributes                     map[string]interface{} `json:"attributes,omitempty"`
	SshRootEnabled                 bool                   `json:"sshRootEnabled,omitempty"`
	AutoDeploymentEnabled          bool                   `json:"autoDeploymentEnabled,omitempty"`
	Description                    string                 `json:"description,omitempty"`
}

// NetworkInterface represents a network interface of a system
type NetworkInterface struct {
	Name     string `json:"name,omitempty"`
	Family   string `json:"family,omitempty"`
	MAC      string `json:"address,omitempty"`
	IP       string `json:"ipAddress,omitempty"`
	Internal bool   `json:"internal,omitempty"`
}
