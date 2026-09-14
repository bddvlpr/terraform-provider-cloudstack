---
layout: "cloudstack"
page_title: "CloudStack: cloudstack_network_permission"
sidebar_current: "docs-cloudstack-resource-network-permission"
description: |-
  Grants a project permission to use a network.
---

# cloudstack_network_permission

Grants a project permission to use a network. Each resource manages one
network and project assignment without changing permissions granted to other
projects.

CloudStack does not allow permissions to be added to project-owned networks,
domain-shared networks, or VPC tiers.

## Example Usage

```hcl
resource "cloudstack_project" "example" {
  name        = "example-project"
  displaytext = "Example Project"
}

resource "cloudstack_network" "l2" {
  name             = "example-l2-network"
  display_text     = "Example L2 Network"
  network_offering = "DefaultL2NetworkOffering"
  zone             = "zone-1"
}

resource "cloudstack_network_permission" "example" {
  network_id = cloudstack_network.l2.id
  project_id = cloudstack_project.example.id
}
```

## Argument Reference

The following arguments are supported:

* `network_id` - (Required) ID of the network to share with the project.
  Changing this forces a new resource to be created.
* `project_id` - (Required) ID of the project that is allowed to use the
  network. Changing this forces a new resource to be created.

## Attributes Reference

The following attributes are exported:

* `id` - The network and project IDs separated by a slash.
