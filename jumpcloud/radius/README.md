# RADIUS

This directory contains resources for managing RADIUS (Remote Authentication Dial-In User Service) configurations in JumpCloud. RADIUS is a protocol for providing centralized authentication, authorization, and accounting management for users connecting to network services.

## Resources

### jumpcloud_radius_server

The `jumpcloud_radius_server` resource allows you to create and manage RADIUS server configurations in JumpCloud. RADIUS servers enable network devices to authenticate users against JumpCloud's directory.

#### Example Usage

```hcl
resource "jumpcloud_radius_server" "example" {
  name                            = "Corporate VPN"
  shared_secret                   = "your-secure-shared-secret"
  network_source_ip               = "203.0.113.10"
  mfa_required                    = true
  user_password_expiration_action = "deny"
  user_lockout_action             = "deny"
  user_attribute                  = "username"
  
  # Optional: Associate with user groups
  targets = [
    jumpcloud_user_group.vpn_users.id
  ]
}

# User group for VPN users
resource "jumpcloud_user_group" "vpn_users" {
  name = "VPN Users"
}
```

## RADIUS Authentication Flow

When configured:

1. A user attempts to authenticate to a network device (e.g., VPN, wireless access point)
2. The network device forwards the authentication request to the JumpCloud RADIUS server
3. JumpCloud validates the credentials against its directory
4. If configured, MFA can be enforced for additional security
5. JumpCloud returns an accept/reject decision to the network device

## Security Considerations

- Always use a strong shared secret
- Consider enabling MFA for enhanced security
- Restrict network access to your RADIUS server
- Use specific user groups as targets rather than allowing all users 