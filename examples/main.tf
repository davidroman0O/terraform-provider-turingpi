terraform {
  required_providers {
    turingpi = {
      source  = "davidroman0O/turingpi"
      version = "~> 0.1"
    }
  }
}

provider "turingpi" {
  host     = var.bmc_host
  username = var.bmc_username
  password = var.bmc_password
}

variable "bmc_host" {
  description = "Turing Pi BMC IP address or hostname"
  type        = string
  default     = "192.168.1.90"
}

variable "bmc_username" {
  description = "BMC username"
  type        = string
  default     = "root"
}

variable "bmc_password" {
  description = "BMC password"
  type        = string
  sensitive   = true
  default     = "turing"
}

# Get BMC information
data "turingpi_info" "bmc" {}

output "bmc_info" {
  description = "BMC information"
  value = {
    version = data.turingpi_info.bmc.version
    ip      = data.turingpi_info.bmc.ip
    mac     = data.turingpi_info.bmc.mac
  }
}

# Get power status of all nodes
data "turingpi_power_status" "all" {}

output "power_status" {
  description = "Power status of all nodes"
  value = {
    node1 = data.turingpi_power_status.all.node1
    node2 = data.turingpi_power_status.all.node2
    node3 = data.turingpi_power_status.all.node3
    node4 = data.turingpi_power_status.all.node4
  }
}
