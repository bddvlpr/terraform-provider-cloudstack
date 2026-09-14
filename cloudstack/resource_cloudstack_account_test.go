//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//

package cloudstack

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCloudStackAccount_basic(t *testing.T) {
	var account cloudstack.Account
	name := fmt.Sprintf("terraform-account-basic-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackAccountConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackAccountExists("cloudstack_account.foo", &account),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "username", name),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "generate_keys", "false"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "user_id"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "api_key", ""),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "secret_key", ""),
				),
			},
			{
				ResourceName:            "cloudstack_account.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"generate_keys", "password"},
			},
		},
	})
}

func TestAccCloudStackAccount_generateKeys(t *testing.T) {
	var account cloudstack.Account
	name := fmt.Sprintf("terraform-account-keys-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackAccountConfig(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackAccountExists("cloudstack_account.foo", &account),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "username", name),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "generate_keys", "true"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "user_id"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "api_key"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "secret_key"),
				),
			},
			{
				Config: testAccCloudStackAccountConfig(name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackAccountExists("cloudstack_account.foo", &account),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "user_id"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "api_key"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "secret_key"),
				),
			},
		},
	})
}

func TestAccCloudStackAccount_updateUser(t *testing.T) {
	var account cloudstack.Account
	name := fmt.Sprintf("terraform-account-update-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackAccountConfig(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackAccountExists("cloudstack_account.foo", &account),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "email", name+"@example.com"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "first_name", "Terraform"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "last_name", "Account"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "username", name),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "user_id"),
				),
			},
			{
				Config: testAccCloudStackAccountConfigUpdatedUser(name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackAccountExists("cloudstack_account.foo", &account),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "email", name+"-updated@example.com"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "first_name", "Updated"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "last_name", "User"),
					resource.TestCheckResourceAttr("cloudstack_account.foo", "username", name+"-updated"),
					resource.TestCheckResourceAttrSet("cloudstack_account.foo", "user_id"),
				),
			},
		},
	})
}

func testAccCheckCloudStackAccountExists(n string, account *cloudstack.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Account ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		a, _, err := cs.Account.GetAccountByID(rs.Primary.ID)
		if err != nil {
			return err
		}
		if a.Id != rs.Primary.ID {
			return fmt.Errorf("Account not found")
		}

		*account = *a

		return nil
	}
}

func testAccCheckCloudStackAccountDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_account" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Account ID is set")
		}

		_, count, err := cs.Account.GetAccountByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Account %s still exists", rs.Primary.ID)
		}
		if count != 0 && !strings.Contains(err.Error(), "Unable to find account") {
			return err
		}
	}

	return nil
}

func testAccCloudStackAccountConfig(name string, generateKeys bool) string {
	return fmt.Sprintf(`
resource "cloudstack_role" "account_role" {
  name        = "%[1]s-role"
  description = "Terraform acceptance test account role"
  is_public   = true
  type        = "User"
}

resource "cloudstack_account" "foo" {
  username      = "%[1]s"
  password      = "password"
  first_name    = "Terraform"
  last_name     = "Account"
  email         = "%[1]s@example.com"
  account_type  = 0
  role_id       = cloudstack_role.account_role.id
  generate_keys = %[2]t
}
`, name, generateKeys)
}

func testAccCloudStackAccountConfigUpdatedUser(name string, generateKeys bool) string {
	return fmt.Sprintf(`
resource "cloudstack_role" "account_role" {
  name        = "%[1]s-role"
  description = "Terraform acceptance test account role"
  is_public   = true
  type        = "User"
}

resource "cloudstack_account" "foo" {
  account       = "%[1]s"
  username      = "%[1]s-updated"
  password      = "updated-password"
  first_name    = "Updated"
  last_name     = "User"
  email         = "%[1]s-updated@example.com"
  account_type  = 0
  role_id       = cloudstack_role.account_role.id
  generate_keys = %[2]t
}
`, name, generateKeys)
}
