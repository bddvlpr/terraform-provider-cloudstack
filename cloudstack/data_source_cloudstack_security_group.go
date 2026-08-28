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
	"log"
	"regexp"
	"strings"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceCloudstackSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCloudstackSecurityGroupRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),

			//Computed values
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}

func dataSourceCloudstackSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)
	p := cs.SecurityGroup.NewListSecurityGroupsParams()

	// If there is a project supplied, we retrieve and set the project id
	if err := setProjectid(p, cs, d); err != nil {
		return err
	}

	securityGroups, err := cs.SecurityGroup.ListSecurityGroups(p)
	if err != nil {
		return fmt.Errorf("failed to list security groups: %s", err)
	}

	filters := d.Get("filter").(*schema.Set)
	var matches []*cloudstack.SecurityGroup

	for _, securityGroup := range securityGroups.SecurityGroups {
		match, err := applySecurityGroupFilters(securityGroup, filters)
		if err != nil {
			return err
		}
		if match {
			matches = append(matches, securityGroup)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no security group matches the specified filters")
	}
	if len(matches) > 1 {
		return fmt.Errorf("multiple security groups match the specified filters")
	}

	securityGroup := matches[0]
	log.Printf("[DEBUG] Selected security group: %s", securityGroup.Name)

	return securityGroupDescriptionAttributes(d, securityGroup)
}

func securityGroupDescriptionAttributes(d *schema.ResourceData, securityGroup *cloudstack.SecurityGroup) error {
	d.SetId(securityGroup.Id)

	if err := d.Set("name", securityGroup.Name); err != nil {
		return fmt.Errorf("failed to set security group name: %s", err)
	}
	if err := d.Set("description", securityGroup.Description); err != nil {
		return fmt.Errorf("failed to set security group description: %s", err)
	}

	setValueOrID(d, "project", securityGroup.Project, securityGroup.Projectid)

	return nil
}

func applySecurityGroupFilters(securityGroup *cloudstack.SecurityGroup, filters *schema.Set) (bool, error) {
	securityGroupJSON, err := json.Marshal(securityGroup)
	if err != nil {
		return false, fmt.Errorf("failed to encode security group: %s", err)
	}

	var fields map[string]interface{}
	if err := json.Unmarshal(securityGroupJSON, &fields); err != nil {
		return false, fmt.Errorf("failed to decode security group: %s", err)
	}

	for _, filter := range filters.List() {
		values := filter.(map[string]interface{})
		pattern, err := regexp.Compile(values["value"].(string))
		if err != nil {
			return false, fmt.Errorf("invalid regex: %s", err)
		}

		name := strings.ReplaceAll(values["name"].(string), "_", "")
		value, ok := fields[name]
		if !ok {
			return false, fmt.Errorf("field %q does not exist in security group", values["name"].(string))
		}

		if !pattern.MatchString(fmt.Sprint(value)) {
			return false, nil
		}
	}

	return true, nil
}
