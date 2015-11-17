package aws

import (
	"testing"

	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

var s3File, s3Err = ioutil.TempFile("", "tf.zip")

func TestAccAWSBeanstalkAppVersion_basic(t *testing.T) {
	ioutil.WriteFile(s3File.Name(), []byte("{anything will do }"), 0644)

	var appVersion elasticbeanstalk.ApplicationVersionDescription

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			if s3Err != nil {
				panic(s3Err)
			}
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckApplicationVersionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccApplicationVersionConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists("aws_elastic_beanstalk_application_version.default", &appVersion),
				),
			},
		},
	})
}

func testAccCheckApplicationVersionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastic_beanstalk_application_version" {
			continue
		}

		describeApplicationVersionOpts := &elasticbeanstalk.DescribeApplicationVersionsInput{
			VersionLabels: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeApplicationVersions(describeApplicationVersionOpts)
		if err == nil {
			if len(resp.ApplicationVersions) > 0 {
				return fmt.Errorf("Elastic Beanstalk Application Verson still exists.")
			}

			return nil
		}
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidParameterValue" {
			return err
		}
	}

	return nil
}

func testAccCheckApplicationVersionExists(n string, app *elasticbeanstalk.ApplicationVersionDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk Application Version is not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticbeanstalkconn
		describeApplicationVersionOpts := &elasticbeanstalk.DescribeApplicationVersionsInput{
			VersionLabels: []*string{aws.String(rs.Primary.ID)},
		}

		log.Printf("[DEBUG] Elastic Beanstalk Application Version TEST describe opts: %s", describeApplicationVersionOpts)

		resp, err := conn.DescribeApplicationVersions(describeApplicationVersionOpts)
		if err != nil {
			return err
		}
		if len(resp.ApplicationVersions) == 0 {
			return fmt.Errorf("Elastic Beanstalk Application Version not found.")
		}

		*app = *resp.ApplicationVersions[0]

		return nil
	}
}

var randomBeanstalkBucket = randInt
var testAccApplicationVersionConfig = fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%d"
}

resource "aws_s3_bucket_object" "default" {
  bucket = "${aws_s3_bucket.default.id}"
  key = "beanstalk/go-v1.zip"
  source = "%s"
}

resource "aws_elastic_beanstalk_application" "default" {
  name = "tf-test-name"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = "tf-test-name"
  name = "tf-test-version-label"
  bucket = "${aws_s3_bucket.default.id}"
  key = "${aws_s3_bucket_object.default.id}"
}
`, randomBeanstalkBucket, s3File.Name())
