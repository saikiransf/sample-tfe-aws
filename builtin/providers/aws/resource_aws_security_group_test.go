package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/hashicorp/aws-sdk-go/aws"
	"github.com/hashicorp/aws-sdk-go/gen/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSecurityGroup_normal(t *testing.T) {
	var group ec2.SecurityGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributes(&group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_self(t *testing.T) {
	var group ec2.SecurityGroup

	checkSelf := func(s *terraform.State) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("bad: %#v", group)
			}
		}()

		if *group.IPPermissions[0].UserIDGroupPairs[0].GroupID != *group.GroupID {
			return fmt.Errorf("bad: %#v", group)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfigSelf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3971148406.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3971148406.from_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3971148406.to_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3971148406.self", "true"),
					checkSelf,
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_vpc(t *testing.T) {
	var group ec2.SecurityGroup

	testCheck := func(*terraform.State) error {
		if *group.VPCID == "" {
			return fmt.Errorf("should have vpc ID")
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfigVpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributes(&group),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "name", "terraform_acceptance_test_example"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "description", "Used in the terraform acceptance tests"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "ingress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.from_port", "80"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.to_port", "8000"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_security_group.web", "egress.3629188364.cidr_blocks.0", "10.0.0.0/8"),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_MultiIngress(t *testing.T) {
	var group ec2.SecurityGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfigMultiIngress,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
				),
			},
		},
	})
}

func TestAccAWSSecurityGroup_Change(t *testing.T) {
	var group ec2.SecurityGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
				),
			},
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfigChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.web", &group),
					testAccCheckAWSSecurityGroupAttributesChanged(&group),
				),
			},
		},
	})
}

func testAccCheckAWSSecurityGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_security_group" {
			continue
		}

		// Retrieve our group
		req := &ec2.DescribeSecurityGroupsRequest{
			GroupIDs: []string{rs.Primary.ID},
		}
		resp, err := conn.DescribeSecurityGroups(req)
		if err == nil {
			if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupID == rs.Primary.ID {
				return fmt.Errorf("Security Group (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		ec2err, ok := err.(aws.APIError)
		if !ok {
			return err
		}
		// Confirm error code is what we want
		if ec2err.Code != "InvalidGroup.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSSecurityGroupExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		req := &ec2.DescribeSecurityGroupsRequest{
			GroupIDs: []string{rs.Primary.ID},
		}
		resp, err := conn.DescribeSecurityGroups(req)
		if err != nil {
			return err
		}

		if len(resp.SecurityGroups) > 0 && *resp.SecurityGroups[0].GroupID == rs.Primary.ID {

			log.Printf("\n==\n===\nfound group\n===\n==\n")
			*group = resp.SecurityGroups[0]

			return nil
		}

		return fmt.Errorf("Security Group not found")
	}
}

func testAccCheckAWSSecurityGroupAttributes(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := ec2.IPPermission{
			FromPort:   aws.Integer(80),
			ToPort:     aws.Integer(8000),
			IPProtocol: aws.String("tcp"),
			IPRanges:   []ec2.IPRange{ec2.IPRange{aws.String("10.0.0.0/8")}},
		}

		if *group.GroupName != "terraform_acceptance_test_example" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}

		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}

		if len(group.IPPermissions) == 0 {
			return fmt.Errorf("No IPPerms")
		}

		// Compare our ingress
		if !reflect.DeepEqual(group.IPPermissions[0], p) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IPPermissions[0],
				p)
		}

		return nil
	}
}

func TestAccAWSSecurityGroup_tags(t *testing.T) {
	var group ec2.SecurityGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSSecurityGroupConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.foo", &group),
					testAccCheckTags(&group.Tags, "foo", "bar"),
				),
			},

			resource.TestStep{
				Config: testAccAWSSecurityGroupConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityGroupExists("aws_security_group.foo", &group),
					testAccCheckTags(&group.Tags, "foo", ""),
					testAccCheckTags(&group.Tags, "bar", "baz"),
				),
			},
		},
	})
}

func testAccCheckAWSSecurityGroupAttributesChanged(group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		p := []ec2.IPPermission{
			ec2.IPPermission{
				FromPort:   aws.Integer(80),
				ToPort:     aws.Integer(9000),
				IPProtocol: aws.String("tcp"),
				IPRanges:   []ec2.IPRange{ec2.IPRange{aws.String("10.0.0.0/8")}},
			},
			ec2.IPPermission{
				FromPort:   aws.Integer(80),
				ToPort:     aws.Integer(8000),
				IPProtocol: aws.String("tcp"),
				IPRanges:   []ec2.IPRange{ec2.IPRange{aws.String("0.0.0.0/0")}, ec2.IPRange{aws.String("10.0.0.0/8")}},
			},
		}

		if *group.GroupName != "terraform_acceptance_test_example" {
			return fmt.Errorf("Bad name: %s", *group.GroupName)
		}

		if *group.Description != "Used in the terraform acceptance tests" {
			return fmt.Errorf("Bad description: %s", *group.Description)
		}

		// Compare our ingress
		if len(group.IPPermissions) != 2 {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IPPermissions,
				p)
		}

		if *group.IPPermissions[0].ToPort == 8000 {
			group.IPPermissions[1], group.IPPermissions[0] =
				group.IPPermissions[0], group.IPPermissions[1]
		}

		if !reflect.DeepEqual(group.IPPermissions, p) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.IPPermissions,
				p)
		}

		return nil
	}
}

const testAccAWSSecurityGroupConfig = `
resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}
`

const testAccAWSSecurityGroupConfigChange = `
resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 9000
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["0.0.0.0/0", "10.0.0.0/8"]
  }
}
`

const testAccAWSSecurityGroupConfigSelf = `
resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    self = true
  }
}
`

const testAccAWSSecurityGroupConfigVpc = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"
  vpc_id = "${aws_vpc.foo.id}"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

	egress {
		protocol = "tcp"
		from_port = 80
		to_port = 8000
		cidr_blocks = ["10.0.0.0/8"]
	}
}
`

const testAccAWSSecurityGroupConfigMultiIngress = `
resource "aws_security_group" "worker" {
  name = "terraform_acceptance_test_example_1"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }
}

resource "aws_security_group" "web" {
  name = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 22
    to_port = 22
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 800
    to_port = 800
    cidr_blocks = ["10.0.0.0/8"]
  }

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    security_groups = ["${aws_security_group.worker.id}"]
  }
}
`

const testAccAWSSecurityGroupConfigTags = `
resource "aws_security_group" "foo" {
	name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags {
    foo = "bar"
  }
}
`

const testAccAWSSecurityGroupConfigTagsUpdate = `
resource "aws_security_group" "foo" {
  name = "terraform_acceptance_test_example"
  description = "Used in the terraform acceptance tests"

  ingress {
    protocol = "tcp"
    from_port = 80
    to_port = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags {
    bar = "baz"
  }
}
`
