package terraform_clc

import (
	"fmt"
	"testing"

	"github.com/CenturyLinkCloud/clc-sdk/server"
)

// things to test:
//   maps to internal specified ip
//   port range
//   update existing rule
//   CIDR restriction

func TestAccPublicIP_Basic(t *testing.T) {
	var resp server.PublicIP
	fmt.Println(resp)
}
