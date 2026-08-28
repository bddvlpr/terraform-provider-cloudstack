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

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSecurityGroupDataSource_basic(t *testing.T) {
	resourceName := "cloudstack_security_group.security-group-resource"
	dataSourceName := "data.cloudstack_security_group.security-group-data-source"
	securityGroupName := "terraform-security-group-data-source-" + id.UniqueId()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testSecurityGroupDataSourceConfig(securityGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
				),
			},
		},
	})
}

func testSecurityGroupDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "cloudstack_security_group" "security-group-resource" {
  name        = "%[1]s"
  description = "Security group data source acceptance test"
}

data "cloudstack_security_group" "security-group-data-source" {
  filter {
    name  = "name"
    value = "^%[1]s$"
  }
  depends_on = [cloudstack_security_group.security-group-resource]
}
`,
		name,
	)
}
