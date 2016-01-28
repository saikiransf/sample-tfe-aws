package ultradns

import (
	"fmt"
	"testing"

	"github.com/Ensighten/udnssdk"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccUltradnsDirpool(t *testing.T) {
	var record udnssdk.RRSet
	domain := "ultradns.phinze.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUltradnsDirpoolDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltraDNSRecordDirpoolMinimal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_dirpool.minimal", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "name", "dirpool-minimal"),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "type", "A"),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "ttl", "300"),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "description", "Minimal directional pool"),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "rdata.0.host", "192.168.0.10"),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "rdata.0.all_non_configured", "true"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "id", "dirpool-minimal.ultradns.phinze.com"),
					resource.TestCheckResourceAttr("ultradns_dirpool.minimal", "hostname", "dirpool-minimal.ultradns.phinze.com."),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckUltraDNSRecordDirpoolMaximal, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUltraDNSRecordExists("ultradns_dirpool.maximal", &record),
					// Specified
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "name", "dirpool-maximal"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "zone", domain),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "type", "A"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "ttl", "300"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "description", "Description of pool"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "conflict_resolve", "GEO"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "rdata.0.host", "1.2.3.4"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "rdata.0.all_non_configured", "true"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "rdata.1.host", "2.3.4.5"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "rdata.1.geo_info.0.name", "North America"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "rdata.2.host", "9.8.7.6"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "rdata.2.ip_info.0.name", "some Ips"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "no_response.0.geo_info.0.name", "nrGeo"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "no_response.0.ip_info.0.name", "nrIP"),
					// Generated
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "id", "dirpool-maximal.ultradns.phinze.com"),
					resource.TestCheckResourceAttr("ultradns_dirpool.maximal", "hostname", "dirpool-maximal.ultradns.phinze.com."),
				),
			},
		},
	})
}

func testAccCheckUltradnsDirpoolDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*udnssdk.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ultradns_dirpool" {
			continue
		}

		k := udnssdk.RRSetKey{
			Zone: rs.Primary.Attributes["zone"],
			Name: rs.Primary.Attributes["name"],
			Type: rs.Primary.Attributes["type"],
		}

		_, err := client.RRSets.Select(k)

		if err == nil {
			return fmt.Errorf("Record still exists")
		}
	}

	return nil
}

const testAccCheckUltraDNSRecordDirpoolMinimal = `
resource "ultradns_dirpool" "minimal" {
  zone        = "%s"
  name        = "dirpool-minimal"
  type        = "A"
  ttl         = 300
  description = "Minimal directional pool"

  rdata {
    host = "192.168.0.10"
    all_non_configured = true
  }
}
`

const testAccCheckUltraDNSRecordDirpoolMaximal = `
resource "ultradns_dirpool" "maximal" {
  zone        = "%s"
  name        = "dirpool-maximal"
  type        = "A"
  ttl         = 300
  description = "Description of pool"

  conflict_resolve = "GEO"

  rdata {
    host               = "1.2.3.4"
    all_non_configured = true
  }

  rdata {
    host = "2.3.4.5"

    geo_info {
      name = "North America"

      codes = [
        "US-OK",
        # "US-DC",
        # "US-MA",
      ]
    }
  }

  rdata {
    host = "9.8.7.6"

    ip_info {
      name = "some Ips"

      ips {
        start = "200.20.0.1"
        end   = "200.20.0.10"
      }

      ips {
        cidr = "20.20.20.0/24"
      }

      ips {
        address = "50.60.70.80"
      }
    }
  }

#   rdata {
#     host = "30.40.50.60"
#
#     geo_info {
#       name             = "accountGeoGroup"
#       is_account_level = true
#     }
#
#     ip_info {
#       name             = "accountIPGroup"
#       is_account_level = true
#     }
#   }

  no_response {
    geo_info {
      name = "nrGeo"

      codes = [
        "Z4",
      ]
    }

    ip_info {
      name = "nrIP"

      ips {
        address = "197.231.41.3"
      }
    }
  }
}
`
