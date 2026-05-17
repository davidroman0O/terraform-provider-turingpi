resource "turingpi_node_usb" "node1_flash" {
  node = 1
  mode = "flash"
  bmc  = true
}
