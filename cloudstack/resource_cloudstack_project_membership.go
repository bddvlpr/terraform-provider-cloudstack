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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceCloudStackProjectMembership() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudStackProjectMembershipCreate,
		Read:   resourceCloudStackProjectMembershipRead,
		Delete: resourceCloudStackProjectMembershipDelete,

		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"account", "username"},
			},
			"username": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"account", "username"},
			},
			"user_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"account"},
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"role_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Regular",
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"Admin", "Regular"}, false),
			},
			"project_role_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceCloudStackProjectMembershipCreate(d *schema.ResourceData, meta any) error {
	cs := meta.(*cloudstack.CloudStackClient)
	projectID := d.Get("project_id").(string)

	if account, ok := d.GetOk("account"); ok {
		p := cs.Project.NewAddAccountToProjectParams(projectID)
		p.SetAccount(account.(string))
		setProjectMembershipOptions(d, p.SetEmail, p.SetRoletype, p.SetProjectroleid)

		log.Printf("[DEBUG] Adding account %s to project %s", account.(string), projectID)
		if _, err := cs.Project.AddAccountToProject(p); err != nil {
			return fmt.Errorf("error adding account %s to project %s: %s", account.(string), projectID, err)
		}

		d.SetId(projectMembershipID(projectID, "account", account.(string)))
		return resourceCloudStackProjectMembershipRead(d, meta)
	}

	username := d.Get("username").(string)
	p := cs.Project.NewAddUserToProjectParams(projectID, username)
	setProjectMembershipOptions(d, p.SetEmail, p.SetRoletype, p.SetProjectroleid)

	log.Printf("[DEBUG] Adding user %s to project %s", username, projectID)
	if _, err := cs.Project.AddUserToProject(p); err != nil {
		return fmt.Errorf("error adding user %s to project %s: %s", username, projectID, err)
	}

	userID := d.Get("user_id").(string)
	if userID == "" {
		var err error
		userID, err = getUserIDByUsername(cs, username)
		if err != nil {
			return err
		}
		d.Set("user_id", userID)
	}

	d.SetId(projectMembershipID(projectID, "user", userID))
	return resourceCloudStackProjectMembershipRead(d, meta)
}

func resourceCloudStackProjectMembershipRead(d *schema.ResourceData, meta any) error {
	cs := meta.(*cloudstack.CloudStackClient)
	projectID := d.Get("project_id").(string)

	p := cs.Account.NewListProjectAccountsParams(projectID)
	if account, ok := d.GetOk("account"); ok {
		p.SetAccount(account.(string))
	} else {
		userID := d.Get("user_id").(string)
		if userID == "" {
			var err error
			userID, err = getUserIDByUsername(cs, d.Get("username").(string))
			if err != nil {
				if strings.Contains(err.Error(), "No match found") {
					d.SetId("")
					return nil
				}
				return err
			}
			d.Set("user_id", userID)
		}
		p.SetUserid(userID)
	}

	l, err := cs.Account.ListProjectAccounts(p)
	if err != nil {
		return fmt.Errorf("error reading project membership %s: %s", d.Id(), err)
	}

	if l.Count == 0 {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceCloudStackProjectMembershipDelete(d *schema.ResourceData, meta any) error {
	cs := meta.(*cloudstack.CloudStackClient)
	projectID := d.Get("project_id").(string)

	if account, ok := d.GetOk("account"); ok {
		p := cs.Project.NewDeleteAccountFromProjectParams(account.(string), projectID)
		log.Printf("[DEBUG] Removing account %s from project %s", account.(string), projectID)
		if _, err := cs.Project.DeleteAccountFromProject(p); err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
				return nil
			}
			return fmt.Errorf("error removing account %s from project %s: %s", account.(string), projectID, err)
		}
		return nil
	}

	userID := d.Get("user_id").(string)
	if userID == "" {
		var err error
		userID, err = getUserIDByUsername(cs, d.Get("username").(string))
		if err != nil {
			if strings.Contains(err.Error(), "No match found") {
				return nil
			}
			return err
		}
	}

	p := cs.Project.NewDeleteUserFromProjectParams(projectID, userID)
	log.Printf("[DEBUG] Removing user %s from project %s", userID, projectID)
	if _, err := cs.Project.DeleteUserFromProject(p); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return fmt.Errorf("error removing user %s from project %s: %s", userID, projectID, err)
	}

	return nil
}

func setProjectMembershipOptions(
	d *schema.ResourceData,
	setEmail func(string),
	setRoleType func(string),
	setProjectRoleID func(string),
) {
	if email, ok := d.GetOk("email"); ok {
		setEmail(email.(string))
	}
	if roleType, ok := d.GetOk("role_type"); ok {
		setRoleType(roleType.(string))
	}
	if projectRoleID, ok := d.GetOk("project_role_id"); ok {
		setProjectRoleID(projectRoleID.(string))
	}
}

func getUserIDByUsername(cs *cloudstack.CloudStackClient, username string) (string, error) {
	p := cs.User.NewListUsersParams()
	p.SetUsername(username)
	p.SetListall(true)

	l, err := cs.User.ListUsers(p)
	if err != nil {
		return "", fmt.Errorf("error retrieving user %s: %s", username, err)
	}

	if l.Count == 0 {
		return "", fmt.Errorf("No match found for user %s", username)
	}
	if l.Count > 1 {
		return "", fmt.Errorf("multiple users found for username %s", username)
	}

	return l.Users[0].Id, nil
}

func projectMembershipID(projectID, membershipType, memberID string) string {
	return strings.Join([]string{projectID, membershipType, memberID}, "/")
}
