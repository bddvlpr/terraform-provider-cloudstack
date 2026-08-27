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
	"testing"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCloudStackNetworkPermission_basic(t *testing.T) {
	var permission cloudstack.NetworkPermission

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			requireMinimumCloudStackVersion(t, 4017, "Network permissions")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackNetworkPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackNetworkPermissionBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackNetworkPermissionExists(
						"cloudstack_network_permission.project", &permission),
					resource.TestCheckResourceAttrPair(
						"cloudstack_network_permission.project", "network_id",
						"cloudstack_network.l2", "id"),
					resource.TestCheckResourceAttrPair(
						"cloudstack_network_permission.project", "project_id",
						"cloudstack_project.network_permission", "id"),
				),
			},
			{
				Config: testAccCloudStackNetworkPermissionWithoutPermission,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackNetworkPermissionMissing(&permission),
				),
			},
		},
	})
}

func testAccCheckCloudStackNetworkPermissionExists(n string, permission *cloudstack.NetworkPermission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Network Permission resource not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network Permission ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		networkID := rs.Primary.Attributes["network_id"]
		projectID := rs.Primary.Attributes["project_id"]
		p := cs.Network.NewListNetworkPermissionsParams(networkID)
		r, err := cs.Network.ListNetworkPermissions(p)
		if err != nil {
			return err
		}

		for _, candidate := range r.NetworkPermissions {
			if candidate.Networkid == networkID && candidate.Projectid == projectID {
				*permission = *candidate
				return nil
			}
		}

		return fmt.Errorf("Network Permission for project %s on network %s not found", projectID, networkID)
	}
}

func testAccCheckCloudStackNetworkPermissionMissing(permission *cloudstack.NetworkPermission) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		p := cs.Network.NewListNetworkPermissionsParams(permission.Networkid)
		r, err := cs.Network.ListNetworkPermissions(p)
		if err != nil {
			return err
		}

		for _, candidate := range r.NetworkPermissions {
			if candidate.Networkid == permission.Networkid && candidate.Projectid == permission.Projectid {
				return fmt.Errorf("Network Permission for project %s on network %s still exists", permission.Projectid, permission.Networkid)
			}
		}

		return nil
	}
}

func testAccCheckCloudStackNetworkPermissionDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_network_permission" {
			continue
		}

		networkID := rs.Primary.Attributes["network_id"]
		projectID := rs.Primary.Attributes["project_id"]
		p := cs.Network.NewListNetworkPermissionsParams(networkID)
		r, err := cs.Network.ListNetworkPermissions(p)
		if err != nil {
			if networkPermissionNetworkNotFound(err) {
				continue
			}
			return err
		}

		for _, permission := range r.NetworkPermissions {
			if permission.Networkid == networkID && permission.Projectid == projectID {
				return fmt.Errorf("Network Permission for project %s on network %s still exists", projectID, networkID)
			}
		}
	}

	return nil
}

const testAccCloudStackNetworkPermissionBasic = `
resource "cloudstack_project" "network_permission" {
  name        = "terraform-network-permission"
  displaytext = "Terraform Network Permission"
}

resource "cloudstack_network" "l2" {
  name             = "terraform-network-permission-l2"
  display_text     = "terraform-network-permission-l2"
  network_offering = "DefaultL2NetworkOffering"
  zone             = "Sandbox-simulator"
}

resource "cloudstack_network_permission" "project" {
  network_id = cloudstack_network.l2.id
  project_id = cloudstack_project.network_permission.id
}
`

const testAccCloudStackNetworkPermissionWithoutPermission = `
resource "cloudstack_project" "network_permission" {
  name        = "terraform-network-permission"
  displaytext = "Terraform Network Permission"
}

resource "cloudstack_network" "l2" {
  name             = "terraform-network-permission-l2"
  display_text     = "terraform-network-permission-l2"
  network_offering = "DefaultL2NetworkOffering"
  zone             = "Sandbox-simulator"
}
`
