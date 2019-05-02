package infoblox

import (
	"fmt"
	"log"
	"net/url"

	infoblox "github.com/fanatic/go-infoblox"
	"github.com/hashicorp/terraform/helper/schema"
)

func infobloxARecord() *schema.Resource {
	return &schema.Resource{
		Create: resourceInfobloxAoldRecordCreate,
		Read:   resourceInfobloxAoldRecordRead,
		Update: resourceInfobloxAoldRecordUpdate,
		Delete: resourceInfobloxAoldRecordDelete,

		Schema: map[string]*schema.Schema{
			// TODO: validate that address is in IPv4 format.
			"ipv4addr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"comment": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ttl": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"view": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
		},
	}
}

// aObjectFromAttributesOld created an infoblox.RecordAObject using the attributes
// as set by terraform.
// The Infoblox WAPI does not allow updates to the "view" field on an A record,
// so we also take a skipView arg to skip setting view.
func aObjectFromAttributesOld(d *schema.ResourceData, skipView bool) infoblox.RecordAObject {
	aObject := infoblox.RecordAObject{}

	aObject.Name = d.Get("name").(string)
	aObject.Ipv4Addr = d.Get("ipv4addr").(string)

	if attr, ok := d.GetOk("comment"); ok {
		aObject.Comment = attr.(string)
	}
	if attr, ok := d.GetOk("ttl"); ok {
		aObject.Ttl = attr.(int)
	}
	if skipView {
		return aObject
	}

	if attr, ok := d.GetOk("view"); ok {
		aObject.View = attr.(string)
	}

	return aObject
}

func resourceInfobloxAoldRecordCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	record := url.Values{}
	aRecordObject := aObjectFromAttributesOld(d, false)

	log.Printf("[DEBUG] Creating Infoblox A record with configuration: %#v", aRecordObject)

	opts := &infoblox.Options{
		ReturnFields: []string{"ipv4addr", "name", "comment", "ttl", "view"},
	}
	recordID, err := client.RecordA().Create(record, opts, aRecordObject)
	if err != nil {
		return fmt.Errorf("error creating infoblox A record: %s", err.Error())
	}

	d.SetId(recordID)
	log.Printf("[INFO] Infoblox A record created with ID: %s", d.Id())

	return resourceInfobloxAoldRecordRead(d, meta)
}

func resourceInfobloxAoldRecordRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	opts := &infoblox.Options{
		ReturnFields: []string{"ipv4addr", "name", "comment", "ttl", "view"},
	}
	record, err := client.GetRecordA(d.Id(), opts)
	if err != nil {
		return handleReadError(d, "A", err)
	}

	d.Set("address", record.Ipv4Addr)
	d.Set("name", record.Name)
	d.Set("comment", record.Comment)
	d.Set("ttl", record.Ttl)
	d.Set("view", record.View)

	return nil
}

func resourceInfobloxAoldRecordUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	opts := &infoblox.Options{
		ReturnFields: []string{"ipv4addr", "name", "comment", "ttl", "view"},
	}
	_, err := client.GetRecordA(d.Id(), opts)
	if err != nil {
		return fmt.Errorf("error finding infoblox A record: %s", err.Error())
	}

	record := url.Values{}
	aRecordObject := aObjectFromAttributesOld(d, true)

	log.Printf("[DEBUG] Updating Infoblox A record with configuration: %#v", record)

	recordID, err := client.RecordAObject(d.Id()).Update(record, opts, aRecordObject)
	if err != nil {
		return fmt.Errorf("error updating Infoblox A record: %s", err.Error())
	}

	d.SetId(recordID)
	log.Printf("[INFO] Infoblox A record updated with ID: %s", d.Id())

	return resourceInfobloxAoldRecordRead(d, meta)
}

func resourceInfobloxAoldRecordDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	log.Printf("[DEBUG] Deleting Infoblox A record: %s, %s", d.Get("name").(string), d.Id())
	_, err := client.GetRecordA(d.Id(), nil)
	if err != nil {
		return fmt.Errorf("error finding Infoblox A record: %s", err.Error())
	}

	err = client.RecordAObject(d.Id()).Delete(nil)
	if err != nil {
		return fmt.Errorf("error deleting Infoblox A record: %s", err.Error())
	}

	return nil
}
