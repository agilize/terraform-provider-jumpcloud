package common

import (
	"fmt"
)

// API version constants
const (
	// APIVersionV1 represents JumpCloud API v1
	APIVersionV1 = "v1"
	// APIVersionV2 represents JumpCloud API v2
	APIVersionV2 = "v2"
)

// Base API paths
const (
	// BaseAPIPath is the base path for all API requests
	BaseAPIPath = "/api"
	// V1APIPath is the base path for v1 API requests
	V1APIPath = BaseAPIPath + "/" + APIVersionV1
	// V2APIPath is the base path for v2 API requests
	V2APIPath = BaseAPIPath + "/" + APIVersionV2
)

// Resource-specific API paths
const (
	// SystemUsersPath is the path for system users operations (v1)
	SystemUsersPath = V1APIPath + "/systemusers"
	// UserGroupsPath is the path for user groups operations (v2)
	UserGroupsPath = V2APIPath + "/usergroups"
	// UserGroupMembershipsPath is the path for user group memberships operations (v2)
	UserGroupMembershipsPath = V2APIPath + "/usergroups/%s/members"
)

// GetSystemUserPath returns the path for a specific system user
func GetSystemUserPath(userID string) string {
	return SystemUsersPath + "/" + userID
}

// GetUserGroupPath returns the path for a specific user group
func GetUserGroupPath(groupID string) string {
	return UserGroupsPath + "/" + groupID
}

// GetUserGroupMembershipsPath returns the path for a specific user group's memberships
func GetUserGroupMembershipsPath(groupID string) string {
	return fmt.Sprintf(UserGroupMembershipsPath, groupID)
}
