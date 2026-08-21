variable "disk_type" {
  type = string
}

check "disk_type_is_known" {
  assert {
    condition     = provider::google-gce-metadata::is_valid_disk_type(var.disk_type)
    error_message = "${var.disk_type} is not a known GCE disk type."
  }
}

# => false
output "unknown" {
  value = provider::google-gce-metadata::is_valid_disk_type("pd-turbo")
}
