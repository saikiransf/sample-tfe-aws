package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSInspectorTemplate_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSInspectorTemplateAssessment,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists("aws_inspector_assessment_template.foo"),
				),
			},
			resource.TestStep{
				Config: testAccCheckAWSInspectorTemplatetModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTargetExists("aws_inspector_assessment_template.foo"),
				),
			},
		},
	})
}

func testAccCheckAWSInspectorTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).inspectorconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_inspector_assessment_template" {
			continue
		}

		resp, err := conn.DescribeAssessmentTemplates(&inspector.DescribeAssessmentTemplatesInput{
			AssessmentTemplateArns: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			if inspectorerr, ok := err.(awserr.Error); ok && inspectorerr.Code() == "InvalidInputException" {
				return nil
			} else {
				return fmt.Errorf("Error finding Inspector Assessment Template: %s", err)
			}
		}

		if len(resp.AssessmentTemplates) > 0 {
			return fmt.Errorf("Found Template, expected none: %s", resp)
		}
	}

	return nil
}

func testAccCheckAWSInspectorTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

var testAccAWSInspectorTemplateAssessment = `
resource "aws_inspector_resource_group" "foo" {
	tags {
	  Name  = "bar"
  }
}

resource "aws_inspector_assessment_target" "foo" {
	name = "foo"
	resource_group_arn =  "${aws_inspector_resource_group.foo.arn}"
}

resource "aws_inspector_assessment_template" "foo" {
  name = "foo template"
  target_arn    = "${aws_inspector_assessment_target.foo.arn}"
  duration      = 3600

  rules_package_arns = [
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-9hgA516p",
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-H5hpSawc",
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-JJOtZiqQ",
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-vg5GGHSD",
  ]
}`

var testAccCheckAWSInspectorTemplatetModified = `
resource "aws_inspector_resource_group" "foo" {
	tags {
	  Name  = "bar"
  }
}

resource "aws_inspector_assessment_target" "foo" {
	name = "foo"
	resource_group_arn =  "${aws_inspector_resource_group.foo.arn}"
}

resource "aws_inspector_assessment_template" "foo" {
  name = "bar template"
  target_arn    = "${aws_inspector_assessment_target.foo.arn}"
  duration      = 3600

  rules_package_arns = [
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-9hgA516p",
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-H5hpSawc",
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-JJOtZiqQ",
	  "arn:aws:inspector:us-west-2:758058086616:rulespackage/0-vg5GGHSD",
  ]
}`
