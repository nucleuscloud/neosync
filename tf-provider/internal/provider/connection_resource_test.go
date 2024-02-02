package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Connection(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("neosync_connection.test1", "id"),
					resource.TestCheckResourceAttr("neosync_connection.test1", "name", "foo"),
					resource.TestCheckResourceAttr("neosync_connection.test1", "postgres.url", "test-url"),
				),
			},
			{
				Config: testAccConnectionConfigUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neosync_connection.test1", "postgres.url", "test-url2"),
				),
			},
		},
	})
}

const testAccConnectionConfig = `
resource "neosync_connection" "test1" {
  name = "foo"

	postgres = {
		url = "test-url"
	}
}
`
const testAccConnectionConfigUpdated = `
resource "neosync_connection" "test1" {
  name = "foo"

	postgres = {
		url = "test-url2"
	}
}
`
