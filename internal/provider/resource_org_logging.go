package cpln

import (
	"context"

	client "terraform-provider-cpln/internal/provider/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrgLogging() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceOrgCreate,
		ReadContext:   resourceOrgRead,
		UpdateContext: resourceOrgUpdate,
		DeleteContext: resourceOrgLoggingDelete,
		Schema: map[string]*schema.Schema{
			"cpln_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"s3_logging": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"s3_logging", "coralogix_logging", "datadog_logging", "logzio_logging"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"region": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "/",
						},
						"credentials": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"coralogix_logging": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"s3_logging", "coralogix_logging", "datadog_logging", "logzio_logging"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cluster": {
							Type:     schema.TypeString,
							Required: true,
						},
						"credentials": {
							Type:     schema.TypeString,
							Required: true,
						},
						"app": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subsystem": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"datadog_logging": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"s3_logging", "coralogix_logging", "datadog_logging", "logzio_logging"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host": {
							Type:     schema.TypeString,
							Required: true,
						},
						"credentials": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"logzio_logging": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				ExactlyOneOf: []string{"s3_logging", "coralogix_logging", "datadog_logging", "logzio_logging"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"listener_host": {
							Type:     schema.TypeString,
							Required: true,
						},
						"credentials": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{},
	}
}

func resourceOrgCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// log.Printf("[INFO] Method: resourceOrgCreate")

	c := m.(*client.Client)

	// org, _, err := c.GetOrg()
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// if org.Spec == nil {
	// 	org.Spec = &client.OrgSpec{}
	// }

	// // Clear out all logging
	// org.Spec.Logging = nil

	var logCreate *client.Logging

	logArray := d.Get("s3_logging").([]interface{})
	if len(logArray) == 1 {
		logCreate = buildS3Logging(logArray)
	}

	logArray = d.Get("coralogix_logging").([]interface{})
	if len(logArray) == 1 {
		logCreate = buildCoralogixLogging(logArray)
	}

	logArray = d.Get("datadog_logging").([]interface{})
	if len(logArray) == 1 {
		logCreate = buildDatadogLogging(logArray)
	}

	logArray = d.Get("logzio_logging").([]interface{})
	if len(logArray) == 1 {
		logCreate = buildLogzioLogging(logArray)
	}

	org, _, err := c.UpdateOrgLogging(logCreate)
	if err != nil {
		return diag.FromErr(err)
	}

	return setOrg(d, org)
}

func resourceOrgRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// log.Printf("[INFO] Method: resourceOrgtRead")

	c := m.(*client.Client)
	org, _, err := c.GetOrg()

	if err != nil {
		return diag.FromErr(err)
	}

	return setOrg(d, org)
}

