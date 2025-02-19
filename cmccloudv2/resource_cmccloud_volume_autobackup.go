package cmccloudv2

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceVolumeAutoBackup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVolumeAutoBackupCreate,
		Read:   resourceVolumeAutoBackupRead,
		Update: resourceVolumeAutoBackupUpdate,
		Delete: resourceVolumeAutoBackupDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVolumeAutoBackupImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(1 * time.Minute),
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        volumeAutoBackupSchema(),
	}
}

func resourceVolumeAutoBackupCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	scheduleTime := d.Get("schedule_time").(string)
	parts := strings.Split(scheduleTime, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid schedule time [%s], correct format is HH:mm (24-h format), eg: 19:05", scheduleTime)
	}
	hour := parts[0]
	minute := parts[1]

	vol, err := client.VolumeAutoBackup.Create(map[string]interface{}{
		"name":        d.Get("name").(string),
		"volume_id":   d.Get("volume_id").(string),
		"hour":        hour,
		"minute":      minute,
		"interval":    d.Get("interval").(int),
		"max_keep":    d.Get("max_keep").(int),
		"incremental": d.Get("incremental").(bool),
	})
	if err != nil {
		return fmt.Errorf("error creating Volume AutoBackup: %s", err)
	}
	d.SetId(vol.ID)

	return resourceVolumeAutoBackupRead(d, meta)
}

func resourceVolumeAutoBackupRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	autobackup, err := client.VolumeAutoBackup.Get(d.Id())
	if err != nil {
		return fmt.Errorf("error retrieving Volume AutoBackup %s: %v", d.Id(), err)
	}

	_ = d.Set("name", autobackup.Name)
	_ = d.Set("schedule_time", autobackup.Time)
	_ = d.Set("interval", autobackup.Interval)
	_ = d.Set("max_keep", autobackup.MaxKeep)
	_ = d.Set("incremental", !autobackup.IsFullBackup)
	_ = d.Set("created_at", autobackup.Created)
	_ = d.Set("last_run", autobackup.LastRun)
	_ = d.Set("volume_size", autobackup.VolumeSize)
	return nil
}

func resourceVolumeAutoBackupUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	id := d.Id()
	scheduleTime := d.Get("schedule_time").(string)
	parts := strings.Split(scheduleTime, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid schedule time [%s], correct format is HH:mm (24-h format), eg: 19:05", scheduleTime)
	}
	hour := parts[0]
	minute := parts[1]

	if d.HasChange("name") || d.HasChange("schedule_time") || d.HasChange("interval") || d.HasChange("max_keep") || d.HasChange("incremental") {
		_, err := client.VolumeAutoBackup.Update(id, map[string]interface{}{
			"name":        d.Get("name").(string),
			"hour":        hour,
			"minute":      minute,
			"interval":    d.Get("interval").(int),
			"max_keep":    d.Get("max_keep").(int),
			"incremental": d.Get("incremental").(bool),
		})
		if err != nil {
			return fmt.Errorf("error when update Volume AutoBackup [%s]: %v", id, err)
		}
	}
	return resourceVolumeAutoBackupRead(d, meta)
}

func resourceVolumeAutoBackupDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*CombinedConfig).goCMCClient()
	_, err := client.VolumeAutoBackup.Delete(d.Id())

	if err != nil {
		return fmt.Errorf("error delete volume autobackup: %v", err)
	}
	_, err = waitUntilVolumeAutoBackupDeleted(d, meta)
	if err != nil {
		return fmt.Errorf("error delete volume autobackup: %v", err)
	}
	return nil
}

func resourceVolumeAutoBackupImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := resourceVolumeAutoBackupRead(d, meta)
	return []*schema.ResourceData{d}, err
}

func waitUntilVolumeAutoBackupDeleted(d *schema.ResourceData, meta interface{}) (interface{}, error) {
	return waitUntilResourceDeleted(d, meta, WaitConf{
		Delay:      3 * time.Second,
		MinTimeout: 30 * time.Second,
	}, func(id string) (any, error) {
		return getClient(meta).VolumeAutoBackup.Get(id)
	})
}
