# terraform-provider-clc

Terraform provider for CenturyLinkCloud.


## Installation

1. Download the plugin from the [releases tab][3]
2. Put it somewhere were it can permanently live, it doesn't need to be in your path.
3. Create or modify your `~/.terraformrc` file. You'll need at least this:

```
providers {
    clc = "terraform-provider-clc"
}
```

If you didn't add terraform-provider-clc to your path, you'll need to put the full path to the location of the plugin.

[3]:https://github.com/CenturyLinkCloud/terraform-provider-clc/releases

## Usage

### Provider Configuration

#### `clc`

```
provider "clc" {
  username = ""
  password = ""
  account = ""
}
```

The provider options will be taken from ENV vars if not specified.

* `username` & `password` - credentials used to log in on https://control.ctl.io
* `account` - Account alias

### Resource Configuration

#### `clc_group`

Creates a new group or resolves to an existing group.

```
resource "clc_group" "group01" {
  location_id = "CA1"
  name = "cluster01"
  parent = "somegroup"
}
```

### `clc_server`

Creates a new server instance.

```
resource "clc_server" "web01" {
    name_template = "SRV" # eventual name may be mutated by the platform
    source_server_id = "UBUNTU-14-64-TEMPLATE"
    group_id = "${clc_group.NAME.id}"
    cpu = 2
    memory_mb = 2048
    password = "supersecure"
}
```



value                             | Type     | Forces New | Value Type | Description
--------------------------------- | -------- | ---------- | ---------- | -----------
`name_tempate`                    | Required | no         | string     | Name of the server. Will be permuted by platform
`source_server_id`                | Required | yes        | string     | VM image to use
`group_id`                        | Required | no         | string     | ID of the group this server will belong to
`type`                            | Required | no         | string     | Type of build: { standard, hyperscale, bareMetal }
`password`                        | Required | no         | string     | Root password, unsalted
`description`                     | Optional | no         | string     | Description of server
`private_ip_address`              | Optional | no         | string     | Generated if not provided
`network_id`                      | Optional | no         | string     | ID of the network to place server in


full API options documented: [https://www.ctl.io/api-docs/v2/#servers-create-server]


#### `clc_public_ip`

Creates a new public IP and attaches to existing server instance

```

resource "clc_public_ip" "mgmt" {
  server_id = "${clc_server.<SOME_NAMED_RESOURCE>.id}"
  internal_ip_address = "${clc_server.<SOME_NAMED_RESOURCE>.private_ip_address}"
  source_restrictions
     { cidr = "108.19.67.15/32" }
  ports
    {
      protocol = "TCP"
      port = 22
    }
  ports
    {
      protocol = "TCP"
      port = 80
    }
}
```
