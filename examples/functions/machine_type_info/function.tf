# vCPU and memory figures come from the Compute Engine API at code generation
# time, not from parsing the machine type name.
# => { custom = false, memory_mb = 16384, name = "n2-standard-4", series = "n2", shape = "standard", shared_core = false, vcpus = 4 }
output "predefined" {
  value = provider::google-gce-metadata::machine_type_info("n2-standard-4")
}

output "custom" {
  value = provider::google-gce-metadata::machine_type_info("n2-custom-8-16384")
}

# Size a node pool's disk from the machine's vCPU count.
output "disk_size_gb" {
  value = provider::google-gce-metadata::machine_type_info(var.machine_type).vcpus * 10
}
