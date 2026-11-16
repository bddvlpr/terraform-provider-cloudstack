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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCloudStackInstance_basic(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceAttributes(&instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "user_data", "0cf3dcdc356ec8369494cb3991985ecd5296cdd5"),
					testAccCheckResourceTags(&instance),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_stopped(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_stopped,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceAttributes(&instance),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_update(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceAttributes(&instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "user_data", "0cf3dcdc356ec8369494cb3991985ecd5296cdd5"),
				),
			},

			{
				Config: testAccCloudStackInstance_renameAndResize,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceRenamedAndResized(&instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "name", "terraform-updated"),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "display_name", "terraform-updated"),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "service_offering", "Medium Instance"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_extraConfig(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_extraConfig(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "extraconfig", ""),
				),
			},
			{
				// CloudStack's simulator hypervisor does not support extra
				// configuration. The expected error verifies that the provider
				// forwards extraconfig when updating an instance.
				Config: testAccCloudStackInstance_extraConfig(
					"%3Cclock%20offset%3D%22localtime%22%2F%3E"),
				ExpectError: regexp.MustCompile(
					"This hypervisor Simulator is not supported for use with this feature"),
			},
			{
				Config: testAccCloudStackInstance_extraConfig(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					testAccCheckCloudStackInstanceState(&instance, "Running"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_fixedIP(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_fixedIP,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "ip_address", "10.1.1.123"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_keyPair(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_keyPair,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "keypair", "terraform-test-keypair"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_keyPairs(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_keyPairs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists("cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "keypairs.#", "2"), // Expecting 2 key pairs
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "keypairs.0", "terraform-test-keypair-foo"),
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "keypairs.1", "terraform-test-keypair-bar"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_keyPair_update(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_keyPair,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "keypair", "terraform-test-keypair"),
				),
			},

			{
				Config: testAccCloudStackInstance_keyPairs,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists("cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "keypairs.#", "2"), // Expecting 2 key pairs
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "keypairs.0", "terraform-test-keypair-foo"),
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "keypairs.1", "terraform-test-keypair-bar"),
				),
			},

			{
				Config: testAccCloudStackInstance_keyPair,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "keypair", "terraform-test-keypair"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_project(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_project,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "project", "terraform"),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_networkProjectInheritance(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_networkProjectInheritance,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists(
						"cloudstack_instance.foobar", &instance),
					// Verify the project was inherited from the network
					resource.TestCheckResourceAttr(
						"cloudstack_instance.foobar", "project", "terraform"),
					testAccCheckCloudStackInstanceProjectInherited(&instance),
				),
			},
		},
	})
}

func TestAccCloudStackInstance_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_basic,
			},

			{
				ResourceName:            "cloudstack_instance.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expunge", "user_data", "uefi"},
			},
		},
	})
}

func TestAccCloudStackInstance_importProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackInstance_project,
			},

			{
				ResourceName:            "cloudstack_instance.foobar",
				ImportState:             true,
				ImportStateIdPrefix:     "terraform/",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expunge", "user_data", "uefi"},
			},
		},
	})
}

func TestAccCloudStackInstance_userData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		ExternalProviders: map[string]resource.ExternalProvider{
			"random": {
				VersionConstraint: ">= 3.6.0",
				Source:            "hashicorp/random",
			},
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccCloudStackInstance_userData,
				ExpectError: regexp.MustCompile("User data has exceeded configurable max length"),
			},
		},
	})
}

func TestAccCloudStackInstance_deleteProtection(t *testing.T) {
	var instance cloudstack.VirtualMachine

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackInstanceDestroy,
		Steps: []resource.TestStep{
			{
				// create vm with delete protection enabled
				Config: fmt.Sprintf(testAccCloudStackInstance_deleteProtection, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists("cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "delete_protection", "true"),
				),
			},
			{
				// attempt to destroy vm. expected to fail due to delete protection is enabled
				Config:      fmt.Sprintf(testAccCloudStackInstance_deleteProtection, true),
				Destroy:     true,
				ExpectError: regexp.MustCompile(".*has delete protection enabled and cannot be deleted\\."),
			},
			{
				// disable delete protection
				Config: fmt.Sprintf(testAccCloudStackInstance_deleteProtection, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackInstanceExists("cloudstack_instance.foobar", &instance),
					resource.TestCheckResourceAttr("cloudstack_instance.foobar", "delete_protection", "false"),
				),
			},
			{
				// destroy vm. expected to pass due to delete protection is disabled
				Config:  fmt.Sprintf(testAccCloudStackInstance_deleteProtection, false),
				Destroy: true,
			},
		},
	})
}

