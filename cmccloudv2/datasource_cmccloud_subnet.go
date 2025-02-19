package cmccloudv2

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cmc-cloud/gocmcapiv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceSubnetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"subnet_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Id of the subnet",
		},
		"vpc_id": {
			Type:        schema.TypeString,
			Description: "Filter by vpc id",
			Optional:    true,
			ForceNew:    true,
		},
		"name": {
			Type:        schema.TypeString,
			Description: "Filter by ip address of subnet",
			Optional:    true,
			ForceNew:    true,
		},
		"cidr": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"gateway_ip": {
			Type:        schema.TypeString,
			Description: "Filter by gateway_ip",
			Optional:    true,
			ForceNew:    true,
		},
		"created_at": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
	}
}

func datasourceSubnet() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceSubnetRead,
		Schema: datasourceSubnetSchema(),
	}
}

func dataSourceSubnetRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()

	var allSubnets []gocmcapiv2.Subnet
	if subnetId := d.Get("subnet_id").(string); subnetId != "" {
		subnet, err := client.Subnet.Get(subnetId)
		if err != nil {
			if errors.Is(err, gocmcapiv2.ErrNotFound) {
				d.SetId("")
				return fmt.Errorf("unable to retrieve subnet [%s]: %s", subnetId, err)
			}
		}
		allSubnets = append(allSubnets, subnet)
	} else {
		params := map[string]string{
			"vpc_id": d.Get("vpc_id").(string),
		}
		subnets, err := client.Subnet.List(params)
		if err != nil {
			return fmt.Errorf("error when get subnets %v", err)
		}
		allSubnets = append(allSubnets, subnets...)
	}
	if len(allSubnets) > 0 {
		var filteredSubnets []gocmcapiv2.Subnet
		for _, subnet := range allSubnets {
			if v := d.Get("cidr").(string); v != "" {
				if v != subnet.Cidr {
					continue
				}
			}
			if v := d.Get("gateway_ip").(string); v != "" {
				if subnet.GatewayIP != v {
					continue
				}
			}
			if v := d.Get("name").(string); v != "" {
				if strings.ToLower(subnet.Name) != strings.ToLower(v) {
					continue
				}
			}
			filteredSubnets = append(filteredSubnets, subnet)
		}
		allSubnets = filteredSubnets
	}
	if len(allSubnets) < 1 {
		return fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}

	if len(allSubnets) > 1 {
		gocmcapiv2.Logo("[DEBUG] Multiple results found: %#v", allSubnets)
		return fmt.Errorf("your query returned more than one result. Please try a more specific search criteria")
	}

	return dataSourceComputeSubnetAttributes(d, allSubnets[0])
}

func dataSourceComputeSubnetAttributes(d *schema.ResourceData, subnet gocmcapiv2.Subnet) error {
	log.Printf("[DEBUG] Retrieved subnet %s: %#v", subnet.ID, subnet)
	d.SetId(subnet.ID)
	return errors.Join(
		d.Set("name", subnet.Name),
		d.Set("cidr", subnet.Cidr),
		d.Set("gateway_ip", subnet.GatewayIP),
		d.Set("vpc_id", subnet.VpcID),
		d.Set("created_at", subnet.CreatedAt),
	)
}
