data "turingpi_power_status" "all" {}

output "node_power_status" {
  value = {
    node1 = data.turingpi_power_status.all.node1
    node2 = data.turingpi_power_status.all.node2
    node3 = data.turingpi_power_status.all.node3
    node4 = data.turingpi_power_status.all.node4
  }
}
