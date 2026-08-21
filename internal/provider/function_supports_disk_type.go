// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/persona-id/terraform-provider-google-gce-metadata/internal/gce"
)

var _ function.Function = &supportsDiskTypeFunction{}

// supportsDiskTypeFunction implements both supports_disk_type and
// supports_boot_disk_type, which differ only in the disk usage they query.
type supportsDiskTypeFunction struct {
	description string
	name        string
	summary     string
	usage       gce.Usage
}

// NewSupportsDiskTypeFunction reports whether a machine type can use a disk
// type in any position.
func NewSupportsDiskTypeFunction() function.Function {
	return &supportsDiskTypeFunction{
		description: "Reports whether the given machine type supports the given disk type in any position. " +
			"Use `supports_boot_disk_type` to check boot disk eligibility specifically. " +
			"Returns `true` or `false`, but an unknown machine type or disk type raises an error rather than " +
			"returning `false`, so a typo fails the plan. Use `is_valid_machine_type` or `is_valid_disk_type` " +
			"to test a name without failing.",
		name:    "supports_disk_type",
		summary: "Whether a GCE machine type supports a disk type",
		usage:   gce.AnyUsage,
	}
}

// NewSupportsBootDiskTypeFunction reports whether a machine type can boot from
// a disk type.
func NewSupportsBootDiskTypeFunction() function.Function {
	return &supportsDiskTypeFunction{
		description: "Reports whether the given machine type can use the given disk type as a **boot disk**. " +
			"This is the relevant check for a GKE node pool's `disk_type`. " +
			"Returns `true` or `false`, but an unknown machine type or disk type raises an error rather than " +
			"returning `false`, so a typo fails the plan. Use `is_valid_machine_type` or `is_valid_disk_type` " +
			"to test a name without failing.",
		name:    "supports_boot_disk_type",
		summary: "Whether a GCE machine type can boot from a disk type",
		usage:   gce.BootDisk,
	}
}

func (f *supportsDiskTypeFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f *supportsDiskTypeFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		MarkdownDescription: f.description,
		Parameters: []function.Parameter{
			function.StringParameter{
				MarkdownDescription: "GCE machine type name, such as `n2-standard-4` or `n2-custom-8-16384`.",
				Name:                "machine_type",
			},
			function.StringParameter{
				MarkdownDescription: "Disk type name, such as `hyperdisk-balanced` or `pd-ssd`.",
				Name:                "disk_type",
			},
		},
		Return:  function.BoolReturn{},
		Summary: f.summary,
	}
}

func (f *supportsDiskTypeFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var diskType, machineType string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &machineType, &diskType))
	if resp.Error != nil {
		return
	}

	supported, err := gce.SupportsDiskType(machineType, diskType, f.usage)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, supported))
}
