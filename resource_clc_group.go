package terraform_clc

import (
	"encoding/json"
	"fmt"

	"github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/group"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCLCGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceCLCGroupCreate,
		Read:   resourceCLCGroupRead,
		Update: resourceCLCGroupUpdate,
		Delete: resourceCLCGroupDelete,
		Schema: map[string]*schema.Schema{
			"location_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"parent": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"custom_fields": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
		},
	}
}

func resourceCLCGroupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	dc := d.Get("location_id").(string)
	m, err := dcGroups(dc, meta)
	if err != nil {
	}
	name := d.Get("name").(string)
	// use an existing group if we have one
	if m[name] != "" {
		LOG.Println("Using EXISTING group: %v => %v", name, m[name])
		d.SetId(m[name])
		return nil
	}
	// otherwise, we're creating one. we'll need a parent
	p := d.Get("parent").(string)
	parent := m[p]

	if parent == "" {
		return fmt.Errorf("Failed resolving parent group %s - %s", p, m)
	}
	LOG.Println(m)
	LOG.Println(parent)

	d.Set("parent_group_id", parent)
	spec := group.Group{
		Name:          name,
		Description:   d.Get("description").(string),
		ParentGroupID: d.Get("parent_group_id").(string),
	}
	resp, err := client.Group.Create(spec)
	if err != nil {
		return fmt.Errorf("Failed creating group: %s", err)
	}
	LOG.Println("Group created")
	b, _ := json.Marshal(resp)
	LOG.Println(string(b))
	d.SetId(resp.ID)
	return nil
}

func resourceCLCGroupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	id := d.Get("id").(string)
	g, err := client.Group.Get(id)
	if err != nil {
		return fmt.Errorf("Failed to find the specified group with id: %s -  %s", id, err)
	}
	d.Set("name", g.Name)
	d.Set("description", g.Description)
	// need to traverse links?
	//d.Set("parent_group_id", g.ParentGroupID)
	return nil
}

func resourceCLCGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	// unimplemented
	return nil
}

func resourceCLCGroupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	id := d.Id()
	LOG.Printf("Deleting group %v", id)
	client.Group.Delete(id)
	return nil
}
