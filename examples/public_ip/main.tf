resource "clc_group" "web" {
  location_id = "WA1"
  name = "web"
  parent = "Default Group"
}

resource "clc_server" "bastion" {
  name_template = "BASTION"
  description = "bastion host for web nodes"
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
  #source_restrictions
  #   { cidr = "108.19.67.15/32" }
  ports
    {
      protocol = "TCP"
      port = 22
    }
}


resource "clc_server" "web01" {
  name_template = "NGX"
  description = "web server"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  type = "standard"
  group_id = "${clc_group.web.id}"
  cpu = 4
  memory_mb = 2048
  password = "Green123456$"
  power_state = "started"

  depends_on = "clc_public_ip.mgmt"

  provisioner "remote-exec" {
    inline = [
      "apt-get update",
      "apt-get install -qy nginx"
    ]
    connection {
      host = "${clc_server.web01.private_ip_address}"
      user = "root"
      password = "${clc_server.web01.password}"
      bastion_host = "${clc_public_ip.mgmt.id}"
      bastion_user = "root"
      bastion_password = "${clc_server.bastion.password}"
    }
  }
}

resource "clc_public_ip" "frontdoor" {
  server_id = "${clc_server.web01.id}"
  ports
    {
      protocol = "TCP"
      port = 80
    }
  ports
    {
      protocol = "TCP"
      port = 443
    }
}

output "ip_bastion" {
  value = "${clc_public_ip.mgmt.id}"
}
output "ip_web01_internal" {
  value = "${clc_server.web01.private_ip_address}"
}
output "ip_web01_public" {
  value = "${clc_public_ip.frontdoor.id}"
}
