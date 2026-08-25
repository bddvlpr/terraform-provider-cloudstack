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

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestResourceCloudStackSystemServiceOfferingSchema(t *testing.T) {
	t.Parallel()

	r := resourceCloudStackSystemServiceOffering()
	for _, name := range []string{"name", "display_text", "system_vm_type", "cpu_number", "cpu_speed", "memory"} {
		if !r.Schema[name].Required {
			t.Errorf("expected %q to be required", name)
		}
	}

	for _, name := range []string{"system_vm_type", "cpu_number", "cpu_speed", "memory", "storage_type", "network_rate", "offer_ha", "limit_cpu_use"} {
		if !r.Schema[name].ForceNew {
			t.Errorf("expected %q to force replacement", name)
		}
	}

	for _, name := range []string{"name", "display_text", "host_tags", "storage_tags", "domain_ids"} {
		if r.Schema[name].ForceNew {
			t.Errorf("expected %q to be updateable", name)
		}
	}

	if r.Schema["domain_ids"].Type != schema.TypeSet {
		t.Errorf("expected domain_ids to be a set")
	}
	if got := r.Schema["storage_type"].Default; got != "shared" {
		t.Errorf("expected storage_type default to be shared, got %v", got)
	}
}

func TestResourceCloudStackSystemServiceOfferingValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		systemVMType   string
		hasNetworkRate bool
		storageType    string
		offerHA        bool
		wantError      string
	}{
		{
			name:           "domain router network rate",
			systemVMType:   "domainrouter",
			hasNetworkRate: true,
			storageType:    "shared",
		},
		{
			name:           "network rate on console proxy",
			systemVMType:   "consoleproxy",
			hasNetworkRate: true,
			storageType:    "shared",
			wantError:      "network_rate",
		},
		{
			name:         "HA with local storage",
			systemVMType: "domainrouter",
			storageType:  "local",
			offerHA:      true,
			wantError:    "offer_ha",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateSystemServiceOfferingConfiguration(test.systemVMType, test.hasNetworkRate, test.storageType, test.offerHA)
			if test.wantError == "" && err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if test.wantError != "" && (err == nil || !strings.Contains(err.Error(), test.wantError)) {
				t.Fatalf("expected error containing %q, got %v", test.wantError, err)
			}
		})
	}
}

func TestAccSystemServiceOffering(t *testing.T) {
	var offering cloudstack.ServiceOffering
	var originalID string
	const resourceName = "cloudstack_system_service_offering.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackSystemServiceOfferingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudStackSystemServiceOfferingConfig("terraform-system-offering", "Terraform System Offering", 1, "compute", "primary"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackSystemServiceOfferingExists(resourceName, &offering),
					testAccCaptureCloudStackSystemServiceOfferingID(resourceName, &originalID),
					resource.TestCheckResourceAttr(resourceName, "system_vm_type", "domainrouter"),
					resource.TestCheckResourceAttr(resourceName, "cpu_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "cpu_speed", "500"),
					resource.TestCheckResourceAttr(resourceName, "memory", "256"),
					resource.TestCheckResourceAttr(resourceName, "storage_type", "shared"),
					resource.TestCheckResourceAttr(resourceName, "network_rate", "100"),
					resource.TestCheckResourceAttr(resourceName, "offer_ha", "true"),
					resource.TestCheckResourceAttr(resourceName, "limit_cpu_use", "true"),
				),
			},
			{
				Config: testAccCloudStackSystemServiceOfferingConfig("terraform-system-offering-updated", "Terraform System Offering Updated", 1, "compute-updated", "secondary"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackSystemServiceOfferingExists(resourceName, &offering),
					testAccCheckCloudStackSystemServiceOfferingID(resourceName, &originalID, false),
					resource.TestCheckResourceAttr(resourceName, "name", "terraform-system-offering-updated"),
					resource.TestCheckResourceAttr(resourceName, "display_text", "Terraform System Offering Updated"),
					resource.TestCheckResourceAttr(resourceName, "host_tags", "compute-updated"),
					resource.TestCheckResourceAttr(resourceName, "storage_tags", "secondary"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCloudStackSystemServiceOfferingConfig("terraform-system-offering-updated", "Terraform System Offering Updated", 2, "compute-updated", "secondary"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackSystemServiceOfferingExists(resourceName, &offering),
					testAccCheckCloudStackSystemServiceOfferingID(resourceName, &originalID, true),
					resource.TestCheckResourceAttr(resourceName, "cpu_number", "2"),
				),
			},
		},
	})
}

func testAccCloudStackSystemServiceOfferingConfig(name, displayText string, cpuNumber int, hostTags, storageTags string) string {
	return fmt.Sprintf(`
resource "cloudstack_system_service_offering" "test" {
  name           = %q
  display_text   = %q
  system_vm_type = "domainrouter"
  cpu_number     = %d
  cpu_speed      = 500
  memory         = 256
  network_rate   = 100
  offer_ha       = true
  limit_cpu_use  = true
  host_tags      = %q
  storage_tags   = %q
}
`, name, displayText, cpuNumber, hostTags, storageTags)
}

func testAccCheckCloudStackSystemServiceOfferingExists(name string, offering *cloudstack.ServiceOffering) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if resourceState.Primary.ID == "" {
			return fmt.Errorf("no System Service Offering ID is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		result, count, err := getSystemServiceOfferingByID(cs, resourceState.Primary.ID)
		if err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("System Service Offering %s not found", resourceState.Primary.ID)
		}

		*offering = *result
		return nil
	}
}

func testAccCaptureCloudStackSystemServiceOfferingID(name string, id *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok || resourceState.Primary.ID == "" {
			return fmt.Errorf("no System Service Offering ID is set")
		}

		*id = resourceState.Primary.ID
		return nil
	}
}

func testAccCheckCloudStackSystemServiceOfferingID(name string, originalID *string, expectChanged bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[name]
		if !ok || resourceState.Primary.ID == "" {
			return fmt.Errorf("no System Service Offering ID is set")
		}

		changed := resourceState.Primary.ID != *originalID
		if changed != expectChanged {
			return fmt.Errorf("unexpected System Service Offering ID change: original=%s current=%s", *originalID, resourceState.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCloudStackSystemServiceOfferingDestroy(state *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "cloudstack_system_service_offering" {
			continue
		}

		_, count, err := getSystemServiceOfferingByID(cs, resourceState.Primary.ID)
		if err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("System Service Offering %s still exists", resourceState.Primary.ID)
		}
	}

	return nil
}
