// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/persona-id/terraform-provider-google-gce-metadata/internal/gce"
)

var _ function.Function = &diskTypesFunction{}

// diskTypesFunction implements both disk_types and boot_disk_types, which
// differ only in the disk usage they query.
type diskTypesFunction struct {
	description string
	name        string
	summary     string
	usage       gce.Usage
}

// NewDiskTypesFunction returns every disk type a machine type can use, in any
// position.
func NewDiskTypesFunction() function.Function {
	return &diskTypesFunction{
		description: "Returns the sorted list of disk types the given machine type supports, in any position. " +
			"Some of these are attachable data disks only; use `boot_disk_types` for disk types usable as a boot disk. " +
			"Errors if the machine type is unknown or belongs to a machine series this provider does not model.",
		name:    "disk_types",
		summary: "Disk types supported by a GCE machine type",
		usage:   gce.AnyUsage,
	}
}

// NewBootDiskTypesFunction returns the disk types a machine type can boot from.
func NewBootDiskTypesFunction() function.Function {
	return &diskTypesFunction{
		description: "Returns the sorted list of disk types the given machine type can use as a **boot disk**. " +
			"This is the relevant list for a GKE node pool's `disk_type`, which configures the node boot disk. " +
			"Hyperdisk Extreme, for example, is supported by `n2-standard-64` as a data disk but cannot be booted from. " +
			"Errors if the machine type is unknown or belongs to a machine series this provider does not model.",
		name:    "boot_disk_types",
		summary: "Boot disk types supported by a GCE machine type",
		usage:   gce.BootDisk,
	}
}

func (f *diskTypesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f *diskTypesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		MarkdownDescription: f.description,
		Parameters: []function.Parameter{
			function.StringParameter{
				MarkdownDescription: "GCE machine type name, such as `n2-standard-4` or `n2-custom-8-16384`.",
				Name:                "machine_type",
			},
		},
		Return:  function.ListReturn{ElementType: types.StringType},
		Summary: f.summary,
	}
}

func (f *diskTypesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var machineType string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &machineType))
	if resp.Error != nil {
		return
	}

	supported, err := gce.SupportedDiskTypes(machineType, f.usage)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, supported))
}
