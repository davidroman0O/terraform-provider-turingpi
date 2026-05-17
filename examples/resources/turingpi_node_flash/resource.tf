resource "turingpi_node_flash" "node1_ubuntu" {
  node      = 1
  image_url = "https://firmware.turingpi.com/turing-rk1/ubuntu_22.04_rockchip_linux/v1.33/ubuntu-22.04.3-preinstalled-server-arm64-turing-rk1_v1.33.img.xz"
  cache     = "local"

  timeouts {
    create = "3h"
  }
}
