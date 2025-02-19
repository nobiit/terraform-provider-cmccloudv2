package cmccloudv2

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cmc-cloud/gocmcapiv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func datasourceServerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"server_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Id of the server",
		},
		"ip_address": {
			Type:        schema.TypeString,
			Description: "Filter by ip address of server",
			Optional:    true,
			ForceNew:    true,
		},
		"name": {
			Type:     schema.TypeString,
			Optional: true,
			ForceNew: true,
		},
		"status": {
			Type:         schema.TypeString,
			Description:  "filter by server status (case-insensitive)",
			ValidateFunc: validation.StringInSlice([]string{"active", "shutoff", "error", "suspended", "build", "reboot", "rebuild", "resize", "resized", "paused", "shelved", "rescue", "revert_resize", "verify_resize"}, true),
			Optional:     true,
			ForceNew:     true,
		},
		"vm_state": {
			Type:         schema.TypeString,
			Description:  "filter by vm_state (case-insensitive)",
			ValidateFunc: validation.StringInSlice([]string{"active", "stopped", "error", "building", "resized", "rescued", "paused", "suspended", "shelved"}, true),
			Optional:     true,
			ForceNew:     true,
		},
		"zone": {
			Type:        schema.TypeString,
			Description: "filter by server zone that contains this text (case-insensitive)",
			Optional:    true,
			ForceNew:    true,
		},
		"created_at": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
		"cpu": {
			Type:     schema.TypeInt,
			Computed: true,
			ForceNew: true,
		},
		"ram": {
			Type:     schema.TypeInt,
			Computed: true,
			ForceNew: true,
		},
		"flavor_name": {
			Type:     schema.TypeString,
			Computed: true,
			ForceNew: true,
		},
	}
}

func datasourceServer() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceServerRead,
		Schema: datasourceServerSchema(),
	}
}

func dataSourceServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()

	var allServers []gocmcapiv2.Server
	if serverId := d.Get("server_id").(string); serverId != "" {
		server, err := client.Server.Get(serverId, false)
		if err != nil {
			if errors.Is(err, gocmcapiv2.ErrNotFound) {
				d.SetId("")
				return fmt.Errorf("unable to retrieve server [%s]: %s", serverId, err)
			}
		}
		allServers = append(allServers, server)
	} else {
		params := map[string]string{
			"name":     d.Get("name").(string),
			"status":   d.Get("status").(string),
			"vm_state": d.Get("vm_state").(string),
			"zone":     d.Get("zone").(string),
		}
		servers, err := client.Server.List(params)
		if err != nil {
			return fmt.Errorf("error when get servers %v", err)
		}
		allServers = append(allServers, servers...)
	}
	if len(allServers) > 0 {
		var filteredServers []gocmcapiv2.Server
		for _, server := range allServers {
			if v := d.Get("name").(string); v != "" {
				if strings.ToLower(server.Name) != strings.ToLower(v) {
					continue
				}
			}
			if v := d.Get("status").(string); v != "" {
				if !strings.Contains(strings.ToLower(server.Status), strings.ToLower(v)) {
					continue
				}
			}
			if v := d.Get("vm_state").(string); v != "" {
				if !strings.Contains(strings.ToLower(server.VMState), strings.ToLower(v)) {
					continue
				}
			}
			if v := d.Get("ip_address").(string); v != "" {
				switch server.Addresses.(type) {
				case []interface{}:
				case map[string]interface{}:
					foundIp := false
					if m, ok := server.Addresses.(map[string]interface{}); ok {
						// Duyệt qua map
						for _, value := range m {
							for _, inter := range value.([]interface{}) {
								intermap := inter.(map[string]interface{})
								ip := intermap["addr"].(string)
								if ip == v {
									foundIp = true
								}
							}
						}
					}
					if !foundIp {
						continue
					}
				default:
					//
				}
			}
			filteredServers = append(filteredServers, server)
		}
		allServers = filteredServers
	}
	if len(allServers) < 1 {
		return fmt.Errorf("your query returned no results. Please change your search criteria and try again")
	}

	if len(allServers) > 1 {
		gocmcapiv2.Logo("[DEBUG] Multiple results found: %#v", allServers)
		return fmt.Errorf("your query returned more than one result. Please try a more specific search criteria")
	}

	return dataSourceComputeServerAttributes(d, allServers[0])
}

func dataSourceComputeServerAttributes(d *schema.ResourceData, server gocmcapiv2.Server) error {
	log.Printf("[DEBUG] Retrieved server %s: %#v", server.ID, server)
	d.SetId(server.ID)
	return errors.Join(
		d.Set("name", server.Name),
		d.Set("status", server.Status),
		d.Set("vm_state", strings.ToLower(server.VMState)),
		d.Set("status", strings.ToLower(server.Status)),
		//d.Set("description", server.Description),
		d.Set("cpu", server.Flavor.CPU),
		d.Set("ram", server.Flavor.RAM/1024),
		d.Set("flavor_name", server.Flavor.OriginalName),
		d.Set("created_at", server.Created),
	)
}
