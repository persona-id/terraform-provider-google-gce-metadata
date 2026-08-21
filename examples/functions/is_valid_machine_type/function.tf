# supports_disk_type and supports_boot_disk_type return booleans too, but they
# raise an error on an unknown machine type, and Terraform cannot recover from a
# function error. This one reports an unknown name as false instead, so a
# configuration can raise the problem in its own words.
variable "machine_type" {
  type = string
}

check "machine_type_is_known" {
  assert {
    condition     = provider::google-gce-metadata::is_valid_machine_type(var.machine_type)
    error_message = "${var.machine_type} is not a machine type this provider models. Accelerator-optimized and memory-optimized series are out of scope."
  }
}

# => false: G2 is an accelerator-optimized series and is not modeled.
output "unmodeled" {
  value = provider::google-gce-metadata::is_valid_machine_type("g2-standard-8")
}
