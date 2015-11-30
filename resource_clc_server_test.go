package terraform_clc

import (
	"fmt"
	"strings"
	"testing"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// things to test:
//   basic crud
//   modify specs
//   power operations
//   add'l disks
//   custom fields? (skip)

func TestAccServer_Basic(t *testing.T) {
	var resp server.Response
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("clc_server.acc_test_server", &resp),
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "name_template", "test"),
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "cpu", "1"),
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "memory_mb", "1024"),
				),
			},
			// update simple attrs
			resource.TestStep{
				Config: testAccCheckServerConfig_cpumem,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("clc_server.acc_test_server", &resp),
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "cpu", "2"),
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "memory_mb", "2048"),
					testAccCheckServerUpdatedSpec("clc_server.acc_test_server", &resp),
				),
			},
			// toggle power
			resource.TestStep{
				Config: testAccCheckServerConfig_power,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("clc_server.acc_test_server", &resp),
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "power_state", "stopped"),
				),
			},
			/* // currently broken since disk updates require diskId
			// add disks
			resource.TestStep{
				Config: testAccCheckServerConfig_disks,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists("clc_server.acc_test_server", &resp),
					// power still off
					resource.TestCheckResourceAttr(
						"clc_server.acc_test_server", "power_state", "stopped"),
					testAccCheckServerUpdatedDisks("clc_server.acc_test_server", &resp),
				),
			},
			*/
		},
	})
}

func testAccCheckServerDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*clc.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "clc_server" {
			continue
		}

		_, err := client.Server.Get(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Server still exists")
		}
	}

	return nil
}

func testAccCheckServerExists(n string, resp *server.Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No server ID is set")
		}

		client := testAccProvider.Meta().(*clc.Client)
		srv, err := client.Server.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if strings.ToUpper(srv.ID) != rs.Primary.ID {
			return fmt.Errorf("Server not found")
		}
		*resp = *srv
		return nil
	}
}

func testAccCheckServerUpdatedSpec(n string, resp *server.Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		client := testAccProvider.Meta().(*clc.Client)
		srv, err := client.Server.Get(rs.Primary.ID)
		if err != nil {
			return err
		}
		cpu := srv.Details.CPU
		mem := srv.Details.MemoryMB
		s_cpu := fmt.Sprintf("%v", cpu)
		s_mem := fmt.Sprintf("%v", mem)
		ex_cpu := rs.Primary.Attributes["cpu"]
		ex_mem := rs.Primary.Attributes["memory_mb"]
		if s_cpu != ex_cpu {
			return fmt.Errorf("Expected CPU to be %v but found %v", ex_cpu, s_cpu)
		}
		if s_mem != ex_mem {
			return fmt.Errorf("Expected MEM to be %v but found %v", ex_mem, s_mem)
		}
		return nil
	}
}

func testAccCheckServerUpdatedDisks(n string, resp *server.Response) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		client := testAccProvider.Meta().(*clc.Client)
		srv, err := client.Server.Get(rs.Primary.ID)
		if err != nil {
			return err
		}

		if len(srv.Details.Disks) <= 3 {
			return fmt.Errorf("Expected total of > 3 drives. found: %v", len(srv.Details.Disks))
		}

		return nil
	}
}

const testAccCheckServerConfig_basic = `
resource "clc_group" "acc_test_group" {
    location_id = "WA1"
    name = "acc_test_group"
    parent = "Default Group"
}

resource "clc_server" "acc_test_server" {
    name_template = "test"
    source_server_id = "UBUNTU-14-64-TEMPLATE"
    group_id = "${clc_group.acc_test_group.id}"
    cpu = 1
    memory_mb = 1024
    password = "Green123$"
}
`

const testAccCheckServerConfig_cpumem = `
resource "clc_group" "acc_test_group" {
    location_id = "WA1"
    name = "acc_test_group"
    parent = "Default Group"
}

resource "clc_server" "acc_test_server" {
    name_template = "test"
    source_server_id = "UBUNTU-14-64-TEMPLATE"
    group_id = "${clc_group.acc_test_group.id}"
    cpu = 2
    memory_mb = 2048
    password = "Green123$"
    power_state = "started"
}
`

const testAccCheckServerConfig_power = `
resource "clc_group" "acc_test_group" {
    location_id = "WA1"
    name = "acc_test_group"
    parent = "Default Group"
}

resource "clc_server" "acc_test_server" {
    name_template = "test"
    source_server_id = "UBUNTU-14-64-TEMPLATE"
    group_id = "${clc_group.acc_test_group.id}"
    cpu = 2
    memory_mb = 2048
    password = "Green123$"
    power_state = "stopped"
}
`

const testAccCheckServerConfig_disks = `
resource "clc_group" "acc_test_group" {
    location_id = "WA1"
    name = "acc_test_group"
    parent = "Default Group"
}

resource "clc_server" "acc_test_server" {
    name_template = "test"
    source_server_id = "UBUNTU-14-64-TEMPLATE"
    group_id = "${clc_group.acc_test_group.id}"
    cpu = 2
    memory_mb = 2048
    password = "Green123$"
    power_state = "stopped"
    additional_disks
        {
            path = "/data1"
            size_gb = 100
            type = "partitioned"
        }

}
`
