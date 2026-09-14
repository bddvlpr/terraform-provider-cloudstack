---
page_title: "CloudStack: cloudstack_project_membership"
sidebar_current: "docs-cloudstack-resource-project-membership"
description: |-
  Adds an account or user to a project.
---

# cloudstack_project_membership

Adds an account or user to a CloudStack project.

## Example Usage

```hcl
resource "cloudstack_project_membership" "account" {
  project_id = cloudstack_project.project.id
  account    = "team-account"
}

resource "cloudstack_project_membership" "user" {
  project_id = cloudstack_project.project.id
  username   = "jane"
  role_type  = "Admin"
}
```

## Argument Reference

* `project_id` - (Required, ForceNew) The ID of the project.
* `account` - (Optional, ForceNew) The account name to add to the project. Exactly one of `account` or `username` must be set.
* `username` - (Optional, ForceNew) The username to add to the project. Exactly one of `account` or `username` must be set.
* `user_id` - (Optional, Computed, ForceNew) The user ID. This is resolved from `username` when omitted and is used when removing user memberships.
* `email` - (Optional, ForceNew) The email address used for project invitations.
* `role_type` - (Optional, ForceNew) The project role type assigned to the member. Valid values are `Admin` and `Regular`. Defaults to `Regular`.
* `project_role_id` - (Optional, ForceNew) The project role ID assigned to the member.

## Attributes Reference

* `id` - The project membership ID.
