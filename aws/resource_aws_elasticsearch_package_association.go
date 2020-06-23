package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsElasticsearchPackageAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticsearchPackageAssociationCreate,
		Read:   resourceAwsElasticsearchPackageAssociationRead,
		Delete: resourceAwsElasticsearchPackageAssociationDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"package_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_package_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"package_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"reference_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsElasticsearchPackageAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	domainName := d.Get("domain_name").(string)
	packageID := d.Get("package_id").(string)

	input := &elasticsearch.AssociatePackageInput{
		DomainName: aws.String(domainName),
		PackageID:  aws.String(packageID),
	}

	output, err := conn.AssociatePackage(input)
	if err != nil {
		return fmt.Errorf("Error associating Elasticsearch Package (%s) with Elasticsearch Domain (%s): %s", packageID, domainName, err)
	}

	id := fmt.Sprintf("association-%s-%s", aws.StringValue(output.DomainPackageDetails.DomainName), aws.StringValue(output.DomainPackageDetails.PackageID))

	d.SetId(id)

	return resourceAwsElasticsearchPackageAssociationRead(d, meta)
}

func resourceAwsElasticsearchPackageAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	input := &elasticsearch.DescribePackagesInput{
		Filters: []*elasticsearch.DescribePackagesFilter{
			{Name: aws.String("PackageID"), Value: []*string{aws.String(d.Id())}},
		},
	}
	out, err := conn.DescribePackages(input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
			log.Printf("[WARN] ElasticSearch Package (%s) not found, removing", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	log.Printf("[DEBUG] Received ElasticSearch Package: %s", out)

	if count := len(out.PackageDetailsList); count != 1 {
		return fmt.Errorf("unexpected number of packages returned: %d", count)
	}

	details, err := getElasticsearchPackageAssociation(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[WARN] ElasticSearch Package (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading ElasticSearch Package: %s", err)
	}

	if details == nil {
		log.Printf("[WARN] ElasticSearch Package (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(aws.StringValue(details.PackageID))
	d.Set("name", details.PackageName)
	d.Set("description", details.PackageDescription)
	d.Set("type", details.PackageType)

	return nil
}

func resourceAwsElasticsearchPackageAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	input := &elasticsearch.DeletePackageInput{PackageID: aws.String(d.Id())}
	_, err := conn.DeletePackage(input)
	if err != nil {
		return fmt.Errorf("Error deleting Elasticsearch Package: %v", err)
	}

	return nil
}

func getElasticsearchPackageAssociation(conn *elasticsearch.ElasticsearchService, packageID string) (*elasticsearch.PackageDetails, error) {
	input := &elasticsearch.DescribePackagesInput{
		Filters: []*elasticsearch.DescribePackagesFilter{
			{
				Name:  aws.String("PackageID"),
				Value: []*string{aws.String(packageID)},
			},
		},
	}
	output, err := conn.DescribePackages(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PackageDetailsList) == 0 {
		return nil, nil
	}

	return output.PackageDetailsList[0], nil
}
