---
layout: default
page_title: "CloudStack: cloudstack_account"
sidebar_current: "docs-cloudstack-resource-account"
description: |-
    Creates a Account
---

# CloudStack: cloudstack_account

A `cloudstack_account` resource manages an account within CloudStack.

## Example Usage

```hcl
resource "cloudstack_account" "example" {
    email = "user@example.com"
    first_name = "John"
    last_name = "Doe"
    password = "securepassword"
    username = "jdoe"
    account_type = 1 # 1 for admin, 2 for domain admin, 0 for regular user
    role_id = "1234abcd" # ID of the role associated with the account
    generate_keys = true
}
```

## Argument Reference

The following arguments are supported:

* `email` - (Required) The email address of the account owner.
* `first_name` - (Required) The first name of the account owner.
* `last_name` - (Required) The last name of the account owner.
* `password` - (Required) The password for the account.
* `username` - (Required) The username of the account.
* `account_type` - (Required) The account type. Possible values are `0` for regular user, `1` for admin, and `2` for domain admin.
* `role_id` - (Required) The ID of the role associated with the account.
* `account` - (Optional) The account name. If not specified, the username will be used as the account name.
* `domain_id` - (Optional) Creates the user under the specified domain.
* `generate_keys` - (Optional) Whether to generate API keys for the account's primary user. Defaults to `false`.

## Attributes Reference

The following attributes are exported:

* `id` - The ID of the account.
* `user_id` - The ID of the account's primary user.
* `api_key` - The generated or current API key for the account's primary user. This attribute is sensitive.
* `secret_key` - The generated or current secret key for the account's primary user. This attribute is sensitive.

## Import

Accounts can be imported; use `<ACCOUNTID>` as the import ID. For example:

```shell
$ terraform import cloudstack_account.example <ACCOUNTID>
```

Account imports populate the user attributes when the account has a single user.
