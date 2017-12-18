package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsGameliftAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftAliasCreate,
		Read:   resourceAwsGameliftAliasRead,
		Update: resourceAwsGameliftAliasUpdate,
		Delete: resourceAwsGameliftAliasDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"routing_strategy": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fleet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"message": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsGameliftAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	rs := expandGameliftRoutingStrategy(d.Get("routing_strategy").([]interface{}))
	input := gamelift.CreateAliasInput{
		Name:            aws.String(d.Get("name").(string)),
		RoutingStrategy: rs,
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	log.Printf("[INFO] Creating Gamelift Alias: %s", input)
	out, err := conn.CreateAlias(&input)
	if err != nil {
		return err
	}

	d.SetId(*out.Alias.AliasId)

	return resourceAwsGameliftAliasRead(d, meta)
}

func resourceAwsGameliftAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Describing Gamelift Alias: %s", d.Id())
	out, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
		AliasId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	a := out.Alias

	d.Set("arn", a.AliasArn)
	d.Set("description", a.Description)
	d.Set("name", a.Name)
	d.Set("routing_strategy", flattenGameliftRoutingStrategy(a.RoutingStrategy))

	return nil
}

func resourceAwsGameliftAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Updating Gamelift Alias: %s", d.Id())
	_, err := conn.UpdateAlias(&gamelift.UpdateAliasInput{
		AliasId:         aws.String(d.Id()),
		Name:            aws.String(d.Get("name").(string)),
		Description:     aws.String(d.Get("description").(string)),
		RoutingStrategy: expandGameliftRoutingStrategy(d.Get("routing_strategy").([]interface{})),
	})
	if err != nil {
		return err
	}

	return resourceAwsGameliftAliasRead(d, meta)
}

func resourceAwsGameliftAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	log.Printf("[INFO] Deleting Gamelift Alias: %s", d.Id())
	_, err := conn.DeleteAlias(&gamelift.DeleteAliasInput{
		AliasId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	return nil
}

func expandGameliftRoutingStrategy(cfg []interface{}) *gamelift.RoutingStrategy {
	strategy := cfg[0].(map[string]interface{})
	return &gamelift.RoutingStrategy{
		FleetId: aws.String(strategy["fleet_id"].(string)),
		Message: aws.String(strategy["message"].(string)),
		Type:    aws.String(strategy["type"].(string)),
	}
}

func flattenGameliftRoutingStrategy(rs *gamelift.RoutingStrategy) []interface{} {
	m := make(map[string]interface{}, 0)
	m["fleet_id"] = *rs.FleetId
	m["message"] = *rs.Message
	m["type"] = *rs.Type

	return []interface{}{m}
}
