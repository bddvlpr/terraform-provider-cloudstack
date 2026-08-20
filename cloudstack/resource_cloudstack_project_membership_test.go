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
	"os"
	"testing"
	"time"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestCloudStackProjectMembershipSchema(t *testing.T) {
	r := resourceCloudStackProjectMembership()
	if r.Schema["account"].ExactlyOneOf[0] != "account" || r.Schema["account"].ExactlyOneOf[1] != "username" {
		t.Fatalf("account should be mutually exclusive with username")
	}
	if r.Schema["role_type"].Default != "Regular" {
		t.Fatalf("expected role_type default to be Regular")
	}
}

func TestCloudStackProjectMembershipID(t *testing.T) {
	got := projectMembershipID("project-id", "account", "account-name")
	want := "project-id/account/account-name"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAccCloudStackProjectMembership_account(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}
	cs := newTestClient(t)
	projectID, accountName, _, cleanup := testAccProjectMembershipFixture(t, cs, "account")
	t.Cleanup(cleanup)

	r := resourceCloudStackProjectMembership()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"project_id": projectID,
		"account":    accountName,
	})

	if err := resourceCloudStackProjectMembershipCreate(d, cs); err != nil {
		t.Fatalf("failed to create account membership: %s", err)
	}
	if d.Id() == "" {
		t.Fatal("expected project membership ID to be set")
	}
	testAccAssertProjectMembership(t, cs, projectID, accountName, "")

	if err := resourceCloudStackProjectMembershipDelete(d, cs); err != nil {
		t.Fatalf("failed to delete account membership: %s", err)
	}
	testAccAssertNoProjectMembership(t, cs, projectID, accountName, "")
}

func TestAccCloudStackProjectMembership_user(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}
	cs := newTestClient(t)
	projectID, _, username, cleanup := testAccProjectMembershipFixture(t, cs, "user")
	t.Cleanup(cleanup)

	r := resourceCloudStackProjectMembership()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
		"project_id": projectID,
		"username":   username,
	})

	if err := resourceCloudStackProjectMembershipCreate(d, cs); err != nil {
		t.Fatalf("failed to create user membership: %s", err)
	}
	userID := d.Get("user_id").(string)
	if d.Id() == "" || userID == "" {
		t.Fatal("expected project membership ID and user_id to be set")
	}
	testAccAssertProjectMembership(t, cs, projectID, "", userID)

	if err := resourceCloudStackProjectMembershipDelete(d, cs); err != nil {
		t.Fatalf("failed to delete user membership: %s", err)
	}
	testAccAssertNoProjectMembership(t, cs, projectID, "", userID)
}

func TestAccCloudStackProjectMembershipTerraform_account(t *testing.T) {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackProjectMembershipTerraformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackProjectMembershipTerraformAccountConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackProjectMembershipTerraformExists("cloudstack_project_membership.account"),
					resource.TestCheckResourceAttr("cloudstack_project_membership.account", "role_type", "Regular"),
					resource.TestCheckResourceAttrSet("cloudstack_project_membership.account", "project_id"),
				),
			},
		},
	})
}

func TestAccCloudStackProjectMembershipTerraform_user(t *testing.T) {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackProjectMembershipTerraformDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackProjectMembershipTerraformUserConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackProjectMembershipTerraformExists("cloudstack_project_membership.user"),
					resource.TestCheckResourceAttrSet("cloudstack_project_membership.user", "project_id"),
					resource.TestCheckResourceAttrSet("cloudstack_project_membership.user", "user_id"),
				),
			},
		},
	})
}

func testAccProjectMembershipFixture(t *testing.T, cs *cloudstack.CloudStackClient, kind string) (string, string, string, func()) {
	t.Helper()
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	domainName := fmt.Sprintf("tf-membership-%s-%s", kind, suffix)
	ownerName := fmt.Sprintf("tf-owner-%s-%s", kind, suffix)
	accountName := fmt.Sprintf("tf-member-%s-%s", kind, suffix)
	projectName := fmt.Sprintf("tf-project-membership-%s-%s", kind, suffix)

	domain, err := cs.Domain.CreateDomain(cs.Domain.NewCreateDomainParams(domainName))
	if err != nil {
		t.Fatalf("failed to create test domain: %s", err)
	}

	owner, err := testAccCreateProjectMembershipAccount(cs, domain.Id, ownerName)
	if err != nil {
		t.Fatalf("failed to create test project owner account: %s", err)
	}

	account, err := testAccCreateProjectMembershipAccount(cs, domain.Id, accountName)
	if err != nil {
		t.Fatalf("failed to create test member account: %s", err)
	}

	projectParams := cs.Project.NewCreateProjectParams(projectName, projectName)
	projectParams.SetAccount(ownerName)
	projectParams.SetDomainid(domain.Id)
	project, err := cs.Project.CreateProject(projectParams)
	if err != nil {
		t.Fatalf("failed to create test project: %s", err)
	}

	cleanup := func() {
		_, _ = cs.Project.DeleteProject(cs.Project.NewDeleteProjectParams(project.Id))
		_, _ = cs.Account.DeleteAccount(cs.Account.NewDeleteAccountParams(account.Id))
		_, _ = cs.Account.DeleteAccount(cs.Account.NewDeleteAccountParams(owner.Id))
		p := cs.Domain.NewDeleteDomainParams(domain.Id)
		p.SetCleanup(true)
		_, _ = cs.Domain.DeleteDomain(p)
	}

	return project.Id, accountName, accountName, cleanup
}

