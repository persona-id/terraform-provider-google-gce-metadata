// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/persona-id/terraform-provider-google-gce-metadata/internal/gce"
)

var _ function.Function = &isValidFunction{}

// isValidFunction implements both is_valid_machine_type and is_valid_disk_type.
//
// What sets them apart is not the boolean return, which the supports_disk_type
// and supports_boot_disk_type functions share, but how they treat an
// unrecognized name: the others reject it with an error so a typo fails the
// plan, while these report it as false. A Terraform configuration cannot recover
// from a function error, so validating a name in a check block or a resource
// precondition needs an answer rather than a failure.
type isValidFunction struct {
	description string
	name        string
	parameter   string
	summary     string
	valid       func(string) bool
}

// NewIsValidDiskTypeFunction reports whether a disk type name is one this
// provider models.
func NewIsValidDiskTypeFunction() function.Function {
	return &isValidFunction{
		description: "Reports whether the given name is a disk type this provider models. " +
			"An unrecognized name is reported as `false` rather than raising an error, which is what " +
			"distinguishes this from the `supports_*` functions, so it can be used in a `check` block or a " +
			"resource precondition to report the problem in your own words.",
		name:      "is_valid_disk_type",
		parameter: "disk_type",
		summary:   "Whether a name is a known disk type",
		valid:     gce.IsValidDiskType,
	}
}

// NewIsValidMachineTypeFunction reports whether a machine type name is one this
// provider models.
func NewIsValidMachineTypeFunction() function.Function {
	return &isValidFunction{
		description: "Reports whether the given name is a machine type this provider models, covering both " +
			"predefined and custom machine types. An unrecognized name is reported as `false` rather than " +
			"raising an error, which is what distinguishes this from the `supports_*` functions, so it can be " +
			"used in a `check` block or a resource precondition to report the problem in your own words. " +
			"Note that `false` also covers machine types in series this provider does not model, such as " +
			"accelerator-optimized ones; it does not mean the machine type does not exist in GCE.",
		name:      "is_valid_machine_type",
		parameter: "machine_type",
		summary:   "Whether a name is a known GCE machine type",
		valid:     gce.IsValidMachineType,
	}
}

func (f *isValidFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = f.name
}

func (f *isValidFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		MarkdownDescription: f.description,
		Parameters: []function.Parameter{
			function.StringParameter{
				MarkdownDescription: "Name to test.",
				Name:                f.parameter,
			},
		},
		Return:  function.BoolReturn{},
		Summary: f.summary,
	}
}

func (f *isValidFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &name))
	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, f.valid(name)))
}
