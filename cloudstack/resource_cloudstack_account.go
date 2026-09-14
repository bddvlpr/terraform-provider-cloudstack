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

func resourceCloudStackAccount() *schema.Resource {
	return &schema.Resource{
		Read:   resourceCloudStackAccountRead,
		Update: resourceCloudStackAccountUpdate,
		Create: resourceCloudStackAccountCreate,
		Delete: resourceCloudStackAccountDelete,
		Importer: &schema.ResourceImporter{
			State: importStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"first_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_type": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"generate_keys": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"api_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceCloudStackAccountCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)
	email := d.Get("email").(string)
	first_name := d.Get("first_name").(string)
	last_name := d.Get("last_name").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	role_id := d.Get("role_id").(string)
	account_type := d.Get("account_type").(int)
	account := d.Get("account").(string)
	domain_id := d.Get("domain_id").(string)

	// Create a new parameter struct
	p := cs.Account.NewCreateAccountParams(email, first_name, last_name, password, username)
	p.SetAccounttype(int(account_type))
	p.SetRoleid(role_id)
	if account != "" {
		p.SetAccount(account)
	} else {
		p.SetAccount(username)
	}
	if domain_id != "" {
		p.SetDomainid(domain_id)
	}

	log.Printf("[DEBUG] Creating Account %s", getCloudStackAccountName(d))
	a, err := cs.Account.CreateAccount(p)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Account %s successfully created", getCloudStackAccountName(d))
	d.SetId(a.Id)

	if d.Get("generate_keys").(bool) {
		if err := setCloudStackAccountUserID(d, cs, a.User); err != nil {
			return err
		}
		if err := registerCloudStackAccountKeys(d, cs); err != nil {
			return err
		}
	}

	return resourceCloudStackAccountRead(d, meta)
}

func resourceCloudStackAccountRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	a, count, err := cs.Account.GetAccountByID(d.Id())
	if err != nil {
		if count == 0 || strings.Contains(err.Error(), "Unable to find account") {
			log.Printf("[DEBUG] Account %s does not exist", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Account: %s", err)
	}

	d.Set("account", a.Name)
	d.Set("account_type", a.Accounttype)
	d.Set("domain_id", a.Domainid)
	d.Set("role_id", a.Roleid)

	if err := setCloudStackAccountUserData(d, a); err != nil {
		return err
	}

	if !d.Get("generate_keys").(bool) {
		return clearCloudStackAccountKeys(d)
	}

	if d.Get("user_id").(string) == "" {
		if err := setCloudStackAccountUserID(d, cs, nil); err != nil {
			return err
		}
	}
	if err := readCloudStackAccountKeys(d, cs); err != nil {
		return err
	}

	return nil
}

func resourceCloudStackAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	if d.HasChange("account") || d.HasChange("role_id") {
		p := cs.Account.NewUpdateAccountParams()
		p.SetId(d.Id())

		if d.HasChange("account") {
			p.SetNewname(getCloudStackAccountName(d))
		}
		if d.HasChange("role_id") {
			p.SetRoleid(d.Get("role_id").(string))
		}

		if _, err := cs.Account.UpdateAccount(p); err != nil {
			return fmt.Errorf("Error updating Account: %s", err)
		}
	}

	if d.HasChange("email") || d.HasChange("first_name") || d.HasChange("last_name") ||
		d.HasChange("password") || d.HasChange("username") {
		if d.Get("user_id").(string) == "" {
			if err := setCloudStackAccountUserID(d, cs, nil); err != nil {
				return err
			}
		}

		p := cs.User.NewUpdateUserParams(d.Get("user_id").(string))
		if d.HasChange("email") {
			p.SetEmail(d.Get("email").(string))
		}
		if d.HasChange("first_name") {
			p.SetFirstname(d.Get("first_name").(string))
		}
		if d.HasChange("last_name") {
			p.SetLastname(d.Get("last_name").(string))
		}
		if d.HasChange("password") {
			p.SetPassword(d.Get("password").(string))
		}
		if d.HasChange("username") {
			p.SetUsername(d.Get("username").(string))
		}

		if _, err := cs.User.UpdateUser(p); err != nil {
			return fmt.Errorf("Error updating Account primary user: %s", err)
		}
	}

	if d.HasChange("generate_keys") && d.Get("generate_keys").(bool) {
		if d.Get("user_id").(string) == "" {
			if err := setCloudStackAccountUserID(d, cs, nil); err != nil {
				return err
			}
		}
		if err := registerCloudStackAccountKeys(d, cs); err != nil {
			return err
		}
	}

	return resourceCloudStackAccountRead(d, meta)
}

func resourceCloudStackAccountDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Create a new parameter struct
	p := cs.Account.NewDeleteAccountParams(d.Id())
	_, err := cs.Account.DeleteAccount(p)

	if err != nil {
		return fmt.Errorf("Error deleting Account: %s", err)
	}

	return nil
}

func registerCloudStackAccountKeys(d *schema.ResourceData, cs *cloudstack.CloudStackClient) error {
	userID := d.Get("user_id").(string)
	if userID == "" {
		return fmt.Errorf("Error registering Account keys: primary user ID is empty")
	}

	p := cs.User.NewRegisterUserKeysParams(userID)
	keys, err := cs.User.RegisterUserKeys(p)
	if err != nil {
		return fmt.Errorf("Error registering Account keys: %s", err)
	}

	if err := d.Set("api_key", keys.Apikey); err != nil {
		return err
	}
	if err := d.Set("secret_key", keys.Secretkey); err != nil {
		return err
	}

	return nil
}

func readCloudStackAccountKeys(d *schema.ResourceData, cs *cloudstack.CloudStackClient) error {
	userID := d.Get("user_id").(string)
	if userID == "" {
		return fmt.Errorf("Error reading Account keys: primary user ID is empty")
	}

	p := cs.User.NewGetUserKeysParams(userID)
	keys, err := cs.User.GetUserKeys(p)
	if err != nil {
		return fmt.Errorf("Error reading Account keys: %s", err)
	}

	if err := d.Set("api_key", keys.Apikey); err != nil {
		return err
	}
	if err := d.Set("secret_key", keys.Secretkey); err != nil {
		return err
	}

	return nil
}

func clearCloudStackAccountKeys(d *schema.ResourceData) error {
	if err := d.Set("api_key", ""); err != nil {
		return err
	}
	if err := d.Set("secret_key", ""); err != nil {
		return err
	}

	return nil
}

func setCloudStackAccountUserID(d *schema.ResourceData, cs *cloudstack.CloudStackClient, users []cloudstack.CreateAccountResponseUser) error {
	if err := setCloudStackAccountUserIDFromCreateResponse(d, users); err != nil {
		return err
	}
	if d.Get("user_id").(string) != "" {
		return nil
	}

	accountName := getCloudStackAccountName(d)
	username := d.Get("username").(string)

	p := cs.User.NewListUsersParams()
	p.SetAccount(accountName)
	p.SetUsername(username)
	if domainID := d.Get("domain_id").(string); domainID != "" {
		p.SetDomainid(domainID)
	}

	usersResponse, err := cs.User.ListUsers(p)
	if err != nil {
		return fmt.Errorf("Error finding Account primary user: %s", err)
	}
	if usersResponse.Count == 0 {
		return fmt.Errorf("Error finding Account primary user: no user found for account %q and username %q", accountName, username)
	}
	if usersResponse.Count > 1 {
		return fmt.Errorf("Error finding Account primary user: multiple users found for account %q and username %q", accountName, username)
	}

	return d.Set("user_id", usersResponse.Users[0].Id)
}

func setCloudStackAccountUserIDFromCreateResponse(d *schema.ResourceData, users []cloudstack.CreateAccountResponseUser) error {
	username := d.Get("username").(string)

	for _, user := range users {
		if user.Username == username {
			return d.Set("user_id", user.Id)
		}
	}

	return nil
}

func setCloudStackAccountUserData(d *schema.ResourceData, account *cloudstack.Account) error {
	username := d.Get("username").(string)

	if username != "" {
		for _, user := range account.User {
			if user.Username == username {
				return setCloudStackAccountUserDataFields(d, user)
			}
		}

		return nil
	}

	if len(account.User) == 0 {
		return nil
	}
	if len(account.User) > 1 {
		return fmt.Errorf("Error reading Account primary user: multiple users found for account %q", account.Name)
	}

	return setCloudStackAccountUserDataFields(d, account.User[0])
}

func setCloudStackAccountUserDataFields(d *schema.ResourceData, user cloudstack.AccountUser) error {
	if err := d.Set("email", user.Email); err != nil {
		return err
	}
	if err := d.Set("first_name", user.Firstname); err != nil {
		return err
	}
	if err := d.Set("last_name", user.Lastname); err != nil {
		return err
	}
	if err := d.Set("username", user.Username); err != nil {
		return err
	}
	return d.Set("user_id", user.Id)
}

func getCloudStackAccountName(d *schema.ResourceData) string {
	if account := d.Get("account").(string); account != "" {
		return account
	}
	return d.Get("username").(string)
}
