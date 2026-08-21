// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package gce

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Usage is how a disk is attached to an instance. A disk type can be supported
// as a data disk but not as a boot disk, so every support query has to say
// which it means.
type Usage int

const (
	// AnyUsage matches disk types the machine type can use in any position.
	AnyUsage Usage = iota

	// BootDisk matches only disk types usable as the instance's boot disk.
	// This is the relevant usage for a GKE node pool's disk_type.
	BootDisk
)

// SupportedDiskTypes returns the disk types the named machine type can use,
// sorted. The result is empty only for a machine type whose series genuinely
// supports nothing in the requested usage; unmodeled and invalid machine types
// return an error instead.
func SupportedDiskTypes(machineTypeName string, usage Usage) ([]string, error) {
	machineType, err := Lookup(machineTypeName)
	if err != nil {
		return nil, err
	}

	definition, modeled := seriesDefs[machineType.Series]
	if !modeled {
		return nil, fmt.Errorf(
			"machine type %q belongs to machine series %q, which this provider does not model",
			machineTypeName, machineType.Series,
		)
	}

	supported := make([]string, 0, len(definition.Disks))

	for diskType, constraint := range definition.Disks {
		if !constraint.permits(machineType) {
			continue
		}

		if usage == BootDisk && !diskTypes[diskType].Bootable {
			continue
		}

		supported = append(supported, diskType)
	}

	slices.Sort(supported)

	return supported, nil
}

// SupportsDiskType reports whether the named machine type can use the named
// disk type in the requested usage.
func SupportsDiskType(machineTypeName, diskTypeName string, usage Usage) (bool, error) {
	if _, known := diskTypes[diskTypeName]; !known {
		return false, fmt.Errorf(
			"%q is not a known disk type (known disk types: %s)",
			diskTypeName, strings.Join(DiskTypeNames(), ", "),
		)
	}

	supported, err := SupportedDiskTypes(machineTypeName, usage)
	if err != nil {
		return false, err
	}

	return slices.Contains(supported, diskTypeName), nil
}

// IsValidDiskType reports whether the name is a disk type this package models.
func IsValidDiskType(diskTypeName string) bool {
	_, known := diskTypes[diskTypeName]

	return known
}

// IsValidMachineType reports whether the name is a machine type this package
// can resolve, covering both predefined and custom machine types.
//
// This is the predicate form of Lookup, for callers that need to test validity
// without failing: Terraform cannot recover from a function error, so a
// configuration that wants to report an invalid machine type itself - in a check
// block or a resource precondition - needs a boolean rather than an error.
func IsValidMachineType(machineTypeName string) bool {
	_, err := Lookup(machineTypeName)

	return err == nil
}

// permits reports whether a machine type satisfies the constraint.
func (c DiskConstraint) permits(machineType MachineType) bool {
	return machineType.VCPUs >= c.MinVCPUs
}

// sortedKeys returns the keys of m in sorted order.
func sortedKeys[V any](m map[string]V) []string {
	return slices.Sorted(maps.Keys(m))
}