func testAccCheckCloudStackInstanceExists(
	n string, instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		vm, _, err := cs.VirtualMachine.GetVirtualMachineByID(
			rs.Primary.ID,
			cloudstack.WithProject(rs.Primary.Attributes["project"]),
		)

		if err != nil {
			return err
		}

		if vm.Id != rs.Primary.ID {
			return fmt.Errorf("Instance not found")
		}

		*instance = *vm

		return nil
	}
}

func testAccCheckCloudStackInstanceAttributes(
	instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if instance.Name != "terraform-test" {
			return fmt.Errorf("Bad name: %s", instance.Name)
		}

		if instance.Displayname != "terraform-test" {
			return fmt.Errorf("Bad display name: %s", instance.Displayname)
		}

		if instance.Serviceofferingname != "Small Instance" {
			return fmt.Errorf("Bad service offering: %s", instance.Serviceofferingname)
		}

		if instance.Templatename != "CentOS 5.6 (64-bit) no GUI (Simulator)" {
			return fmt.Errorf("Bad template: %s", instance.Templatename)
		}

		return nil
	}
}

func testAccCheckCloudStackInstanceRenamedAndResized(
	instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if instance.Name != "terraform-updated" {
			return fmt.Errorf("Bad name: %s", instance.Name)
		}

		if instance.Displayname != "terraform-updated" {
			return fmt.Errorf("Bad display name: %s", instance.Displayname)
		}

		if instance.Serviceofferingname != "Medium Instance" {
			return fmt.Errorf("Bad service offering: %s", instance.Serviceofferingname)
		}

		return nil
	}
}

func testAccCheckCloudStackInstanceState(
	instance *cloudstack.VirtualMachine, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if instance.State != expected {
			return fmt.Errorf("Bad state: expected %s, got %s", expected, instance.State)
		}

		return nil
	}
}

func testAccCheckCloudStackInstanceProjectInherited(
	instance *cloudstack.VirtualMachine) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if instance.Project != "terraform" {
			return fmt.Errorf("Expected project to be 'terraform' (inherited from network), got: %s", instance.Project)
		}

		return nil
	}
}

func testAccCheckCloudStackInstanceDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_instance" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No instance ID is set")
		}

		_, _, err := cs.VirtualMachine.GetVirtualMachineByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Virtual Machine %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccCloudStackInstance_basic = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  user_data = "foobar\nfoo\nbar"
  expunge = true
  tags = {
    terraform-tag = "true"
  }
}`

func testAccCloudStackInstance_extraConfig(extraConfig string) string {
	return fmt.Sprintf(`
resource "cloudstack_configuration" "extra_config" {
  name = "enable.additional.vm.configuration"
  value = "true"
}

resource "cloudstack_network" "foo" {
  name = "terraform-network-extra-config"
  display_text = "terraform-network-extra-config"
  cidr = "10.1.2.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-extra-config"
  display_name = "terraform-extra-config"
  service_offering = "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  expunge = true
  extraconfig = %q

  depends_on = [cloudstack_configuration.extra_config]
}`, extraConfig)
}

const testAccCloudStackInstance_stopped = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  start_vm = false
  expunge = true
}`

const testAccCloudStackInstance_renameAndResize = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-updated"
  display_name = "terraform-updated"
  service_offering= "Medium Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  user_data = "foobar\nfoo\nbar"
  expunge = true
}`

const testAccCloudStackInstance_fixedIP = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  ip_address = "10.1.1.123"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  expunge = true
}`

const testAccCloudStackInstance_keyPair = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_ssh_keypair" "foo" {
  name = "terraform-test-keypair"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  keypair = cloudstack_ssh_keypair.foo.name
  expunge = true
}`

const testAccCloudStackInstance_keyPairs = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_ssh_keypair" "foo" {
  name = "terraform-test-keypair-foo"
}

resource "cloudstack_ssh_keypair" "bar" {
  name = "terraform-test-keypair-bar"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = "${cloudstack_network.foo.id}"
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  keypairs = ["${cloudstack_ssh_keypair.foo.name}", "${cloudstack_ssh_keypair.bar.name}"]
  expunge = true
}`

