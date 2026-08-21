# Only the disk types an n2-standard-64 can actually boot from. Hyperdisk
# Extreme is supported by this machine type as a data disk, but is not bootable,
# so it does not appear here.
# => ["pd-balanced", "pd-ssd", "pd-standard"]
output "bootable" {
  value = provider::google-gce-metadata::boot_disk_types("n2-standard-64")
}

# Choose a GKE node pool boot disk without hardcoding one per machine type:
# prefer hyperdisk-balanced where the machine type can boot from it, and fall
# back to pd-balanced otherwise. Most machine types support more than one
# bootable disk type, so the list has to be narrowed by preference rather than
# assumed to hold a single element.
locals {
  preferred_boot_disk_types = ["hyperdisk-balanced", "pd-balanced"]

  boot_disk_type = coalesce([
    for disk_type in local.preferred_boot_disk_types :
    disk_type
    if contains(provider::google-gce-metadata::boot_disk_types(var.machine_type), disk_type)
  ]...)
}

resource "google_container_node_pool" "example" {
  name = "example"

  node_config {
    disk_type    = local.boot_disk_type
    machine_type = var.machine_type
  }
}
