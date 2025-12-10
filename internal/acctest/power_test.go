// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNodePowerResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccNodePowerResourceConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("turingpi_node_power.test", "node", "1"),
					resource.TestCheckResourceAttr("turingpi_node_power.test", "power_on", "true"),
				),
			},
			// Update testing
			{
				Config: testAccNodePowerResourceConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("turingpi_node_power.test", "power_on", "false"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "turingpi_node_power.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNodePowerResourceConfig(powerOn bool) string {
	powerOnStr := "false"
	if powerOn {
		powerOnStr = "true"
	}
	return `
resource "turingpi_node_power" "test" {
  node     = 1
  power_on = ` + powerOnStr + `
}
`
}