const testAccCloudStackInstance_project = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  project = "terraform"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  project = "terraform"
  zone = cloudstack_network.foo.zone
  expunge = true
}`

const testAccCloudStackInstance_userData = `
resource "random_bytes" "string" {
  length = 32768
}

resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = cloudstack_network.foo.zone
  expunge = true
  user_data = <<-EOFTF
#!/bin/bash

echo <<EOF
${random_bytes.string.base64}
EOF
EOFTF
}`

const testAccCloudStackInstance_deleteProtection = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = "Sandbox-simulator"
  user_data = "foobar\nfoo\nbar"
  expunge = true
  delete_protection = %t
  tags = {
    terraform-tag = "true"
  }
}`

const testAccCloudStackInstance_networkProjectInheritance = `
resource "cloudstack_network" "foo" {
  name = "terraform-network"
  display_text = "terraform-network"
  cidr = "10.1.1.0/24"
  network_offering = "DefaultIsolatedNetworkOfferingWithSourceNatService"
  project = "terraform"
  zone = "Sandbox-simulator"
}

resource "cloudstack_instance" "foobar" {
  name = "terraform-test"
  display_name = "terraform-test"
  service_offering= "Small Instance"
  network_id = cloudstack_network.foo.id
  template = "CentOS 5.6 (64-bit) no GUI (Simulator)"
  zone = cloudstack_network.foo.zone
  expunge = true
  # Note: project is NOT specified here - it should be inherited from the network
}`

func TestAccCloudStackInstance_extraConfigs(t *testing.T) {
	for _, start := range []bool{true, false} {
		t.Run(fmt.Sprintf("start_vm=%t", start), func(t *testing.T) {
			var instance cloudstack.VirtualMachine
			var originalID string
			a := "<metadata><a>" + strings.Repeat("a", 1900) + "</a></metadata>"
			b := "<metadata><b>" + strings.Repeat("b", 1900) + "</b></metadata>"
			c := "<metadata><c>" + strings.Repeat("c", 1900) + "</c></metadata>"
			config := func(values []string) string {
				encoded := make([]string, len(values))
				for i, value := range values {
					encoded[i] = fmt.Sprintf("%q", url.QueryEscape(value))
				}
				return strings.Replace(strings.Replace(testAccCloudStackInstance_extraConfig(""), "expunge = true", fmt.Sprintf("expunge = true\n start_vm = %t", start), 1), "extraconfig = \"\"", "extraconfigs = ["+strings.Join(encoded, ",")+"]", 1)
			}
			check := func(values []string) resource.TestCheckFunc {
				return func(s *terraform.State) error {
					expectedState := "Stopped"
					if start {
						expectedState = "Running"
					}
					if instance.State != expectedState {
						return fmt.Errorf("instance state %s, want %s", instance.State, expectedState)
					}
					if originalID == "" {
						originalID = instance.Id
					} else if instance.Id != originalID {
						return fmt.Errorf("instance was replaced")
					}
					remote, err := listInstanceExtraConfigs(testAccProvider.Meta().(*cloudstack.CloudStackClient), instance.Id)
					if err != nil {
						return err
					}
					rs := s.RootModule().Resources["cloudstack_instance.foobar"]
					if rs.Primary.Attributes["extraconfigs.#"] != fmt.Sprint(len(values)) {
						return fmt.Errorf("Terraform list length differs")
					}
					for i, value := range values {
						if rs.Primary.Attributes[fmt.Sprintf("extraconfigs.%d", i)] != url.QueryEscape(value) {
							return fmt.Errorf("Terraform fragment %d differs", i)
						}
					}
					count := 0
					for key, value := range remote {
						if !strings.HasPrefix(key, "extraconfig-") {
							continue
						}
						count++
						if len(values) == 2 && value == b {
							return fmt.Errorf("deleted middle fragment still exists at %s", key)
						}
					}
					if count != len(values) {
						return fmt.Errorf("got %d extra configurations, want %d", count, len(values))
					}
					for i, value := range values {
						if remote[fmt.Sprintf("extraconfig-%d", i+1)] != value {
							return fmt.Errorf("remote configuration %d differs", i+1)
						}
					}
					return nil
				}
			}
			steps := []resource.TestStep{}
			for _, values := range [][]string{{a, b, c}, {a, c}, {a, "<metadata>updated</metadata>"}, {c}, {}} {
				steps = append(steps, resource.TestStep{
					Config: config(values),
					Check:  resource.ComposeTestCheckFunc(testAccCheckCloudStackInstanceExists("cloudstack_instance.foobar", &instance), check(values)),
				})
			}
			resource.Test(t, resource.TestCase{PreCheck: func() { testAccPreCheck(t) }, Providers: testAccProviders, CheckDestroy: testAccCheckCloudStackInstanceDestroy, Steps: steps})
		})
	}
}

