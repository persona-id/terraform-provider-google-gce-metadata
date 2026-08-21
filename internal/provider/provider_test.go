// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/persona-id/terraform-provider-google-gce-metadata/internal/provider"
)

// testAccProtoV6ProviderFactories registers the provider under the local name
// "gce", so test configurations call provider::gce::<function>.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"gce": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// providerFunctionsRequire is the minimum Terraform version with support for
// provider-defined functions.
func providerFunctionsRequire() []tfversion.TerraformVersionCheck {
	return []tfversion.TerraformVersionCheck{
		tfversion.SkipBelow(tfversion.Version1_8_0),
	}
}

func TestDiskTypesFunction(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config: `output "test" { value = provider::gce::disk_types("n2-standard-64") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("hyperdisk-extreme"),
						knownvalue.StringExact("hyperdisk-throughput"),
						knownvalue.StringExact("pd-balanced"),
						knownvalue.StringExact("pd-extreme"),
						knownvalue.StringExact("pd-ssd"),
						knownvalue.StringExact("pd-standard"),
					})),
				},
			},
		},
	})
}

// TestBootDiskTypesFunction covers the boot/data distinction: n2-standard-64
// supports Hyperdisk Extreme as a data disk, but cannot boot from it.
func TestBootDiskTypesFunction(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config: `output "test" { value = provider::gce::boot_disk_types("n2-standard-64") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("pd-balanced"),
						knownvalue.StringExact("pd-ssd"),
						knownvalue.StringExact("pd-standard"),
					})),
				},
			},
		},
	})
}

func TestSupportsDiskTypeFunction(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config: `
					output "custom_large_enough" {
					  value = provider::gce::supports_disk_type("n2-custom-64-117760", "hyperdisk-extreme")
					}
					output "custom_too_small" {
					  value = provider::gce::supports_disk_type("n2-custom-8-16384", "hyperdisk-extreme")
					}
					output "not_bootable" {
					  value = provider::gce::supports_boot_disk_type("n2-standard-64", "hyperdisk-extreme")
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("custom_large_enough", knownvalue.Bool(true)),
					statecheck.ExpectKnownOutputValue("custom_too_small", knownvalue.Bool(false)),
					statecheck.ExpectKnownOutputValue("not_bootable", knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestMachineTypeInfoFunction(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config: `output "test" { value = provider::gce::machine_type_info("n2-custom-8-16384") }`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"custom":      knownvalue.Bool(true),
						"memory_mb":   knownvalue.Int64Exact(16384),
						"name":        knownvalue.StringExact("n2-custom-8-16384"),
						"series":      knownvalue.StringExact("n2"),
						"shape":       knownvalue.StringExact("custom"),
						"shared_core": knownvalue.Bool(false),
						"vcpus":       knownvalue.Int64Exact(8),
					})),
				},
			},
		},
	})
}

func TestIsValidFunctions(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config: `
					output "modeled"       { value = provider::gce::is_valid_machine_type("n2-custom-8-16384") }
					output "unmodeled"     { value = provider::gce::is_valid_machine_type("g2-standard-8") }
					output "nonsense"      { value = provider::gce::is_valid_machine_type("not-a-machine") }
					output "known_disk"    { value = provider::gce::is_valid_disk_type("pd-ssd") }
					output "unknown_disk"  { value = provider::gce::is_valid_disk_type("pd-turbo") }
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("modeled", knownvalue.Bool(true)),
					statecheck.ExpectKnownOutputValue("unmodeled", knownvalue.Bool(false)),
					statecheck.ExpectKnownOutputValue("nonsense", knownvalue.Bool(false)),
					statecheck.ExpectKnownOutputValue("known_disk", knownvalue.Bool(true)),
					statecheck.ExpectKnownOutputValue("unknown_disk", knownvalue.Bool(false)),
				},
			},
		},
	})
}

// TestUnknownMachineTypeErrors asserts an unmodeled machine series fails the
// plan rather than reporting an empty disk type list.
func TestUnknownMachineTypeErrors(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config:      `output "test" { value = provider::gce::disk_types("g2-standard-8") }`,
				ExpectError: regexp.MustCompile(`does not model`),
			},
		},
	})
}

func TestUnknownDiskTypeErrors(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks:   providerFunctionsRequire(),
		Steps: []resource.TestStep{
			{
				Config:      `output "test" { value = provider::gce::supports_disk_type("n2-standard-4", "pd-turbo") }`,
				ExpectError: regexp.MustCompile(`not a known disk type`),
			},
		},
	})
}
