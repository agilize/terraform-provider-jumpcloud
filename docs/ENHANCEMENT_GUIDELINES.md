# JumpCloud Provider Enhancement Guide

This guide outlines the systematic approach to fix issues in the JumpCloud provider and add new functionality. Follow these steps to apply similar improvements to other resources.

## 1. Resource Improvements

### 1.1 Attribute Name Handling

**Problem:** Special characters in attribute names were causing issues with the JumpCloud API.

**Solution:**
1. Sanitize attribute names when sending to the API:
   ```go
   // Sanitize attribute name for API
   sanitizedName := sanitizeAttributeName(name)
   
   attributes = append(attributes, UserAttribute{
       Name:  sanitizedName, // Use sanitized name for API
       Value: strValue,
   })
   ```

2. Preserve original attribute names in state:
   ```go
   // In the Read function
   if len(user.Attributes) > 0 {
       // Get the original attributes from the configuration
       oldAttrs := d.Get("attributes").(map[string]interface{})
       
       // Create a map of sanitized name -> original name
       sanitizedToOriginal := make(map[string]string)
       for origName := range oldAttrs {
           sanitizedToOriginal[sanitizeAttributeName(origName)] = origName
       }
       
       // Create new attribute map preserving original names
       attrMap := make(map[string]any)
       for _, attr := range user.Attributes {
           if origName, exists := sanitizedToOriginal[attr.Name]; exists {
               attrMap[origName] = attr.Value
           } else {
               attrMap[attr.Name] = attr.Value
           }
       }
       d.Set("attributes", attrMap)
   }
   ```

### 1.2 Phone Number Formatting

**Problem:** Phone numbers with hyphens were being stripped, causing update loops.

**Solution:**
1. Preserve original format in state:
   ```go
   // In the Read function
   if len(user.PhoneNumbers) > 0 {
       // Get the original phone numbers from the configuration
       oldPhones := d.Get("phone_numbers").([]interface{})
       oldPhoneMap := make(map[string]string)
       
       // Create a map of type -> number from the old configuration
       for _, oldPhone := range oldPhones {
           oldPhoneData := oldPhone.(map[string]interface{})
           oldPhoneMap[oldPhoneData["type"].(string)] = oldPhoneData["number"].(string)
       }
       
       // Create new phone list preserving original formatting
       phones := make([]map[string]interface{}, 0, len(user.PhoneNumbers))
       for _, phone := range user.PhoneNumbers {
           originalNumber, exists := oldPhoneMap[phone.Type]
           
           phoneMap := map[string]interface{}{
               "type": phone.Type,
           }
           
           // Use original formatted number if digits match
           if exists && sanitizePhoneNumber(originalNumber) == sanitizePhoneNumber(phone.Number) {
               phoneMap["number"] = originalNumber
           } else {
               phoneMap["number"] = phone.Number
           }
           
           phones = append(phones, phoneMap)
       }
       d.Set("phone_numbers", phones)
   }
   ```

2. Use original format in Create/Update but sanitize for API:
   ```go
   // When sending to API, sanitize but keep original in state
   sanitizedNumber := sanitizePhoneNumber(phoneMap["number"].(string))
   // Use sanitized version for API calls
   ```

### 1.3 Boolean Fields

**Problem:** Boolean fields like `mfa_enabled` were not being properly persisted in state.

**Solution:**
Use configuration values for boolean fields in the Read function:
```go
// For critical boolean fields that are causing loops, use the configuration value
configMfaEnabled := d.Get("mfa_enabled").(bool)
d.Set("mfa_enabled", configMfaEnabled)

// For other boolean fields
configEnableManagedUID := d.Get("enable_managed_uid").(bool)
d.Set("enable_managed_uid", configEnableManagedUID)
```

### 1.4 Special String Fields

**Problem:** Fields like `delegated_authority` were not being properly persisted.

**Solution:**
Preserve configuration values for these fields:
```go
// For delegated_authority
if v, ok := d.GetOk("delegated_authority"); ok {
    d.Set("delegated_authority", v.(string))
} else {
    d.Set("delegated_authority", user.DelegatedAuthority)
}
```

