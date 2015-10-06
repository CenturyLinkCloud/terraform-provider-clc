package terraform_clc

import (
	"encoding/json"
	"fmt"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/api"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/CenturyLinkCloud/clc-sdk/status"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCLCServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceCLCServerCreate,
		Read:   resourceCLCServerRead,
		Update: resourceCLCServerUpdate,
		Delete: resourceCLCServerDelete,
		Schema: map[string]*schema.Schema{
			"name_template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"source_server_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"memory_mb": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			// optional
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "standard",
			},
			"network_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"private_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Default:  nil,
			},
			// computed
			"public_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"password": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				StateFunc: passwordState,
			},

			"power_state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_date": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			/*
				"ttl": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"custom_fields": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
				},
				"additional_disks": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
				},
				"packages": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
				},
				"metal_configuration_id": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
			*/
		},
	}
}

func resourceCLCServerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	spec := server.Server{
		Name:           d.Get("name_template").(string),
		Password:       d.Get("password").(string),
		Description:    d.Get("description").(string),
		GroupID:        d.Get("group_id").(string),
		CPU:            d.Get("cpu").(int),
		MemoryGB:       d.Get("memory_mb").(int) / 1024,
		SourceServerID: d.Get("source_server_id").(string),
		Type:           d.Get("type").(string),
		IPaddress:      d.Get("private_ip_address").(string),
	}
	resp, err := client.Server.Create(spec)
	if err != nil || !resp.IsQueued {
		return fmt.Errorf("Failed creating server: %v", err)
	}
	// server's UUID returned under rel=self link
	_, uuid := resp.Links.GetID("self")

	js, _ := json.MarshalIndent(resp, "", "  ")
	LOG.Println(string(js))

	ok, st := resp.GetStatusID()
	if !ok {
		return fmt.Errorf("Failed extracting status to poll on %v: %v", resp, err)
	}
	waitStatus(client, st)

	s, err := client.Server.Get(uuid)
	d.SetId(s.Name)
	LOG.Printf("Server created. id: %v", s.Name)
	return resourceCLCServerRead(d, meta)
}

func resourceCLCServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	s, err := client.Server.Get(d.Id())
	if err != nil {
		//return fmt.Errorf("Failed fetching server: %v - %v", d.Id(), err)
		LOG.Println("Failed finding server: %v. Marking destroyed", d.Id())
		d.SetId("")
		return nil
	}
	js, _ := json.Marshal(s.Details)
	LOG.Println(string(js))

	if len(s.Details.IPaddresses) > 0 {
		d.Set("private_ip_address", s.Details.IPaddresses[0].Internal)
		d.Set("public_ip_address", s.Details.IPaddresses[0].Public)
	}

	d.Set("name", s.Name)
	d.Set("groupId", s.GroupID)
	d.Set("status", s.Status)
	d.Set("power_state", s.Details.Powerstate)
	d.Set("cpu", s.Details.CPU)
	d.Set("memory_mb", s.Details.MemoryMB)
	d.Set("disk_gb", s.Details.Storagegb)
	d.Set("status", s.Status)
	d.Set("created_date", s.ChangeInfo.CreatedDate)
	d.Set("modified_date", s.ChangeInfo.ModifiedDate)
	return nil
}

func resourceCLCServerUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	id := d.Id()

	var err error
	var edits []api.Update = make([]api.Update, 0)
	var updates []api.Update = make([]api.Update, 0)
	var i int
	d.Partial(true)
	s, err := client.Server.Get(id)
	if err != nil {
		return fmt.Errorf("Failed fetching server: %v - %v", d.Id(), err)
	}
	// edits happen synchronously
	if delta, orig := d.Get("description").(string), s.Description; delta != orig {
		d.SetPartial("description")
		edits = append(edits, server.UpdateDescription(delta))
	}
	if delta, orig := d.Get("group_id").(string), s.GroupID; delta != orig {
		d.SetPartial("group_id")
		edits = append(edits, server.UpdateGroup(delta))
	}
	if len(edits) > 0 {
		err = client.Server.Edit(id, edits...)
		if err != nil {
			return fmt.Errorf("Failed saving edits: %v", err)
		}
	}
	// updates are queue processed
	if i = d.Get("cpu").(int); i != s.Details.CPU {
		d.SetPartial("cpu")
		updates = append(updates, server.UpdateCPU(i))
	}
	if i = d.Get("memory_mb").(int); i != s.Details.MemoryMB {
		d.SetPartial("memory_mb")
		updates = append(updates, server.UpdateMemory(i/1024)) // takes GB
	}
	js, _ := json.Marshal(updates)
	LOG.Printf("updates: %v", string(js))
	if len(updates) > 0 {
		resp, err := client.Server.Update(id, updates...)
		if err != nil {
			return fmt.Errorf("Failed saving updates: %v", err)
		}

		poll := make(chan *status.Response, 1)
		client.Status.Poll(resp.ID, poll)
		status := <-poll
		LOG.Printf("Server updated! status: %v", status)
	}
	d.Partial(false)
	return nil
}

func resourceCLCServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	id := d.Id()
	resp, err := client.Server.Delete(id)
	if err != nil || !resp.IsQueued {
		return fmt.Errorf("Failed queueing delete of %v - %v", id, err)
	}

	ok, st := resp.GetStatusID()
	if !ok {
		return fmt.Errorf("Failed extracting status to poll on %v: %v", resp, err)
	}
	waitStatus(client, st)
	fmt.Printf("Server sucessfully deleted: %v", st)
	return nil
}
