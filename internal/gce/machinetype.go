// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package gce

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MachineType is the definition of a single GCE machine type.
type MachineType struct {
	// Custom reports whether this is a custom machine type.
	Custom bool

	// MemoryMB is the machine type's memory in MiB.
	MemoryMB int

	// Name is the machine type name, such as "n2-standard-4".
	Name string

	// Series is the machine series, such as "n2".
	Series string

	// Shape is the predefined shape, such as "standard", "highmem", or
	// "custom". Shared-core machine types record their size ("micro", "small",
	// "medium") as the shape.
	Shape string

	// SharedCore reports whether the machine type has a fractional vCPU
	// allocation, in which case VCPUs is the burstable maximum.
	SharedCore bool

	// VCPUs is the machine type's vCPU count.
	VCPUs int
}

// reCustomMachineType matches the custom machine type name forms:
//
//	n2-custom-8-16384      series-qualified
//	n2-custom-8-16384-ext  with extended memory
//	custom-8-16384         legacy N1 form, which omits the series
//
// This is the only place in the package where a machine type name is
// interpreted rather than looked up in a table. Custom machine types are an
// unbounded name space, so they cannot be enumerated the way predefined machine
// types are. Unlike predefined names, this form is unambiguous: the two numbers
// are always the vCPU count and the memory in MiB.
var reCustomMachineType = regexp.MustCompile(`^(?:([a-z][a-z0-9]*)-)?custom-(\d+)-(\d+)(?:-ext)?$`)

// Lookup returns the definition of a machine type by name, resolving both
// predefined machine types and custom machine types.
//
// Machine types in series this package does not model, and names that are not
// valid machine types at all, return an error rather than a zero value, so that
// callers never mistake "not modeled" for "supports no disk types".
func Lookup(name string) (MachineType, error) {
	if machineType, ok := machineTypes[name]; ok {
		return machineType, nil
	}

	if match := reCustomMachineType.FindStringSubmatch(name); match != nil {
		return customMachineType(name, match)
	}

	// The series is recovered here only to produce a useful error message.
	if series, _, found := strings.Cut(name, "-"); found {
		if _, modeled := seriesDefs[series]; !modeled {
			return MachineType{}, fmt.Errorf(
				"machine type %q belongs to machine series %q, which this provider does not model (modeled series: %s)",
				name, series, strings.Join(SeriesNames(), ", "),
			)
		}
	}

	return MachineType{}, fmt.Errorf("%q is not a known GCE machine type", name)
}

// customMachineType builds a MachineType from a matched custom machine type name.
func customMachineType(name string, match []string) (MachineType, error) {
	series := match[1]
	if series == "" {
		// The legacy "custom-8-16384" form, with no series, is always N1.
		series = "n1"
	}

	definition, modeled := seriesDefs[series]
	if !modeled {
		return MachineType{}, fmt.Errorf(
			"machine type %q belongs to machine series %q, which this provider does not model (modeled series: %s)",
			name, series, strings.Join(SeriesNames(), ", "),
		)
	}

	if !definition.SupportsCustom {
		return MachineType{}, fmt.Errorf(
			"machine series %q does not offer custom machine types, so %q is not a valid machine type",
			series, name,
		)
	}

	vcpus, err := strconv.Atoi(match[2])
	if err != nil {
		return MachineType{}, fmt.Errorf("machine type %q has an unparseable vCPU count %q: %w", name, match[2], err)
	}

	memoryMB, err := strconv.Atoi(match[3])
	if err != nil {
		return MachineType{}, fmt.Errorf("machine type %q has an unparseable memory size %q: %w", name, match[3], err)
	}

	return MachineType{
		Custom:   true,
		MemoryMB: memoryMB,
		Name:     name,
		Series:   series,
		Shape:    "custom",
		VCPUs:    vcpus,
	}, nil
}

// MachineTypeNames returns every predefined machine type this package models, sorted.
func MachineTypeNames() []string {
	return sortedKeys(machineTypes)
}
