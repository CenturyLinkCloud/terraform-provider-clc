package terraform_clc

import (
	"fmt"
	"log"
	"os"

	"github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/api"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// someplace to dump plugin logs
var LOG log.Logger

func Provider() terraform.ResourceProvider {
	fout := os.Stdout
	if os.Getenv("DEBUG") != "" {
		fout, _ = os.OpenFile("plugin.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	}
	LOG = *(log.New(fout, "[CLC] ", log.Lshortfile))
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLC_USERNAME", nil),
				Description: "Your CLC username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLC_PASSWORD", nil),
				Description: "Your CLC password",
			},
			"account": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLC_ACCOUNT", nil),
				Description: "Your CLC account alias",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"clc_server":             resourceCLCServer(),
			"clc_group":              resourceCLCGroup(),
			"clc_public_ip":          resourceCLCPublicIP(),
			"clc_load_balancer":      resourceCLCLoadBalancer(),
			"clc_load_balancer_pool": resourceCLCLoadBalancerPool(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	un := d.Get("username").(string)
	pw := d.Get("password").(string)
	ac := d.Get("account").(string)
	config := api.NewConfig(un, pw, ac)
	client := clc.New(config)
	err := client.Authenticate()
	if err != nil {
		return nil, fmt.Errorf("Failed authenticated with provided credentials: %v", err)
	}

	alerts, err := client.Alert.GetAll()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the CLC api because %s", err)
	}
	for _, a := range alerts.Items {
		LOG.Printf("Received alert: %v", a)
	}
	LOG.Printf("account: %v %v", ac, un)
	return client, nil
}
