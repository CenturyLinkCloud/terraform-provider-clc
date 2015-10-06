package main

import (
	"github.com/CenturyLinkCloud/terraform-provider-clc"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: terraform_clc.Provider,
	})
}
