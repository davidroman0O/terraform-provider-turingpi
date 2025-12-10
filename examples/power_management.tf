# Power management example
# Manages power state of node 1

resource "turingpi_node_power" "node1" {
  node     = 1
  power_on = true
}
