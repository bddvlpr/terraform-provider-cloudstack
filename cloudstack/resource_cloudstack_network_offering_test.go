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

// TestAccCloudStackNetworkOffering_update creates a network offering and then updates its
// display_text in a second step, exercising resourceCloudStackNetworkOfferingUpdate. It guards
// against the update reading state back through the wrong resource's Read function.
func TestAccCloudStackNetworkOffering_update(t *testing.T) {
	name := "tf-acc-no-" + resource.UniqueId()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackNetworkOfferingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkOfferingUpdateConfig(name, "display text one"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudstack_network_offering.foo", "name", name),
					resource.TestCheckResourceAttr("cloudstack_network_offering.foo", "display_text", "display text one"),
				),
			},
			{
				Config: testAccNetworkOfferingUpdateConfig(name, "display text two"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("cloudstack_network_offering.foo", "name", name),
					resource.TestCheckResourceAttr("cloudstack_network_offering.foo", "display_text", "display text two"),
				),
			},
		},
	})
}

func TestAccCloudStackNetworkOffering_tags(t *testing.T) {
	const resourceName = "cloudstack_network_offering.tags"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackNetworkOfferingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackNetworkOfferingTagsConfig("terraform,network-offering"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags", "terraform,network-offering"),
					testAccCheckCloudStackNetworkOfferingTags(resourceName, "terraform,network-offering"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudStackNetworkOfferingTagsConfig("terraform,updated"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags", "terraform,updated"),
					testAccCheckCloudStackNetworkOfferingTags(resourceName, "terraform,updated"),
				),
			},
			{
				Config: testAccCloudStackNetworkOfferingTagsConfig(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags", ""),
					testAccCheckCloudStackNetworkOfferingTags(resourceName, ""),
				),
			},
		},
	})
}

func testAccNetworkOfferingUpdateConfig(name, displayText string) string {
	return fmt.Sprintf(`
resource "cloudstack_network_offering" "foo" {
  name          = "%s"
  display_text  = "%s"
  guest_ip_type = "Isolated"
  traffic_type  = "Guest"
}
`, name, displayText)
}

func testAccCheckCloudStackNetworkOfferingTags(name, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No network offering ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		offering, _, err := cs.NetworkOffering.GetNetworkOfferingByID(rs.Primary.ID)
		if err != nil {
			return err
		}

		if offering.Tags != expected {
			return fmt.Errorf("Expected network offering tags %q, got %q", expected, offering.Tags)
		}

		return nil
	}
}

func testAccCheckCloudStackNetworkOfferingDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_network_offering" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No network offering ID is set")
		}

		_, _, err := cs.NetworkOffering.GetNetworkOfferingByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Network offering %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCloudStackNetworkOfferingTagsConfig(tags string) string {
	tagsArgument := ""
	if tags != "" {
		tagsArgument = fmt.Sprintf("  tags         = %q\n", tags)
	}

	return fmt.Sprintf(`
resource "cloudstack_network_offering" "tags" {
  name          = "terraform-network-offering-tags"
  display_text  = "Terraform Network Offering Tags"
  guest_ip_type = "Isolated"
  traffic_type  = "Guest"
%s}
`, tagsArgument)
}
