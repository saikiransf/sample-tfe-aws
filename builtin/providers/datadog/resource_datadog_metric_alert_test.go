package datadog

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogMetricAlert_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMetricAlertDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckDatadogMetricAlertConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMetricAlertExists("datadog_metric_alert.foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "name", "name for metric_alert foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "message", "description for metric_alert foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "metric", "aws.ec2.cpu"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "tags.0", "environment:foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "tags.1", "host:foo"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "keys.0", "host"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "keys.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "time_aggr", "avg"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "time_window", "last_1h"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "space_aggr", "avg"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "operator", "<"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "notify", "@hipchat-<name>"),
					resource.TestCheckResourceAttr(
						"datadog_metric_alert.foo", "threshold", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogMetricAlertDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	if err := destroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckDatadogMetricAlertExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		if err := existsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

const testAccCheckDatadogMetricAlertConfigBasic = `
resource "datadog_metric_alert" "foo" {
  name = "name for metric_alert foo"
  message = "description for metric_alert foo"

  metric = "aws.ec2.cpu"
  tags = ["environment:foo", "host:foo"]
  keys = ["host"]

  time_aggr = "avg" // avg, sum, max, min, change, or pct_change
  time_window = "last_1h" // last_#m (5, 10, 15, 30), last_#h (1, 2, 4), or last_1d
  space_aggr = "avg" // avg, sum, min, or max
  operator = "<" // <, <=, >, >=, ==, or !=

  threshold = 1
  notify = "@hipchat-<name>"

  notify_no_data = false
  renotify_interval = 60
}
`
