# Guard a GKE node pool's boot disk choice, which is the case a plain
# supports_disk_type check would wave through: hyperdisk-extreme is supported by
# this machine type, but not as a boot disk.
# => false
output "bootable" {
  value = provider::google-gce-metadata::supports_boot_disk_type("n2-standard-64", "hyperdisk-extreme")
}

check "node_pool_disk_type_is_bootable" {
  assert {
    condition     = provider::google-gce-metadata::supports_boot_disk_type(var.machine_type, var.disk_type)
    error_message = "${var.machine_type} cannot boot from ${var.disk_type}."
  }
}
