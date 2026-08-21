# Every disk type an n2-standard-64 can use, in any position.
# => ["hyperdisk-extreme", "hyperdisk-throughput", "pd-balanced", "pd-extreme", "pd-ssd", "pd-standard"]
output "n2_disk_types" {
  value = provider::google-gce-metadata::disk_types("n2-standard-64")
}

# Custom machine types resolve too, and the vCPU count is taken from the name.
# An 8 vCPU custom N2 is below the 64 vCPU floor for hyperdisk-extreme.
output "custom_disk_types" {
  value = provider::google-gce-metadata::disk_types("n2-custom-8-16384")
}