func TestInstanceExtraConfigsReconcile(t *testing.T) {
	stored := map[string]string{"extraconfig-1": "A", "extraconfig-2": "B", "extraconfig-3": "C", "unrelated": "keep"}
	failAdd := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		switch r.Form.Get("command") {
		case "listResourceDetails":
			details := []map[string]string{}
			for k, v := range stored {
				details = append(details, map[string]string{"key": k, "value": v})
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"listresourcedetailsresponse": map[string]interface{}{"count": len(details), "resourcedetail": details}})
		case "addResourceDetail":
			if failAdd {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"addresourcedetailresponse":{"errorcode":530,"errortext":"test failure"}}`))
				return
			}
			for i := 0; ; i++ {
				key := r.Form.Get("details[" + strconv.Itoa(i) + "].key")
				if key == "" {
					break
				}
				stored[key] = r.Form.Get("details[" + strconv.Itoa(i) + "].value")
			}
			w.Write([]byte(`{"addresourcedetailresponse":{"success":true}}`))
		case "removeResourceDetail":
			delete(stored, r.Form.Get("key"))
			w.Write([]byte(`{"removeresourcedetailresponse":{"success":true}}`))
		default:
			t.Errorf("unexpected command %q", r.Form.Get("command"))
		}
	}))
	defer server.Close()
	cs := cloudstack.NewClient(server.URL, "key", "secret", false)
	if err := reconcileInstanceExtraConfigs(cs, "vm", nil, []interface{}{"replacement"}); err == nil {
		t.Fatal("accepted collision with existing setting")
	}
	if stored["extraconfig-1"] != "A" {
		t.Fatal("collision mutated settings")
	}
	old := []interface{}{"A", "B", "C"}
	value := "<metadata>é + %</metadata>"
	next := []interface{}{"A", url.QueryEscape(value)}
	if err := reconcileInstanceExtraConfigs(cs, "vm", old, next); err != nil {
		t.Fatal(err)
	}
	if stored["extraconfig-2"] != value {
		t.Fatalf("encoding or replacement failed: %q", stored["extraconfig-2"])
	}
	if _, ok := stored["extraconfig-3"]; ok {
		t.Fatal("obsolete setting survives")
	}
	if stored["unrelated"] != "keep" {
		t.Fatal("unrelated setting changed")
	}
	failAdd = true
	if err := reconcileInstanceExtraConfigs(cs, "vm", next, []interface{}{"replacement"}); err == nil {
		t.Fatal("expected API error")
	}
	if stored["extraconfig-2"] != value {
		t.Fatal("settings removed after failed write")
	}
	failAdd = false
	if err := reconcileInstanceExtraConfigs(cs, "vm", next, nil); err != nil {
		t.Fatal(err)
	}
	if len(stored) != 1 {
		t.Fatalf("clear left settings: %v", stored)
	}
}

func TestInstanceExtraConfigValidation(t *testing.T) {
	for _, value := range []string{"", "%zz"} {
		if _, errs := validateInstanceExtraConfig(value, "extraconfigs"); len(errs) == 0 {
			t.Fatalf("accepted %q", value)
		}
	}
	if _, errs := validateInstanceExtraConfig(url.QueryEscape("<metadata>é + %</metadata>"), "extraconfigs"); len(errs) > 0 {
		t.Fatal(errs)
	}
}