func setOrg(d *schema.ResourceData, org *client.Org) diag.Diagnostics {

	if org == nil {
		d.SetId("")
		return nil
	}

	d.SetId(*org.Name)

	if err := SetBase(d, org.Base); err != nil {
		return diag.FromErr(err)
	}

	if org.Spec != nil && org.Spec.Logging != nil {

		if org.Spec.Logging.S3 != nil {
			if err := d.Set("s3_logging", flattenS3Logging(org.Spec.Logging.S3)); err != nil {
				return diag.FromErr(err)
			}
		}

		if org.Spec.Logging.Coralogix != nil {
			if err := d.Set("coralogix_logging", flattenCoralogixLogging(org.Spec.Logging.Coralogix)); err != nil {
				return diag.FromErr(err)
			}
		}

		if org.Spec.Logging.Datadog != nil {
			if err := d.Set("datadog_logging", flattenDatadogLogging(org.Spec.Logging.Datadog)); err != nil {
				return diag.FromErr(err)
			}
		}

		if org.Spec.Logging.Logzio != nil {
			if err := d.Set("logzio_logging", flattenLogzioLogging(org.Spec.Logging.Logzio)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return nil
}

func flattenS3Logging(log *client.S3Logging) []interface{} {

	if log != nil {

		outputMap := make(map[string]interface{})

		outputMap["bucket"] = *log.Bucket
		outputMap["region"] = *log.Region
		outputMap["prefix"] = *log.Prefix
		outputMap["credentials"] = *log.Credentials

		output := make([]interface{}, 1)
		output[0] = outputMap

		return output
	}

	return nil
}

func flattenCoralogixLogging(log *client.CoralogixLogging) []interface{} {

	if log != nil {

		outputMap := make(map[string]interface{})

		outputMap["cluster"] = *log.Cluster
		outputMap["credentials"] = *log.Credentials
		outputMap["app"] = *log.App
		outputMap["subsystem"] = *log.Subsystem

		output := make([]interface{}, 1)
		output[0] = outputMap

		return output
	}

	return nil
}

func flattenDatadogLogging(log *client.DatadogLogging) []interface{} {

	if log != nil {

		outputMap := make(map[string]interface{})

		outputMap["host"] = *log.Host
		outputMap["credentials"] = *log.Credentials

		output := make([]interface{}, 1)
		output[0] = outputMap

		return output
	}

	return nil
}

func flattenLogzioLogging(log *client.LogzioLogging) []interface{} {

	if log != nil {

		outputMap := make(map[string]interface{})

		outputMap["listener_host"] = *log.ListenerHost
		outputMap["credentials"] = *log.Credentials

		output := make([]interface{}, 1)
		output[0] = outputMap

		return output
	}

	return nil
}

func buildS3Logging(logging []interface{}) *client.Logging {

	if len(logging) == 1 {

		log := logging[0].(map[string]interface{})

		iLog := &client.S3Logging{}
		iLog.Bucket = GetString(log["bucket"])
		iLog.Region = GetString(log["region"])
		iLog.Prefix = GetString(log["prefix"])
		iLog.Credentials = GetString(log["credentials"])

		return &client.Logging{
			S3: iLog,
		}
	}

	return nil
}

func buildCoralogixLogging(logging []interface{}) *client.Logging {

	if len(logging) == 1 {

		log := logging[0].(map[string]interface{})

		iLog := &client.CoralogixLogging{}
		iLog.Cluster = GetString(log["cluster"])
		iLog.Credentials = GetString(log["credentials"])
		iLog.App = GetString(log["app"])
		iLog.Subsystem = GetString(log["subsystem"])

		return &client.Logging{
			Coralogix: iLog,
		}
	}

	return nil
}

func buildDatadogLogging(logging []interface{}) *client.Logging {

	if len(logging) == 1 {

		log := logging[0].(map[string]interface{})

		iLog := &client.DatadogLogging{}
		iLog.Host = GetString(log["host"])
		iLog.Credentials = GetString(log["credentials"])

		return &client.Logging{
			Datadog: iLog,
		}
	}

	return nil
}

func buildLogzioLogging(logging []interface{}) *client.Logging {

	if len(logging) == 1 {

		log := logging[0].(map[string]interface{})

		iLog := &client.LogzioLogging{}
		iLog.ListenerHost = GetString(log["listener_host"])
		iLog.Credentials = GetString(log["credentials"])

		return &client.Logging{
			Logzio: iLog,
		}
	}

	return nil
}

func resourceOrgUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// log.Printf("[INFO] Method: resourceOrgUpdate")

	if d.HasChanges("s3_logging", "coralogix_logging", "datadog_logging", "logzio_logging") {

		c := m.(*client.Client)

		// org, _, err := c.GetOrg()
		// if err != nil {
		// 	return diag.FromErr(err)
		// }

		// if org.Spec == nil {
		// 	org.Spec = &client.OrgSpec{}
		// }

		// // Clear out all logging
		// org.Spec.Logging = nil

		var logUpdate *client.Logging

		if d.HasChange("s3_logging") {
			logArray := d.Get("s3_logging").([]interface{})

			if logArray != nil {
				logUpdate = buildS3Logging(logArray)
			}
		}

		if d.HasChange("coralogix_logging") {
			logArray := d.Get("coralogix_logging").([]interface{})

			if logArray != nil {
				logUpdate = buildCoralogixLogging(logArray)
			}
		}

		if d.HasChange("datadog_logging") {
			logArray := d.Get("datadog_logging").([]interface{})

			if logArray != nil {
				logUpdate = buildDatadogLogging(logArray)
			}
		}

		if d.HasChange("logzio_logging") {
			logArray := d.Get("logzio_logging").([]interface{})

			if logArray != nil {
				logUpdate = buildLogzioLogging(logArray)
			}
		}

		org, _, err := c.UpdateOrgLogging(logUpdate)
		if err != nil {
			return diag.FromErr(err)
		}

		return setOrg(d, org)
	}

	return nil
}

func resourceOrgLoggingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// log.Printf("[INFO] Method: resourceOrgDelete")

	c := m.(*client.Client)

	_, _, err := c.UpdateOrgLogging(nil)

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
