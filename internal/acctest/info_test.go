// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInfoDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInfoDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.turingpi_info.test", "version"),
					resource.TestCheckResourceAttrSet("data.turingpi_info.test", "ip"),
				),
			},
		},
	})
}

const testAccInfoDataSourceConfig = `
data "turingpi_info" "test" {}
`

func TestAccPowerStatusDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { PreCheck(t) },
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPowerStatusDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.turingpi_power_status.test", "node1"),
					resource.TestCheckResourceAttrSet("data.turingpi_power_status.test", "node2"),
					resource.TestCheckResourceAttrSet("data.turingpi_power_status.test", "node3"),
					resource.TestCheckResourceAttrSet("data.turingpi_power_status.test", "node4"),
				),
			},
		},
	})
}

const testAccPowerStatusDataSourceConfig = `
data "turingpi_power_status" "test" {}
`
