package terraform_clc

import (
	"encoding/base64"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/group"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/CenturyLinkCloud/clc-sdk/status"
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
