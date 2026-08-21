// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package gce

// DiskConstraint records the conditions under which a machine series supports a
// disk type. The zero value means the series supports the disk type on every
// machine type in the series.
type DiskConstraint struct {
	// MinVCPUs is the minimum vCPU count a machine type in the series must have
	// to use the disk type, or 0 when the series imposes no minimum.
	MinVCPUs int
}

// Series is the static definition of one GCE machine series.
type Series struct {
	// Disks maps each disk type the series supports to the constraints on its
	// use. A disk type absent from this map is not supported by the series.
	Disks map[string]DiskConstraint

	// SupportsCustom reports whether the series offers custom machine types.
	SupportsCustom bool
}

// seriesDefs is the static disk support matrix, keyed by machine series.
//
// Scope: general-purpose and compute-optimized series. Accelerator-optimized
// (A*, G*), memory-optimized (M*), storage-optimized (Z*), and network- or
// HPC-optimized (C4N, H4D, X4) series are deliberately not modeled; a machine
// type in an unmodeled series is reported as an error rather than as having no
// disk support. Keep the series list in sync with the SERIES_FILTER in
// tools/generate-machine-types.sh.
//
// Sources, per disk type:
//   - https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-balanced
//   - https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-balanced-ha
//   - https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-extreme
//   - https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-ml
//   - https://docs.cloud.google.com/compute/docs/disks/hd-types/hyperdisk-throughput
//   - https://docs.cloud.google.com/compute/docs/disks/extreme-persistent-disk
//   - https://docs.cloud.google.com/compute/docs/disks/persistent-disks
//
// Transcribed 2026-08-20.
//
// REVIEW: the MinVCPUs figures below are transcribed from each disk type's
// "machine series support" section. At least one such figure turned out to gate
// documented performance rather than attachment (see the pd-extreme note on N2
// below), so every minimum here deserves an empirical check before it is
// trusted to reject a configuration.
var seriesDefs = map[string]Series{
	"c2": {
		Disks: map[string]DiskConstraint{
			DiskPDBalanced: {},
			DiskPDSSD:      {},
			DiskPDStandard: {},
		},
	},

	"c2d": {
		Disks: map[string]DiskConstraint{
			DiskPDBalanced: {},
			DiskPDSSD:      {},
			DiskPDStandard: {},
		},
	},

	"c3": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			// "C3 machine types require at least 88 vCPUs."
			DiskHyperdiskExtreme:    {MinVCPUs: 88},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
			DiskPDBalanced:          {},
			DiskPDSSD:               {},
			DiskPDStandard:          {},
		},
	},

	"c3d": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			// "C3D machine types require at least 60 vCPUs."
			DiskHyperdiskExtreme:    {MinVCPUs: 60},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
			DiskPDBalanced:          {},
			DiskPDSSD:               {},
			DiskPDStandard:          {},
		},
	},

	// C4, C4A, C4D, N4, N4A, and N4D are 4th generation series and are
	// Hyperdisk-only: "Persistent Disk isn't available with the latest machine
	// series", and the C4D reference states outright that Persistent Disk is
	// not supported.
	"c4": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			// "C4 and G4 machine types require at least 96 vCPUs."
			DiskHyperdiskExtreme:    {MinVCPUs: 96},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
		},
	},

	"c4a": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			// "C4A, C4D, M3, and M4 machine types require at least 64 vCPUs."
			DiskHyperdiskExtreme:    {MinVCPUs: 64},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
		},
	},

	// C4D is absent from the Hyperdisk Throughput supported series list, unlike
	// its C4 and C4A siblings.
	"c4d": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			DiskHyperdiskExtreme:    {MinVCPUs: 64},
			DiskHyperdiskML:         {},
		},
	},

	"e2": {
		Disks: map[string]DiskConstraint{
			DiskPDBalanced: {},
			DiskPDSSD:      {},
			DiskPDStandard: {},
		},
		SupportsCustom: true,
	},

	"h3": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskThroughput: {},
			DiskPDBalanced:          {},
			DiskPDSSD:               {},
			DiskPDStandard:          {},
		},
	},

	"n1": {
		Disks: map[string]DiskConstraint{
			DiskPDBalanced: {},
			DiskPDSSD:      {},
			DiskPDStandard: {},
		},
		SupportsCustom: true,
	},

	// N2 is the series carrying the constraint this provider was written for.
	"n2": {
		Disks: map[string]DiskConstraint{
			// The Hyperdisk Extreme page states "N2 machine types require at
			// least 80 vCPUs; Custom N2 machine types aren't supported".
			// Testing on 2026-08-20 contradicted both halves: an n2-standard-64
			// and an n2-custom-64-117760 each took a Hyperdisk Extreme disk,
			// while the console refused to offer the combination for an
			// n2-standard-48. N2 has no size between 48 and 64, so the real
			// floor is 64 and there is no custom exclusion.
			DiskHyperdiskExtreme:    {MinVCPUs: 64},
			DiskHyperdiskThroughput: {},
			DiskPDBalanced:          {},
			// No minimum, despite the Extreme Persistent Disk page saying "N2
			// VMs require at least 64 vCPUs": an n2-standard-2 with a
			// pd-extreme disk was created successfully on 2026-08-20. That 64
			// figure gates the documented performance limits, not attachment -
			// the same page notes "N2 VMs with 64 or 80 vCPUs require the Intel
			// Ice Lake CPU platform to reach the stated performance limits".
			DiskPDExtreme:  {},
			DiskPDSSD:      {},
			DiskPDStandard: {},
		},
		SupportsCustom: true,
	},

	"n2d": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskThroughput: {},
			DiskPDBalanced:          {},
			DiskPDSSD:               {},
			DiskPDStandard:          {},
		},
		SupportsCustom: true,
	},

	// N4, N4A, and N4D are absent from the Hyperdisk Extreme supported series
	// list, so they support Hyperdisk Balanced but not Hyperdisk Extreme.
	"n4": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
		},
		SupportsCustom: true,
	},

	"n4a": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
		},
		SupportsCustom: true,
	},

	"n4d": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskBalanced:   {},
			DiskHyperdiskBalancedHA: {},
			DiskHyperdiskML:         {},
			DiskHyperdiskThroughput: {},
		},
		SupportsCustom: true,
	},

	"t2a": {
		Disks: map[string]DiskConstraint{
			DiskPDBalanced: {},
			DiskPDSSD:      {},
			DiskPDStandard: {},
		},
	},

	"t2d": {
		Disks: map[string]DiskConstraint{
			DiskHyperdiskThroughput: {},
			DiskPDBalanced:          {},
			DiskPDSSD:               {},
			DiskPDStandard:          {},
		},
	},
}

// SeriesNames returns every machine series this package models, sorted.
func SeriesNames() []string {
	return sortedKeys(seriesDefs)
}
