# An n2-standard-64 has enough vCPUs for Hyperdisk Extreme.
# => true
output "large_enough" {
  value = provider::google-gce-metadata::supports_disk_type("n2-standard-64", "hyperdisk-extreme")
}

# An n2-standard-32 does not.
# => false
output "too_small" {
  value = provider::google-gce-metadata::supports_disk_type("n2-standard-32", "hyperdisk-extreme")
}
