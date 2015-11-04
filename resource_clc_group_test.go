package terraform_clc

import (
	"fmt"
	"testing"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/group"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// things to test:
//   resolves to existing group
//   does not nuke a group w/ no parents (root group)
//   change a name on a group

func TestAccGroup_Basic(t *testing.T) {
	var resp group.Response
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists("clc_group.foobar", &resp),
					testAccCheckGroupParent(&resp),
					resource.TestCheckResourceAttr(
						"clc_group.foobar", "name", "foobar"),
					resource.TestCheckResourceAttr(
						"clc_group.foobar", "location_id", "WA1"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clc.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "clc_group" {
			continue
		}

		_, err := client.Group.Get(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Group still exists")
		}
	}

	return nil
}

func testAccCheckGroupParent(resp *group.Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*clc.Client)
		ok, l := resp.Links.GetLink("parentGroup")
		if !ok {
			return fmt.Errorf("Missing parent group: %v", resp)
		}
		resp, err := client.Group.Get(l.ID)
		if err != nil {
			return fmt.Errorf("Failed fetching parent %v: %v", l.ID, err)
		}
		if resp.Name != "Default Group" {
			return fmt.Errorf("Incorrect parent %v: %v", l, err)
		}
		// would be good to test parent but we'd have to make a bunch of calls
		return nil
	}
}

func testAccCheckGroupExists(n string, resp *group.Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Group ID is set")
		}

		client := testAccProvider.Meta().(*clc.Client)
		g, err := client.Group.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if g.ID != rs.Primary.ID {
			return fmt.Errorf("Group not found")
		}
		*resp = *g
		return nil
	}
}

const testAccCheckGroupConfig_basic = `
resource "clc_group" "foobar" {
    location_id = "WA1"
    name = "foobar"
    parent = "Default Group"
}`
