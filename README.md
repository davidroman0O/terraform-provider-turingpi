# Terraform Provider for Turing Pi 2

Manage your Turing Pi 2 BMC resources with Terraform.

## Features

- **Data Sources**: Query BMC info, power status, and USB status
- **Power Management**: Control power state of individual nodes (idempotent)
- **USB Configuration**: Set USB mode (host/device/flash) and routing
- **OS Flashing**: Flash OS images to nodes with caching support

## Requirements

- Terraform >= 1.0
- Go >= 1.23 (for building)
- Turing Pi 2 BMC accessible via network

## Installation

### From Terraform Registry (coming soon)

```hcl
terraform {
  required_providers {
    turingpi = {
      source  = "davidroman0O/turingpi"
      version = "~> 0.1"
    }
  }
}
```

### Local Development

```bash
make install
```

## Usage

### Provider Configuration

```hcl
provider "turingpi" {
  host     = "192.168.1.90"
  username = "root"
  password = "turing"
}
```

### Data Sources

```hcl
# Get BMC information
data "turingpi_info" "bmc" {}

output "bmc_version" {
  value = data.turingpi_info.bmc.version
}

# Get power status of all nodes
data "turingpi_power_status" "all" {}

output "node1_power" {
  value = data.turingpi_power_status.all.node1
}
```

### Resources

#### Power Management

```hcl
resource "turingpi_node_power" "node1" {
  node     = 1
  power_on = true
}
```

#### Flash OS Image

```hcl
resource "turingpi_node_flash" "node1_ubuntu" {
  node      = 1
  image_url = "https://firmware.turingpi.com/turing-rk1/ubuntu_22.04_rockchip_linux/v1.33/ubuntu-22.04.3-preinstalled-server-arm64-turing-rk1_v1.33.img.xz"
  cache     = "local"  # Options: "local", "bmc", "none"

  timeouts {
    create = "3h"
  }
}
```

## Caching

The flash resource supports caching to speed up repeated flashes:

- `local`: Cache images in `~/.cache/terraform-provider-turingpi/`
- `bmc`: Cache images on the BMC via SFTP (faster for flashing multiple nodes)
- `none`: No caching (download each time)

## Development

```bash
# Build
make build

# Install locally
make install

# Run tests
make test

# Run acceptance tests (requires BMC access)
export TURINGPI_HOST="192.168.1.90"
export TURINGPI_USERNAME="root"
export TURINGPI_PASSWORD="turing"
make testacc
```

## License

MPL-2.0
