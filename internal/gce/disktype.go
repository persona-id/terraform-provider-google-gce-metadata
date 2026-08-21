// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package gce

// Disk type names, as accepted by the Compute Engine API and the
// google_compute_disk / google_container_node_pool Terraform resources.
const (
	DiskHyperdiskBalanced   = "hyperdisk-balanced"
	DiskHyperdiskBalancedHA = "hyperdisk-balanced-high-availability"
	DiskHyperdiskExtreme    = "hyperdisk-extreme"
	DiskHyperdiskML         = "hyperdisk-ml"
	DiskHyperdiskThroughput = "hyperdisk-throughput"
	DiskPDBalanced          = "pd-balanced"
	DiskPDExtreme           = "pd-extreme"
	DiskPDSSD               = "pd-ssd"
	DiskPDStandard          = "pd-standard"
)

// DiskType records the properties of a disk type that hold regardless of which
// machine series the disk is attached to.
type DiskType struct {
	// Bootable reports whether the disk type can be used as a boot disk. This
	// is the distinction that matters for GKE node pools, whose disk_type
	// configures the node boot disk rather than an attached data disk.
	Bootable bool
}

// diskTypes records the series-independent properties of each disk type.
//
// Boot disk support is a property of the disk type itself, not of the machine
// series: Hyperdisk Extreme, for example, is never usable as a boot disk on any
// series that otherwise supports it.
//
// REVIEW: only Hyperdisk Extreme carries an explicit "can't be used as a boot
// disk" statement in the documentation. Hyperdisk Balanced High Availability,
// Hyperdisk ML, Hyperdisk Throughput, and Extreme Persistent Disk are recorded
// as data disks only because their reference pages describe them exclusively as
// attached volumes, not because a prohibition is documented. Confirm those four
// before relying on them.
//
// See: https://docs.cloud.google.com/compute/docs/disks
var diskTypes = map[string]DiskType{
	// Hyperdisk Balanced lists boot disks as a supported use case.
	// https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-balanced
	DiskHyperdiskBalanced: {Bootable: true},

	DiskHyperdiskBalancedHA: {Bootable: false},

	// "You can't use a Hyperdisk Extreme volume as a boot disk."
	// https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-extreme
	DiskHyperdiskExtreme: {Bootable: false},

	DiskHyperdiskML:         {Bootable: false},
	DiskHyperdiskThroughput: {Bootable: false},
	DiskPDBalanced:          {Bootable: true},
	DiskPDExtreme:           {Bootable: false},
	DiskPDSSD:               {Bootable: true},
	DiskPDStandard:          {Bootable: true},
}

// DiskTypeNames returns every disk type this package models, sorted.
func DiskTypeNames() []string {
	return sortedKeys(diskTypes)
}
