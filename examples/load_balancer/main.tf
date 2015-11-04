resource "clc_group" "web" {
  location_id = "CA1"
  name = "web"
  parent = "CHOI"
}

resource "clc_server" "bastion" {
  name_template = "bastion"
  description = "bastion"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  type = "standard"
  group_id = "${clc_group.web.id}"
  cpu = 1
  memory_mb = 1024
  password = "Green123$"
  power_state = "started"
}

resource "clc_public_ip" "mgmt" {
  server_id = "${clc_server.bastion.id}"
  internal_ip_address = "${clc_server.bastion.private_ip_address}"
  ports
    {
      protocol = "TCP"
      port = 22
    }
}

resource "clc_server" "node01" {
  name_template = "node"
  description = "load balanced node"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  type = "standard"
  group_id = "${clc_group.web.id}"
  cpu = 1
  memory_mb = 1024
  password = "Green123$"
  power_state = "started"

  depends_on = "clc_public_ip.mgmt"

  provisioner "remote-exec" {
    inline = [
      "apt-get update && apt-get install -qy nginx"
    ]
    connection {
      host = "${clc_server.node01.private_ip_address}"
      password = "${clc_server.node01.password}"
      bastion_host = "${clc_public_ip.mgmt.id}"
      bastion_password = "${clc_server.bastion.password}"
    }
  }
}

resource "clc_server" "node02" {
  name_template = "node"
  description = "load balanced node"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  type = "standard"
  group_id = "${clc_group.web.id}"
  cpu = 1
  memory_mb = 1024
  password = "Green123$"
  power_state = "started"

  depends_on = "clc_public_ip.mgmt"

  provisioner "remote-exec" {
    inline = [
      "apt-get update && apt-get install -qy nginx"
    ]
    connection {
      host = "${clc_server.node02.private_ip_address}"
      password = "${clc_server.node02.password}"
      bastion_host = "${clc_public_ip.mgmt.id}"
      bastion_password = "${clc_server.bastion.password}"
    }
  }
}
  
resource "clc_load_balancer" "lb" {
  data_center = "${clc_group.web.location_id}"
  name = "lb-terraformed"
  status = "enabled"
  description = "terraform generated load balancer"
}

resource "clc_load_balancer_pool" "pool" {
  port = 80
  data_center = "${clc_group.web.location_id}"
  load_balancer = "${clc_load_balancer.lb.id}"

  #depends_on = ["clc_load_balancer.lb", "clc_server.node01", "clc_server.node02"]
  depends_on = ["clc_server.node01", "clc_server.node02"]

  nodes
    {
      status = "enabled"
      ipAddress = "${clc_server.node01.private_ip_address}"
      privatePort = 80
    }
  nodes
    {
      status = "disabled"
      ipAddress = "${clc_server.node02.private_ip_address}"
      privatePort = 80
    }  
}

