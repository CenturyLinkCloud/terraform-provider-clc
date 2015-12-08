resource "clc_group" "web" {
  location_id = "CA1"
  name = "terraform-public-ip"
  parent = "Default Group"
}

resource "clc_server" "generic" {
  name_template = "generic"
  description = "generic host"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  type = "standard"
  group_id = "${clc_group.web.id}"
  cpu = 1
  memory_mb = 1024
  password = "Green987$"
  power_state = "started"

  #depends_on = "clc_public_ip.public"
  
}


resource "clc_server" "slave01" {
  name_template = "slave"
  description = "mesos-slave"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  type = "standard"
  group_id = "${clc_group.web.id}"
  cpu = 4
  memory_mb = 4096
  password = "Green987$"
  power_state = "started"

  #depends_on = "clc_public_ip.public"

  provisioner "file" {
    source = "~/.ssh/id_rsa.pub"
    destination = "/root/.ssh/authorized_keys"
    connection {
      host = "${clc_public_ip.public.id}"
      keyfile = "~/.ssh/id_rsa"
      bastion_host = "${clc_public_ip.public.id}"
      bastion_user = "root"
      bastion_keyfile = "~/.ssh/id_rsa"
    }
  }


  
  
}

resource "clc_public_ip" "public" {
  server_id = "${clc_server.generic.id}"
  internal_ip_address = "${clc_server.generic.private_ip_address}"
  #source_restrictions
  #   { cidr = "108.19.67.15/32" }
  ports
    {
      protocol = "TCP"
      port = 22
    }
  ports
    {
      protocol = "TCP"
      port = 2000
      port_to = 9000
    }

  provisioner "file" {
    source = "~/.ssh/id_rsa.pub"
    destination = "/root/.ssh/authorized_keys"
    connection {
      host = "${clc_public_ip.public.id}"
      password = "${clc_server.generic.password}"
    }
  }

}


output "ip_public" {
  value = "${clc_public_ip.public.id}"
}
output "ip_private" {
  value = "${clc_server.generic.private_ip_address}"
}
