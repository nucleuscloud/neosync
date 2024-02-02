package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAcc_Connection_DataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccExampleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.neosync_connection.test", "id", "d3fec0df-4333-4709-bd83-31e783ae9cb0"),
					resource.TestCheckResourceAttr("data.neosync_connection.test", "name", "asdfsdf"),
					// resource.TestCheckResourceAttr("data.neosync_connection.test", "name", "foosdf"),
				),
			},
		},
	})
}

const testAccExampleDataSourceConfig = `
data "neosync_connection" "test" {
  id = "d3fec0df-4333-4709-bd83-31e783ae9cb0"
}
`
