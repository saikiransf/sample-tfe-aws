package scaleway

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccScalewayDataSourceBootscript_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayBootscriptConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBootscriptID("data.scaleway_bootscript.debug"),
					resource.TestCheckResourceAttr("data.scaleway_bootscript.debug", "architecture", "x86_64"),
					resource.TestCheckResourceAttr("data.scaleway_bootscript.debug", "public", "true"),
					resource.TestMatchResourceAttr("data.scaleway_bootscript.debug", "kernel", regexp.MustCompile("4.8.3")),
				),
			},
		},
	})
}

func TestAccScalewayDataSourceBootscript_Filtered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckScalewayBootscriptFilterConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBootscriptID("data.scaleway_bootscript.debug"),
					resource.TestCheckResourceAttr("data.scaleway_bootscript.debug", "architecture", "arm"),
					resource.TestCheckResourceAttr("data.scaleway_bootscript.debug", "public", "true"),
				),
			},
		},
	})
}

func testAccCheckBootscriptID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find bootscript data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("bootscript data source ID not set")
		}
		return nil
	}
}

const testAccCheckScalewayBootscriptConfig = `
data "scaleway_bootscript" "debug" {
  name = "x86_64 4.8.3 debug #1"
}
`

const testAccCheckScalewayBootscriptFilterConfig = `
data "scaleway_bootscript" "debug" {
  architecture = "arm"
  name_filter = "Rescue"
}
`
