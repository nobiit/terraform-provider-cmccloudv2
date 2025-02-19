package cmccloudv2

import (
	"fmt"
	"time"

	// "strconv"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAutoScalingAZPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAutoScalingAZPolicyCreate,
		Read:   resourceAutoScalingAZPolicyRead,
		Update: resourceAutoScalingAZPolicyUpdate,
		Delete: resourceAutoScalingAZPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAutoScalingAZPolicyImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        autoscalingAZPolicySchema(),
	}
}
func resourceAutoScalingAZPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	params := map[string]interface{}{
		"name":  d.Get("name").(string),
		"zones": getStringArrayFromTypeSet(d.Get("zones").(*schema.Set)),
	}
	res, err := client.AutoScalingPolicy.CreateAZPolicy(params)

	if err != nil {
		return fmt.Errorf("error creating az policy: %v", err.Error())
	}
	d.SetId(res.ID)
	return resourceAutoScalingAZPolicyRead(d, meta)
}

func resourceAutoScalingAZPolicyRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	policy, err := client.AutoScalingPolicy.Get(d.Id())
	if err != nil {
		return fmt.Errorf("error retrieving az policy %s: %v", d.Id(), err)
	}

	zoneNames := make([]string, len(policy.Spec.Properties.Zones))
	for zoneIndex, zone := range policy.Spec.Properties.Zones {
		zoneNames[zoneIndex] = zone.Name
	}

	_ = d.Set("name", policy.Name)
	_ = d.Set("zones", zoneNames)
	return nil
}

func resourceAutoScalingAZPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	if d.HasChange("name") {
		params := map[string]interface{}{
			"name": d.Get("name").(string),
		}
		_, err := client.AutoScalingPolicy.UpdateAZPolicy(d.Id(), params) //Update(id, map[string]interface{}{"name": d.Get("name").(string)})
		if err != nil {
			return fmt.Errorf("error when update az policy [%s]: %v", d.Id(), err)
		}
	}
	return resourceAutoScalingAZPolicyRead(d, meta)
}

func resourceAutoScalingAZPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	_, err := client.AutoScalingPolicy.Delete(d.Id())

	if err != nil {
		return fmt.Errorf("error delete az policy: %v", err)
	}
	_, err = waitUntilAutoScalingPolicyDeleted(d, meta)
	if err != nil {
		return fmt.Errorf("error delete az policy: %v", err)
	}
	return nil
}

func resourceAutoScalingAZPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceAutoScalingAZPolicyRead(d, meta)
	return []*schema.ResourceData{d}, err
}
