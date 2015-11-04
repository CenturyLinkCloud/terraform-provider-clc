package terraform_clc

import (
	"fmt"
	"testing"

	"github.com/CenturyLinkCloud/clc-sdk/server"
)

// things to test:
//   updates name/desc
//   toggles status
//   created w/o pool
//   created w/ pool
//   works for 80 and 443 together

func TestAccLoadBalancer_Basic(t *testing.T) {
	var resp server.PublicIP
	fmt.Println(resp)
}
