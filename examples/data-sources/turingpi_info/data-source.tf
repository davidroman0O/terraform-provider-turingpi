data "turingpi_info" "bmc" {}

output "bmc_info" {
  value = {
    version = data.turingpi_info.bmc.version
    ip      = data.turingpi_info.bmc.ip
    mac     = data.turingpi_info.bmc.mac
  }
}
