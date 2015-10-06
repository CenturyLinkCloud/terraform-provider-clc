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

var LOG log.Logger

func Provider() terraform.ResourceProvider {
	logFile, _ := os.OpenFile("/tmp/plugin.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	LOG = *(log.New(logFile, "[CLC] ", log.Lshortfile))
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
				DefaultFunc: schema.EnvDefaultFunc("CLC_ALIAS", nil),
				Description: "Your CLC account alias",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"clc_server": resourceCLCServer(),
			"clc_group":  resourceCLCGroup(),
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
	alerts, err := client.Alert.GetAll()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to the CLC api because %s", err)
	}
	for _, a := range alerts.Items {
		fmt.Println(a)
	}
	return client, nil
}
