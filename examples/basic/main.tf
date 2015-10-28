provider "clc" {
  # THESE fields pulled from env vars if not provided
  # username = "login" 
  # password = "passwd" 
  # account = "AF"
}


resource "clc_group" "web" {
  location_id = "WA1"
  name = "TERRA"
  parent = "Default Group"
}

resource "clc_server" "srv" {
  name_template = "trusty"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  group_id = "${clc_group.web.id}"
  cpu = 2
  memory_mb = 2048
  password = "Green123$"
}