### 1.5 Computed Fields

**Problem:** Computed fields like `security_keys` were showing as "(known after apply)".

**Solution:**
Set empty values for computed fields when they're not present:
```go
if len(user.SecurityKeys) > 0 {
    // Set the keys
    d.Set("security_keys", keys)
} else {
    // Set an empty list to prevent "(known after apply)" in plans
    d.Set("security_keys", []map[string]interface{}{})
}
```

## 2. Adding Data Sources

### 2.1 Data Source Structure

1. Create a new file `data_source_resource.go` with this structure:
   ```go
   package resource_package
   
   import (
       "context"
       "fmt"
       "net/http"
       
       "github.com/hashicorp/terraform-plugin-log/tflog"
       "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
       "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
       "registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
   )
   
   // DataSourceResource returns a schema for the JumpCloud resource data source
   func DataSourceResource() *schema.Resource {
       return &schema.Resource{
           ReadContext: dataSourceResourceRead,
           Schema: map[string]*schema.Schema{
               // Define lookup parameters
               "resource_id": {
                   Type:          schema.TypeString,
                   Optional:      true,
                   ConflictsWith: []string{"other_lookup_param"},
                   Description:   "The ID of the resource to retrieve",
               },
               // Define other lookup parameters
               
               // Define all output fields (same as resource)
               "field1": {
                   Type:        schema.TypeString,
                   Computed:    true,
                   Description: "Description of field1",
               },
               // Add all other fields from the resource
           },
       }
   }
   
   func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
       var diags diag.Diagnostics
       
       // Get client
       c, diagErr := common.GetClientFromMeta(meta)
       if diagErr != nil {
           return diagErr
       }
       
       // Determine lookup method
       var path string
       if resourceID, ok := d.GetOk("resource_id"); ok {
           path = fmt.Sprintf("/endpoint/%s", resourceID.(string))
       } else if otherParam, ok := d.GetOk("other_lookup_param"); ok {
           // For other lookup methods, get all resources and filter
           path = "/endpoint"
       } else {
           return diag.FromErr(fmt.Errorf("one of resource_id or other_lookup_param must be provided"))
       }
       
       // Make API request
       resp, err := c.DoRequest(http.MethodGet, path, nil)
       if err != nil {
           return diag.FromErr(fmt.Errorf("error reading resource: %v", err))
       }
       
       // Parse response
       var resource ResourceType
       if err := json.Unmarshal(resp, &resource); err != nil {
           return diag.FromErr(fmt.Errorf("error parsing response: %v", err))
       }
       
       // Set ID and fields
       d.SetId(resource.ID)
       if err := setResourceFields(d, &resource); err != nil {
           return diag.FromErr(err)
       }
       
       return diags
   }
   
   // Helper to set all fields
   func setResourceFields(d *schema.ResourceData, resource *ResourceType) error {
       // Set all fields from the resource
       if err := d.Set("field1", resource.Field1); err != nil {
           return fmt.Errorf("error setting field1: %v", err)
       }
       // Set all other fields
       
       return nil
   }
   ```

2. Register the data source in `provider.go`:
   ```go
   DataSourcesMap: map[string]*schema.Resource{
       // Existing data sources
       "jumpcloud_resource": resource_package.DataSourceResource(),
   },
   ```

### 2.2 Data Source Tests

Create a test file `data_source_resource_test.go` in the _test package:

```go
package resource_package_test

import (
    "os"
    "testing"
    
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "registry.terraform.io/agilize/jumpcloud/jumpcloud"
)

// Test helpers
func testAccPreCheck(t *testing.T) {
    if v := os.Getenv("JUMPCLOUD_API_KEY"); v == "" {
        t.Skip("JUMPCLOUD_API_KEY must be set for acceptance tests")
    }
}

var testAccProviders map[string]*schema.Provider

func init() {
    testAccProvider := jumpcloud.Provider()
    testAccProviders = map[string]*schema.Provider{
        "jumpcloud": testAccProvider,
    }
}

// Basic test
func TestAccDataSourceResource_basic(t *testing.T) {
    if os.Getenv("TF_ACC") == "" {
        t.Skip("Acceptance tests skipped unless TF_ACC=1 is set")
    }
    
    resource.Test(t, resource.TestCase{
        PreCheck:  func() { testAccPreCheck(t) },
        Providers: testAccProviders,
        Steps: []resource.TestStep{
            {
                Config: testAccDataSourceResourceConfig_basic(),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttrSet("data.jumpcloud_resource.test", "id"),
                    resource.TestCheckResourceAttr("data.jumpcloud_resource.test", "field1", "expected_value"),
                ),
            },
        },
    })
}

func testAccDataSourceResourceConfig_basic() string {
    return `
