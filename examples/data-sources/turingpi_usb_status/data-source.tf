data "turingpi_usb_status" "current" {}

output "usb_status" {
  value = {
    node  = data.turingpi_usb_status.current.node
    mode  = data.turingpi_usb_status.current.mode
    route = data.turingpi_usb_status.current.route
  }
}
