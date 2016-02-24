package aws

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/mitchellh/go-homedir"

	"errors"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsLambdaFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLambdaFunctionCreate,
		Read:   resourceAwsLambdaFunctionRead,
		Update: resourceAwsLambdaFunctionUpdate,
		Delete: resourceAwsLambdaFunctionDelete,

		Schema: map[string]*schema.Schema{
			"filename": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"s3_bucket", "s3_key", "s3_object_version"},
			},
			"s3_bucket": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filename"},
			},
			"s3_key": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filename"},
			},
			"s3_object_version": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filename"},
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"handler": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"memory_size": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  128,
			},
			"role": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"runtime": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "nodejs",
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3,
			},
			"update_code": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"vpc_config": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": &schema.Schema{
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"security_group_ids": &schema.Schema{
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_code_hash": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_code_hash": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// resourceAwsLambdaFunction maps to:
// CreateFunction in the API / SDK
func resourceAwsLambdaFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	functionName := d.Get("function_name").(string)
	iamRole := d.Get("role").(string)

	log.Printf("[DEBUG] Creating Lambda Function %s with role %s", functionName, iamRole)

	var functionCode *lambda.FunctionCode
	if v, ok := d.GetOk("filename"); ok {
		zipfile, shaSum, err := loadLocalZipFile(v.(string))
		if err != nil {
			return err
		}
		d.Set("source_code_hash", shaSum)
		functionCode = &lambda.FunctionCode{
			ZipFile: zipfile,
		}
	} else {
		s3Bucket, bucketOk := d.GetOk("s3_bucket")
		s3Key, keyOk := d.GetOk("s3_key")
		s3ObjectVersion, versionOk := d.GetOk("s3_object_version")
		if !bucketOk || !keyOk || !versionOk {
			return errors.New("s3_bucket, s3_key and s3_object_version must all be set while using S3 code source")
		}
		functionCode = &lambda.FunctionCode{
			S3Bucket:        aws.String(s3Bucket.(string)),
			S3Key:           aws.String(s3Key.(string)),
			S3ObjectVersion: aws.String(s3ObjectVersion.(string)),
		}
	}

	params := &lambda.CreateFunctionInput{
		Code:         functionCode,
		Description:  aws.String(d.Get("description").(string)),
		FunctionName: aws.String(functionName),
		Handler:      aws.String(d.Get("handler").(string)),
		MemorySize:   aws.Int64(int64(d.Get("memory_size").(int))),
		Role:         aws.String(iamRole),
		Runtime:      aws.String(d.Get("runtime").(string)),
		Timeout:      aws.Int64(int64(d.Get("timeout").(int))),
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		configs := v.([]interface{})

		if len(configs) > 1 {
			return errors.New("Only a single vpc_config block is expected")
		} else if len(configs) == 1 {
			config := configs[0].(map[string]interface{})
			var subnetIds []*string
			for _, id := range config["subnet_ids"].(*schema.Set).List() {
				subnetIds = append(subnetIds, aws.String(id.(string)))
			}

			var securityGroupIds []*string
			for _, id := range config["security_group_ids"].(*schema.Set).List() {
				securityGroupIds = append(securityGroupIds, aws.String(id.(string)))
			}

			var vpcConfig = &lambda.VpcConfig{
				SubnetIds:        subnetIds,
				SecurityGroupIds: securityGroupIds,
			}
			params.VpcConfig = vpcConfig
		}
	}

	// IAM profiles can take ~10 seconds to propagate in AWS:
	//  http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-roles-for-amazon-ec2.html#launch-instance-with-role-console
	// Error creating Lambda function: InvalidParameterValueException: The role defined for the task cannot be assumed by Lambda.
	err := resource.Retry(1*time.Minute, func() error {
		_, err := conn.CreateFunction(params)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() == "InvalidParameterValueException" {
					log.Printf("[DEBUG] InvalidParameterValueException creating Lambda Function: %s", awserr)
					// Retryable
					return awserr
				}
			}
			log.Printf("[DEBUG] Error creating Lambda Function: %s", err)
			// Not retryable
			return resource.RetryError{Err: err}
		}
		// No error
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error creating Lambda function: %s", err)
	}

	d.SetId(d.Get("function_name").(string))

	return resourceAwsLambdaFunctionRead(d, meta)
}

