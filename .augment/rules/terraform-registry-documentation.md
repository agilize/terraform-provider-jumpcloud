# Terraform Registry Documentation Standards

## Overview

All Terraform provider documentation must follow Terraform Registry standards for proper organization and discoverability.

---

## File Structure

```
docs/
├── index.md                    # Provider overview and configuration
├── resources/                  # Resource documentation
│   └── <resource_name>.md
└── data-sources/              # Data source documentation
    └── <data_source_name>.md
```

---

## YAML Frontmatter (Required)

Every documentation file **must** include YAML frontmatter at the top:

### Resources

```yaml
---
page_title: "JumpCloud: jumpcloud_<resource_name>"
subcategory: "<Subcategory Name>"
description: |-
  <Brief description of what this resource manages>
---

# jumpcloud_<resource_name>

<Detailed description>
```

### Data Sources

```yaml
---
page_title: "JumpCloud: jumpcloud_<data_source_name>"
subcategory: "<Subcategory Name>"
description: |-
  <Brief description of what this data source retrieves>
---

# jumpcloud_<data_source_name> (Data Source)

<Detailed description>
```

---

## Subcategory Mapping

Use these standardized subcategories based on the codebase domain structure:

| Subcategory | Domain Directory | Use For |
|-------------|------------------|---------|
| **User Management** | `user_management/users/`, `user_management/user_groups/` | Users, user groups, user group memberships |
| **Device Management** | `device_management/devices/`, `device_management/device_groups/` | Devices, device groups, policies, device associations |
| **Software Management** | `device_management/software/` | Software update policies |
| **User Authentication** | `user_authentication/` | RADIUS servers, SCIM servers, password managers, SSO |
| **Security Management** | `security_management/` | MFA settings, password policies, IP allow lists |
| **Organization Settings** | `organization_settings/` | Organizations, API keys, webhooks, admin settings |
| **Application Management** | `application_management/` | Applications, app-user mappings, app-group mappings |
| **Mobile Device Management** | `device_management/mdm/` | MDM profiles, commands, enrollment |

---

## Documentation Content Standards

### 1. Language
- **American English only**
- Professional, clear language
- No slang or regional terms
- Consistent terminology

### 2. Structure

Each documentation file should include:

1. **YAML Frontmatter** (required)
2. **Title** - `# jumpcloud_<name>` or `# jumpcloud_<name> (Data Source)`
3. **Description** - Brief overview of the resource/data source
4. **Example Usage** - At least 2-3 practical examples
5. **Argument Reference** - All input parameters
6. **Attribute Reference** - All exported attributes

### 3. Examples

- Use realistic, practical examples
- Include comments explaining the purpose
- Show common use cases
- Use proper HCL formatting

```hcl
# Create a user with MFA enabled
resource "jumpcloud_user" "admin" {
  username  = "john.doe"
  email     = "john.doe@example.com"
  firstname = "John"
  lastname  = "Doe"
  
  enable_mfa = true
}
```

---

## Provider Index (docs/index.md)

The provider index should:

1. Include provider configuration examples
2. List authentication methods
3. Organize resources and data sources **by subcategory**
4. Use current provider version in examples
5. Include links to important guides

---

## Checklist for New Documentation

When creating new resource/data source documentation:

- [ ] YAML frontmatter added with correct subcategory
- [ ] Title follows naming convention
- [ ] Description is clear and concise
- [ ] At least 2-3 example usage scenarios
- [ ] All arguments documented with types and descriptions
- [ ] All attributes documented
- [ ] Examples use realistic values
- [ ] Content is in American English
- [ ] No Portuguese or other languages
- [ ] Sensitive fields marked appropriately
- [ ] Provider index updated with new resource/data source

---

## Common Mistakes to Avoid

❌ **Don't:**
- Mix languages (Portuguese/English)
- Forget YAML frontmatter
- Use inconsistent subcategories
- Leave out example usage
- Use outdated provider versions in examples

✅ **Do:**
- Use consistent subcategories from the mapping table
- Include practical, realistic examples
- Keep descriptions concise but complete
- Follow the established structure
- Test examples before publishing

---

**Last Updated:** 2025-11-19  
**Next Review:** When adding new resource types or domains

