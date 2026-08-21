---
layout: "cloudstack"
page_title: "CloudStack: cloudstack_security_group"
description: |-
  Gets information about a security group.
---

# cloudstack_security_group

Use this data source to get information about a security group for use in other resources.

## Example Usage

```hcl
data "cloudstack_security_group" "web" {
  project = "my-project"

  filter {
    name  = "name"
    value = "^web-servers$"
  }
}
```

## Argument Reference

The following arguments are supported:

* `filter` - (Required) One or more name/regular-expression pairs used to select the security group. Filter names use snake case, for example `name`, `description`, or `project`.
* `project` - (Optional) The name or ID of the project containing the security group.

The filters must identify exactly one security group.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the security group.
* `name` - The name of the security group.
* `description` - The description of the security group.
* `project` - The name of the project containing the security group, or its ID when the data source was configured with a project ID.
