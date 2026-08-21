// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/persona-id/terraform-provider-google-gce-metadata/internal/gce"
)

var _ function.Function = &machineTypeInfoFunction{}

// machineTypeInfo is the object returned by the machine_type_info function.
type machineTypeInfo struct {
	Custom     types.Bool   `tfsdk:"custom"`
	MemoryMB   types.Int64  `tfsdk:"memory_mb"`
	Name       types.String `tfsdk:"name"`
	Series     types.String `tfsdk:"series"`
	Shape      types.String `tfsdk:"shape"`
	SharedCore types.Bool   `tfsdk:"shared_core"`
	VCPUs      types.Int64  `tfsdk:"vcpus"`
}

// machineTypeInfoAttributes is the object type returned by machine_type_info.
var machineTypeInfoAttributes = map[string]attr.Type{
	"custom":      types.BoolType,
	"memory_mb":   types.Int64Type,
	"name":        types.StringType,
	"series":      types.StringType,
	"shape":       types.StringType,
	"shared_core": types.BoolType,
	"vcpus":       types.Int64Type,
}

type machineTypeInfoFunction struct{}

// NewMachineTypeInfoFunction returns the decomposed definition of a machine type.
func NewMachineTypeInfoFunction() function.Function {
	return &machineTypeInfoFunction{}
}

func (f *machineTypeInfoFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "machine_type_info"
}

func (f *machineTypeInfoFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		MarkdownDescription: "Returns the definition of the given machine type: its series, shape, vCPU count, and memory. " +
			"vCPU and memory figures for predefined machine types come from the Compute Engine API at code generation " +
			"time rather than being derived from the machine type name. " +
			"Errors if the machine type is unknown or belongs to a machine series this provider does not model.",
		Parameters: []function.Parameter{
			function.StringParameter{
				MarkdownDescription: "GCE machine type name, such as `n2-standard-4` or `n2-custom-8-16384`.",
				Name:                "machine_type",
			},
		},
		Return:  function.ObjectReturn{AttributeTypes: machineTypeInfoAttributes},
		Summary: "Series, shape, vCPU, and memory detail for a GCE machine type",
	}
}

func (f *machineTypeInfoFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var machineTypeName string

	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &machineTypeName))
	if resp.Error != nil {
		return
	}

	machineType, err := gce.Lookup(machineTypeName)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewArgumentFuncError(0, err.Error()))

		return
	}

	info := machineTypeInfo{
		Custom:     types.BoolValue(machineType.Custom),
		MemoryMB:   types.Int64Value(int64(machineType.MemoryMB)),
		Name:       types.StringValue(machineType.Name),
		Series:     types.StringValue(machineType.Series),
		Shape:      types.StringValue(machineType.Shape),
		SharedCore: types.BoolValue(machineType.SharedCore),
		VCPUs:      types.Int64Value(int64(machineType.VCPUs)),
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, info))
}
