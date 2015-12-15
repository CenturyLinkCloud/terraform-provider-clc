package clc

import (
	"encoding/base64"
	"fmt"
	"strconv"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/api"
	"github.com/CenturyLinkCloud/clc-sdk/group"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/CenturyLinkCloud/clc-sdk/status"
	"github.com/hashicorp/terraform/helper/schema"
	"golang.org/x/crypto/sha3"
)

func passwordState(val interface{}) string {
	return hashedString(val.(string))
}

func hashedString(key string) string {
	hash := sha3.Sum256([]byte(key))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func waitStatus(client *clc.Client, id string) error {
	// block until queue is processed and server is up
	poll := make(chan *status.Response, 1)
	err := client.Status.Poll(id, poll)
	if err != nil {
		return nil
	}
	status := <-poll
	LOG.Printf("status %v", status)
	if status.Failed() {
		return fmt.Errorf("unsuccessful job %v failed with status: %v", id, status.Status)
	}
	return nil
}

func dcGroups(dcname string, meta interface{}) (map[string]string, error) {
	client := meta.(*clc.Client)
	dc, _ := client.DC.Get(dcname)
	_, id := dc.Links.GetID("group")
	m := map[string]string{}
	resp, _ := client.Group.Get(id)
	m[resp.Name] = resp.ID // top
	for _, x := range resp.Groups {
		deepGroups(x, &m)
	}
	return m, nil
}

func deepGroups(g group.Groups, m *map[string]string) {
	(*m)[g.Name] = g.ID
	for _, sg := range g.Groups {
		deepGroups(sg, m)
	}
}

func stateFromString(st string) server.PowerState {
	switch st {
	case "on", "started":
		return server.On
	case "off", "stopped":
		return server.Off
	case "pause", "paused":
		return server.Pause
	case "reboot":
		return server.Reboot
	case "reset":
		return server.Reset
	case "shutdown":
		return server.ShutDown
	case "start_maintenance":
		return server.StartMaintenance
	case "stop_maintenance":
		return server.StopMaintenance
	}
	return -1
}

func parseCustomfields(d *schema.ResourceData) ([]api.Customfields, error) {
	var fields []api.Customfields
	if v := d.Get("custom_fields"); v != nil {
		for _, v := range v.([]interface{}) {
			m := v.(map[string]interface{})
			f := api.Customfields{
				ID:    m["id"].(string),
				Value: m["value"].(string),
			}
			fields = append(fields, f)
		}
	}
	return fields, nil
}

func parseAdditionalDisks(d *schema.ResourceData) ([]server.Disk, error) {
	// some complexity here: create has a different format than update
	// on-create: { path, sizeGB, type }
	// on-update: { diskId, sizeGB, (path), (type=partitioned) }
	var disks []server.Disk
	if v := d.Get("additional_disks"); v != nil {
		for _, v := range v.([]interface{}) {
			m := v.(map[string]interface{})
			ty := m["type"].(string)
			var pa string
			if nil != m["path"] {
				pa = m["path"].(string)
			}
			sz, err := strconv.Atoi(m["size_gb"].(string))
			if err != nil {
				LOG.Printf("Failed parsing size '%v'. skipping", m["size_gb"])
				return nil, fmt.Errorf("Unable to parse %v as int", m["size_gb"])
			}
			if ty != "raw" && ty != "partitioned" {
				return nil, fmt.Errorf("Expected type of { raw | partitioned }. received %v", ty)
			}
			if ty == "raw" && pa != "" {
				return nil, fmt.Errorf("Path can not be specified for raw disks")
			}
			disk := server.Disk{
				SizeGB: sz,
				Type:   ty,
			}
			if pa != "" {
				disk.Path = pa
			}
			disks = append(disks, disk)
		}
	}
	return disks, nil
}