// resourceAwsLambdaFunctionRead maps to:
// GetFunction in the API / SDK
func resourceAwsLambdaFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[DEBUG] Fetching Lambda Function: %s", d.Id())

	params := &lambda.GetFunctionInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	getFunctionOutput, err := conn.GetFunction(params)
	if err != nil {
		return err
	}

	// getFunctionOutput.Code.Location is a pre-signed URL pointing at the zip
	// file that we uploaded when we created the resource. You can use it to
	// download the code from AWS. The other part is
	// getFunctionOutput.Configuration which holds metadata.

	function := getFunctionOutput.Configuration
	// TODO error checking / handling on the Set() calls.
	d.Set("arn", function.FunctionArn)
	d.Set("description", function.Description)
	d.Set("handler", function.Handler)
	d.Set("memory_size", function.MemorySize)
	d.Set("last_modified", function.LastModified)
	d.Set("role", function.Role)
	d.Set("runtime", function.Runtime)
	d.Set("timeout", function.Timeout)
	d.Set("vpc_config", flattenLambdaVpcConfigResponse(function.VpcConfig))

	// Compare code hashes, and see if an update is required to code. If there
	// is, set the "update_code" attribute.

	remoteSum, err := decodeBase64(*function.CodeSha256)
	if err != nil {
		return err
	}
	_, localSum, err := loadLocalZipFile(d.Get("filename").(string))
	if err != nil {
		return err
	}
	d.Set("remote_code_hash", remoteSum)
	d.Set("source_code_hash", localSum)

	if remoteSum != localSum {
		d.Set("update_code", true)
	} else {
		d.Set("update_code", false)
	}

	return nil
}

// resourceAwsLambdaFunction maps to:
// DeleteFunction in the API / SDK
func resourceAwsLambdaFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	log.Printf("[INFO] Deleting Lambda Function: %s", d.Id())

	params := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
	}

	_, err := conn.DeleteFunction(params)
	if err != nil {
		return fmt.Errorf("Error deleting Lambda Function: %s", err)
	}

	d.SetId("")

	return nil
}

// resourceAwsLambdaFunctionUpdate maps to:
// UpdateFunctionCode in the API / SDK
func resourceAwsLambdaFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lambdaconn

	d.Partial(true)

	codeReq := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(d.Id()),
	}

	codeUpdate := false
	if sourceHash, ok := d.GetOk("source_code_hash"); ok {
		zipfile, shaSum, err := loadLocalZipFile(d.Get("filename").(string))
		if err != nil {
			return err
		}
		if sourceHash != shaSum {
			d.SetPartial("filename")
			d.SetPartial("source_code_hash")
		}
		codeReq.ZipFile = zipfile
		codeUpdate = true
	}
	if d.HasChange("s3_bucket") || d.HasChange("s3_key") || d.HasChange("s3_object_version") {
		d.SetPartial("s3_bucket")
		d.SetPartial("s3_key")
		d.SetPartial("s3_object_version")

		codeReq.S3Bucket = aws.String(d.Get("s3_bucket").(string))
		codeReq.S3Key = aws.String(d.Get("s3_key").(string))
		codeReq.S3ObjectVersion = aws.String(d.Get("s3_object_version").(string))
		codeUpdate = true
	}

	log.Printf("[DEBUG] Send Update Lambda Function Code request: %#v", codeReq)
	if codeUpdate {
		_, err := conn.UpdateFunctionCode(codeReq)
		if err != nil {
			return fmt.Errorf("Error modifying Lambda Function Code %s: %s", d.Id(), err)
		}
	}

	configReq := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(d.Id()),
	}

	configUpdate := false
	if d.HasChange("description") {
		d.SetPartial("description")
		configReq.Description = aws.String(d.Get("description").(string))
		configUpdate = true
	}
	if d.HasChange("handler") {
		d.SetPartial("handler")
		configReq.Handler = aws.String(d.Get("handler").(string))
		configUpdate = true
	}
	if d.HasChange("memory_size") {
		d.SetPartial("memory_size")
		configReq.MemorySize = aws.Int64(int64(d.Get("memory_size").(int)))
		configUpdate = true
	}
	if d.HasChange("role") {
		d.SetPartial("role")
		configReq.Role = aws.String(d.Get("role").(string))
		configUpdate = true
	}
	if d.HasChange("timeout") {
		d.SetPartial("timeout")
		configReq.Timeout = aws.Int64(int64(d.Get("timeout").(int)))
		configUpdate = true
	}

	log.Printf("[DEBUG] Send Update Lambda Function Configuration request: %#v", configReq)
	if configUpdate {
		_, err := conn.UpdateFunctionConfiguration(configReq)
		if err != nil {
			return fmt.Errorf("Error modifying Lambda Function Configuration %s: %s", d.Id(), err)
		}
	}
	d.Partial(false)

	return resourceAwsLambdaFunctionRead(d, meta)
}

// loads the local ZIP data and the SHA sum of the data.
func loadLocalZipFile(v string) ([]byte, string, error) {
	filename, err := homedir.Expand(v)
	if err != nil {
		return nil, "", err
	}
	zipfile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(zipfile)
	return zipfile, fmt.Sprintf("%x", sum), nil
}

// Decodes a base64 string to a string.
func decodeBase64(s string) (string, error) {
	sum, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sum), nil
}
