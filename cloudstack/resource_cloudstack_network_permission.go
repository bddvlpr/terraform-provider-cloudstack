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
	"log"
	"strings"

	"github.com/apache/cloudstack-go/v2/cloudstack"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudStackNetworkPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudStackNetworkPermissionCreate,
		Read:   resourceCloudStackNetworkPermissionRead,
		Delete: resourceCloudStackNetworkPermissionDelete,

		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the network to share with the project.",
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the project that is allowed to use the network.",
			},
		},
	}
}

func resourceCloudStackNetworkPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)
	networkID := d.Get("network_id").(string)
	projectID := d.Get("project_id").(string)

	p := cs.Network.NewCreateNetworkPermissionsParams(networkID)
	p.SetProjectids([]string{projectID})

	log.Printf("[DEBUG] Granting project %s permission to use network %s", projectID, networkID)
	r, err := cs.Network.CreateNetworkPermissions(p)
	if err != nil {
		return fmt.Errorf("error creating Network Permission for project %s on network %s: %s", projectID, networkID, err)
	}
	if !r.Success {
		return fmt.Errorf("error creating Network Permission for project %s on network %s: %s", projectID, networkID, r.Displaytext)
	}

	d.SetId(fmt.Sprintf("%s/%s", networkID, projectID))
	return resourceCloudStackNetworkPermissionRead(d, meta)
}

func resourceCloudStackNetworkPermissionRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)
	networkID := d.Get("network_id").(string)
	projectID := d.Get("project_id").(string)

	p := cs.Network.NewListNetworkPermissionsParams(networkID)
	r, err := cs.Network.ListNetworkPermissions(p)
	if err != nil {
		if networkPermissionNetworkNotFound(err) {
			log.Printf("[DEBUG] Network %s no longer exists", networkID)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error listing Network Permissions for network %s: %s", networkID, err)
	}

	for _, permission := range r.NetworkPermissions {
		if permission.Networkid == networkID && permission.Projectid == projectID {
			return nil
		}
	}

	log.Printf("[DEBUG] Project %s no longer has permission to use network %s", projectID, networkID)
	d.SetId("")
	return nil
}

func resourceCloudStackNetworkPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)
	networkID := d.Get("network_id").(string)
	projectID := d.Get("project_id").(string)

	p := cs.Network.NewRemoveNetworkPermissionsParams(networkID)
	p.SetProjectids([]string{projectID})

	log.Printf("[DEBUG] Removing project %s permission to use network %s", projectID, networkID)
	r, err := cs.Network.RemoveNetworkPermissions(p)
	if err != nil {
		if networkPermissionNetworkNotFound(err) {
			return nil
		}
		return fmt.Errorf("error removing Network Permission for project %s on network %s: %s", projectID, networkID, err)
	}
	if !r.Success {
		return fmt.Errorf("error removing Network Permission for project %s on network %s: %s", projectID, networkID, r.Displaytext)
	}

	return nil
}

func networkPermissionNetworkNotFound(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "unable to find network with id") ||
		strings.Contains(message, "entity does not exist")
}