resource "jumpcloud_resource" "test" {
  field1 = "value1"
  field2 = "value2"
}

data "jumpcloud_resource" "test" {
  resource_id = jumpcloud_resource.test.id
  depends_on = [jumpcloud_resource.test]
}
`
}

// Add more tests for different lookup methods
```

### 2.3 Documentation

Create documentation in `docs/data-sources/resource.md`:

```markdown
---
page_title: "JumpCloud: jumpcloud_resource"
subcategory: "Resources"
description: |-
  Provides information about a JumpCloud resource.
---

# jumpcloud_resource

This data source provides information about a JumpCloud resource.

## Example Usage

### Lookup by ID

```hcl
data "jumpcloud_resource" "example" {
  resource_id = "5f1b2c3d4e5f6g7h8i9j0k"
}

output "resource_field1" {
  value = data.jumpcloud_resource.example.field1
}
```

### Lookup by Other Parameter

```hcl
data "jumpcloud_resource" "example" {
  other_lookup_param = "value"
}
```

## Argument Reference

The following arguments are supported:

* `resource_id` - (Optional) The ID of the resource to retrieve. Conflicts with `other_lookup_param`.
* `other_lookup_param` - (Optional) Another way to look up the resource. Conflicts with `resource_id`.

-> **Note:** Exactly one of these arguments must be provided.

## Attributes Reference

In addition to the arguments listed above, the following attributes are exported:

* `id` - The ID of the resource.
* `field1` - Description of field1.
* `field2` - Description of field2.
* ... (list all attributes)
```

## 3. Testing and Validation

### 3.1 Unit Tests

Run unit tests for your changes:
```bash
go test ./...
```

### 3.2 Linting and Vet

Run linting and vet checks:
```bash
go vet ./...
golangci-lint run
```

### 3.3 Acceptance Tests

Run acceptance tests (requires API credentials):
```bash
TF_ACC=1 go test ./path/to/package/... -v
```

### 3.4 Manual Testing

1. Build the provider:
   ```bash
   go build -o dist/terraform-provider-jumpcloud
   ```

2. Create a test configuration:
   ```hcl
   provider "jumpcloud" {
     api_key = "your_api_key"
   }
   
   resource "jumpcloud_resource" "test" {
     field1 = "value1"
     field2 = "value2"
   }
   
   data "jumpcloud_resource" "test_data" {
     resource_id = jumpcloud_resource.test.id
     depends_on = [jumpcloud_resource.test]
   }
   
   output "test_output" {
     value = data.jumpcloud_resource.test_data.field1
   }
   ```

3. Test with Terraform:
   ```bash
   terraform init
   terraform apply
   ```

## 4. Common Pitfalls to Avoid

1. **Import Cycles**: Use _test packages for tests to avoid import cycles.

2. **Unnecessary fmt.Sprintf**: Use string literals instead of fmt.Sprintf when not formatting values:
   ```go
   // Instead of
   return fmt.Sprintf(`
   resource "jumpcloud_resource" "test" {
     field1 = "value1"
   }
   `)
   
   // Use
   return `
   resource "jumpcloud_resource" "test" {
     field1 = "value1"
   }
   `
   ```

3. **Boolean Field Loops**: Always preserve configuration values for boolean fields in the Read function.

4. **Special Characters**: Sanitize attribute names and other fields that might contain special characters when sending to the API.

5. **Computed Fields**: Always set empty values for computed fields to prevent "(known after apply)" messages.

By following this guide, you can systematically improve other resources in the JumpCloud provider, ensuring consistent behavior and fixing common issues.
