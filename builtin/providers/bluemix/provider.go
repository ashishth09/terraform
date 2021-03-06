/*
* Licensed Materials - Property of IBM
* (C) Copyright IBM Corp. 2017. All Rights Reserved.
* US Government Users Restricted Rights - Use, duplication or
* disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
 */

package bluemix

import (
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider for BlueMix
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The BlueMix user name.",
				DefaultFunc: func() (interface{}, error) {
					return ValueFromEnv("username"), nil
				},
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The BlueMix password.",
				DefaultFunc: func() (interface{}, error) {
					return ValueFromEnv("password"), nil
				},
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The BlueMix Region (for example 'ng').",
				DefaultFunc: func() (interface{}, error) {
					return ValueFromEnv("region"), nil
				},
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The timeout (in seconds) to set for any BlueMix API calls made.",
				// TypeInt doesn't need default value to not prompt
			},
			"softlayer_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The SoftLayer user name.",
				DefaultFunc: func() (interface{}, error) {
					return "", nil
				},
			},
			"softlayer_api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The API key for SoftLayer API operations.",
				DefaultFunc: func() (interface{}, error) {
					return "", nil
				},
			},
			"softlayer_endpoint_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The endpoint url for the SoftLayer API.",
				DefaultFunc: func() (interface{}, error) {
					return "", nil
				},
			},
			"softlayer_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The timeout (in seconds) to set for any SoftLayer API calls made.",
				// TypeInt doesn't need default value to not prompt
			},
			"softlayer_account_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The SoftLayer IMS account number.",
				DefaultFunc: func() (interface{}, error) {
					return "", nil
				},
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"bluemix_infrastructure_ssh_key": dataSourceInfrastructureSSHKey(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"bluemix_infrastructure_ssh_key":       resourceInfrastructureSSHKey(),
			"bluemix_infrastructure_virtual_guest": resourceInfrastructureVirtualGuest(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	timeout := ""
	softlayerTimeout := ""
	if rawTimeout, ok := d.GetOk("timeout"); ok {
		timeout = strconv.Itoa(rawTimeout.(int))
	}
	if rawSoftlayerTimeout, ok := d.GetOk("softlayer_timeout"); ok {
		softlayerTimeout = strconv.Itoa(rawSoftlayerTimeout.(int))
	}

	config := Config{
		Username:               d.Get("username").(string),
		SoftlayerAPIKey:        d.Get("softlayer_api_key").(string),
		Password:               d.Get("password").(string),
		Region:                 d.Get("region").(string),
		Timeout:                timeout,
		SoftlayerUsername:      d.Get("softlayer_username").(string),
		SoftlayerEndpointURL:   d.Get("softlayer_endpoint_url").(string),
		SoftlayerTimeout:       softlayerTimeout,
		SoftlayerAccountNumber: d.Get("softlayer_account_number").(string),
	}

	return config.Config()

}
