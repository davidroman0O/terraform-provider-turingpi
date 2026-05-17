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
  description = "Turing Pi 2 BMC hostname or IP address."
  type        = string
}

variable "bmc_username" {
  description = "Turing Pi 2 BMC username."
  type        = string
  default     = "root"
}

variable "bmc_password" {
  description = "Turing Pi 2 BMC password."
  type        = string
  sensitive   = true
}
