package terraform_clc

import (
	"encoding/json"
	"fmt"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/lb"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCLCLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Create: resourceCLCLoadBalancerCreate,
		Read:   resourceCLCLoadBalancerRead,
		Update: resourceCLCLoadBalancerUpdate,
		Delete: resourceCLCLoadBalancerDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"data_center": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"status": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			// optional
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			// computed
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"pools": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeMap},
			},
		},
	}
}

func resourceCLCLoadBalancerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	dc := d.Get("data_center").(string)
	name := d.Get("name").(string)
	desc := d.Get("description").(string)
	status := d.Get("status").(string)
	r1 := lb.LoadBalancer{
		Name:        name,
		Description: desc,
		Status:      status,
	}
	l, err := client.LB.Create(dc, r1)
	if err != nil {
		return fmt.Errorf("Failed creating load balancer under %v/%v: %v", dc, name, err)
	}
	d.SetId(l.ID)

	a1, _ := json.Marshal(l)
	LOG.Println(string(a1))
	return resourceCLCLoadBalancerRead(d, meta)
}

func resourceCLCLoadBalancerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	dc := d.Get("data_center").(string)
	lbname := d.Id()
	resp, err := client.LB.Get(dc, lbname)
	if err != nil {
		LOG.Printf("Failed finding load balancer %v/%v. Marking destroyed", dc, lbname)
		d.SetId("")
		return nil
	}
	d.Set("description", resp.Description)
	d.Set("ip_address", resp.IPaddress)
	d.Set("status", resp.Status)
	d.Set("pools", resp.Pools)
	d.Set("links", resp.Links)
	return nil
}

func resourceCLCLoadBalancerUpdate(d *schema.ResourceData, meta interface{}) error {
	//client := meta.(*clc.Client)
	return nil
}

func resourceCLCLoadBalancerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clc.Client)
	dc := d.Get("data_center").(string)
	id := d.Id()
	err := client.LB.Delete(dc, id)
	if err != nil {
		return fmt.Errorf("Failed deleting loadbalancer %v: %v", id, err)
	}
	return nil
}