func testAccCreateProjectMembershipAccount(cs *cloudstack.CloudStackClient, domainID, accountName string) (*cloudstack.CreateAccountResponse, error) {
	p := cs.Account.NewCreateAccountParams(
		fmt.Sprintf("%s@example.com", accountName),
		"Terraform",
		"Membership",
		"password",
		accountName,
	)
	p.SetAccount(accountName)
	p.SetAccounttype(0)
	p.SetRoleid("4")
	p.SetDomainid(domainID)
	return cs.Account.CreateAccount(p)
}

func testAccAssertProjectMembership(t *testing.T, cs *cloudstack.CloudStackClient, projectID, account, userID string) {
	t.Helper()
	exists, err := testAccProjectMembershipExists(cs, projectID, account, userID)
	if err != nil {
		t.Fatalf("failed to list project membership: %s", err)
	}
	if !exists {
		t.Fatalf("project membership was not found")
	}
}

func testAccAssertNoProjectMembership(t *testing.T, cs *cloudstack.CloudStackClient, projectID, account, userID string) {
	t.Helper()
	exists, err := testAccProjectMembershipExists(cs, projectID, account, userID)
	if err != nil {
		t.Fatalf("failed to list project membership: %s", err)
	}
	if exists {
		t.Fatalf("project membership still exists")
	}
}

func testAccProjectMembershipExists(cs *cloudstack.CloudStackClient, projectID, account, userID string) (bool, error) {
	p := cs.Account.NewListProjectAccountsParams(projectID)
	if account != "" {
		p.SetAccount(account)
	} else {
		p.SetUserid(userID)
	}

	l, err := cs.Account.ListProjectAccounts(p)
	if err != nil {
		return false, err
	}

	return l.Count != 0, nil
}

func testAccCheckCloudStackProjectMembershipTerraformExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No project membership ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		exists, err := testAccProjectMembershipExists(
			cs,
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["account"],
			rs.Primary.Attributes["user_id"],
		)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Project membership %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCloudStackProjectMembershipTerraformDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_project_membership" {
			continue
		}

		exists, err := testAccProjectMembershipExists(
			cs,
			rs.Primary.Attributes["project_id"],
			rs.Primary.Attributes["account"],
			rs.Primary.Attributes["user_id"],
		)
		if err != nil {
			continue
		}
		if exists {
			return fmt.Errorf("project membership %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCloudStackProjectMembershipTerraformAccountConfig(suffix string) string {
	return fmt.Sprintf(`
resource "cloudstack_domain" "membership_account" {
  name = "tf-tf-membership-account-%[1]s"
}

resource "cloudstack_role" "membership_account" {
  name = "tf-tf-membership-account-%[1]s"
  type = "User"
}

resource "cloudstack_account" "owner" {
  username     = "tf-tf-owner-account-%[1]s"
  password     = "password"
  first_name   = "Terraform"
  last_name    = "Owner"
  email        = "tf-tf-owner-account-%[1]s@example.com"
  account_type = 0
  role_id      = cloudstack_role.membership_account.id
  domain_id    = cloudstack_domain.membership_account.id
}

resource "cloudstack_account" "member" {
  username     = "tf-tf-member-account-%[1]s"
  password     = "password"
  first_name   = "Terraform"
  last_name    = "Member"
  email        = "tf-tf-member-account-%[1]s@example.com"
  account_type = 0
  role_id      = cloudstack_role.membership_account.id
  domain_id    = cloudstack_domain.membership_account.id
}

resource "cloudstack_project" "membership_account" {
  name        = "tf-tf-project-account-%[1]s"
  displaytext = "tf-tf-project-account-%[1]s"
  domain      = cloudstack_domain.membership_account.name
  account     = cloudstack_account.owner.username
}

resource "cloudstack_project_membership" "account" {
  project_id = cloudstack_project.membership_account.id
  account    = cloudstack_account.member.username
}
`, suffix)
}

func testAccCloudStackProjectMembershipTerraformUserConfig(suffix string) string {
	return fmt.Sprintf(`
resource "cloudstack_domain" "membership_user" {
  name = "tf-tf-membership-user-%[1]s"
}

resource "cloudstack_role" "membership_user" {
  name = "tf-tf-membership-user-%[1]s"
  type = "User"
}

resource "cloudstack_account" "owner" {
  username     = "tf-tf-owner-user-%[1]s"
  password     = "password"
  first_name   = "Terraform"
  last_name    = "Owner"
  email        = "tf-tf-owner-user-%[1]s@example.com"
  account_type = 0
  role_id      = cloudstack_role.membership_user.id
  domain_id    = cloudstack_domain.membership_user.id
}

resource "cloudstack_account" "member" {
  username     = "tf-tf-member-user-%[1]s"
  password     = "password"
  first_name   = "Terraform"
  last_name    = "Member"
  email        = "tf-tf-member-user-%[1]s@example.com"
  account_type = 0
  role_id      = cloudstack_role.membership_user.id
  domain_id    = cloudstack_domain.membership_user.id
}

resource "cloudstack_project" "membership_user" {
  name        = "tf-tf-project-user-%[1]s"
  displaytext = "tf-tf-project-user-%[1]s"
  domain      = cloudstack_domain.membership_user.name
  account     = cloudstack_account.owner.username
}

resource "cloudstack_project_membership" "user" {
  project_id = cloudstack_project.membership_user.id
  username   = cloudstack_account.member.username
}
`, suffix)
}
