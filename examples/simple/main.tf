provider "clc" {
  username = "achoi"
  account = "AF"
}


resource "clc_group" "grp01" {
  location_id = "CA1"
  name = "ernesto"
  parent = "CHOI"
}

resource "clc_server" "srv01" {
  name = "UB01"
  description = "terraformed!"
  source_server_id = "UBUNTU-14-64-TEMPLATE"
  group_id = "${clc_group.grp01.id}"
  cpu = 2
  memory_mb = 2048
  password = "Green123$"
}