// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package gce

import (
	"slices"
	"testing"
)

func TestLookup(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		machineType    string
		wantCustom     bool
		wantErr        bool
		wantSeries     string
		wantSharedCore bool
		wantVCPUs      int
	}{
		"predefined": {
			machineType: "n2-standard-4",
			wantSeries:  "n2",
			wantVCPUs:   4,
		},
		"predefined large": {
			machineType: "n2-standard-64",
			wantSeries:  "n2",
			wantVCPUs:   64,
		},
		"predefined shared core": {
			machineType:    "e2-micro",
			wantSeries:     "e2",
			wantSharedCore: true,
			wantVCPUs:      2,
		},
		"custom series qualified": {
			machineType: "n2-custom-8-16384",
			wantCustom:  true,
			wantSeries:  "n2",
			wantVCPUs:   8,
		},
		"custom with extended memory": {
			machineType: "n2-custom-8-16384-ext",
			wantCustom:  true,
			wantSeries:  "n2",
			wantVCPUs:   8,
		},
		"custom legacy n1 form": {
			machineType: "custom-8-16384",
			wantCustom:  true,
			wantSeries:  "n1",
			wantVCPUs:   8,
		},
		// C4 does not offer custom machine types, so this name is not a valid
		// machine type even though it is a well-formed custom name.
		"custom on series without custom support": {
			machineType: "c4-custom-8-16384",
			wantErr:     true,
		},
		// Accelerator-optimized series are deliberately out of scope. These
		// must error rather than report no disk support.
		"unmodeled accelerator series": {
			machineType: "a2-highgpu-1g",
			wantErr:     true,
		},
		"unmodeled g2 series": {
			machineType: "g2-standard-8",
			wantErr:     true,
		},
		"unmodeled memory optimized series": {
			machineType: "m1-ultramem-40",
			wantErr:     true,
		},
		"nonexistent size in modeled series": {
			machineType: "n2-standard-7",
			wantErr:     true,
		},
		"not a machine type": {
			machineType: "definitely-not-real",
			wantErr:     true,
		},
		"empty": {
			machineType: "",
			wantErr:     true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := Lookup(test.machineType)

			if test.wantErr {
				if err == nil {
					t.Fatalf("Lookup(%q) = %+v, want error", test.machineType, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("Lookup(%q) returned unexpected error: %v", test.machineType, err)
			}

			if got.Custom != test.wantCustom {
				t.Errorf("Custom = %t, want %t", got.Custom, test.wantCustom)
			}

			if got.Series != test.wantSeries {
				t.Errorf("Series = %q, want %q", got.Series, test.wantSeries)
			}

			if got.SharedCore != test.wantSharedCore {
				t.Errorf("SharedCore = %t, want %t", got.SharedCore, test.wantSharedCore)
			}

			if got.VCPUs != test.wantVCPUs {
				t.Errorf("VCPUs = %d, want %d", got.VCPUs, test.wantVCPUs)
			}
		})
	}
}

func TestSupportedDiskTypes(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		machineType string
		usage       Usage
		want        []string
		wantErr     bool
	}{
		// Hyperdisk Extreme on N2 needs 64 vCPUs and nothing else: the
		// documented exclusion of custom N2 machine types was disproven by
		// testing, as was the documented 80 vCPU floor.
		"n2 large enough for hyperdisk extreme": {
			machineType: "n2-standard-64",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskExtreme,
				DiskHyperdiskThroughput,
				DiskPDBalanced,
				DiskPDExtreme,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		"n2 too small for hyperdisk extreme": {
			machineType: "n2-standard-32",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskThroughput,
				DiskPDBalanced,
				DiskPDExtreme,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		"custom n2 large enough for hyperdisk extreme": {
			machineType: "n2-custom-64-117760",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskExtreme,
				DiskHyperdiskThroughput,
				DiskPDBalanced,
				DiskPDExtreme,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		"custom n2 too small for hyperdisk extreme": {
			machineType: "n2-custom-8-16384",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskThroughput,
				DiskPDBalanced,
				DiskPDExtreme,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		// pd-extreme has no vCPU floor: an n2-standard-2 with a pd-extreme disk
		// was created successfully.
		"smallest n2 still supports pd-extreme": {
			machineType: "n2-standard-2",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskThroughput,
				DiskPDBalanced,
				DiskPDExtreme,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		// Hyperdisk Extreme, Hyperdisk Throughput, and pd-extreme are all data
		// disks only, so they drop out of the boot list.
		"n2 boot disks exclude every non bootable type": {
			machineType: "n2-standard-64",
			usage:       BootDisk,
			want: []string{
				DiskPDBalanced,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		"e2 supports persistent disk only": {
			machineType: "e2-micro",
			usage:       AnyUsage,
			want: []string{
				DiskPDBalanced,
				DiskPDSSD,
				DiskPDStandard,
			},
		},
		// 4th generation series are Hyperdisk-only, and c4-standard-4 is below
		// the 96 vCPU floor for Hyperdisk Extreme.
		"c4 is hyperdisk only and below the extreme floor": {
			machineType: "c4-standard-4",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskBalanced,
				DiskHyperdiskBalancedHA,
				DiskHyperdiskML,
				DiskHyperdiskThroughput,
			},
		},
		"n4 is hyperdisk only and has no extreme support": {
			machineType: "n4-standard-4",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskBalanced,
				DiskHyperdiskBalancedHA,
				DiskHyperdiskML,
				DiskHyperdiskThroughput,
			},
		},
		// C4D is absent from the Hyperdisk Throughput supported series list,
		// unlike its C4 and C4A siblings.
		"c4d has no hyperdisk throughput support": {
			machineType: "c4d-standard-4",
			usage:       AnyUsage,
			want: []string{
				DiskHyperdiskBalanced,
				DiskHyperdiskBalancedHA,
				DiskHyperdiskML,
			},
		},
		// Extreme and ML are not bootable, and Throughput is unsupported on
		// C4D, leaving hyperdisk-balanced as the only C4D boot disk.
		"c4d can only boot from hyperdisk balanced": {
			machineType: "c4d-standard-4",
			usage:       BootDisk,
			want:        []string{DiskHyperdiskBalanced},
		},
		"unmodeled series errors": {
			machineType: "g2-standard-8",
			usage:       AnyUsage,
			wantErr:     true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := SupportedDiskTypes(test.machineType, test.usage)

			if test.wantErr {
				if err == nil {
					t.Fatalf("SupportedDiskTypes(%q) = %v, want error", test.machineType, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("SupportedDiskTypes(%q) returned unexpected error: %v", test.machineType, err)
			}

			if !slices.Equal(got, test.want) {
				t.Errorf("SupportedDiskTypes(%q) = %v, want %v", test.machineType, got, test.want)
			}
		})
	}
}

func TestSupportsDiskType(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		diskType    string
		machineType string
		usage       Usage
		want        bool
		wantErr     bool
	}{
		"supported data disk": {
			diskType:    DiskHyperdiskExtreme,
			machineType: "n2-standard-64",
			usage:       AnyUsage,
			want:        true,
		},
		"supported as data disk but not as boot disk": {
			diskType:    DiskHyperdiskExtreme,
			machineType: "n2-standard-64",
			usage:       BootDisk,
			want:        false,
		},
		"below vcpu minimum": {
			diskType:    DiskHyperdiskExtreme,
			machineType: "n2-standard-32",
			usage:       AnyUsage,
			want:        false,
		},
		"custom large enough": {
			diskType:    DiskHyperdiskExtreme,
			machineType: "n2-custom-64-117760",
			usage:       AnyUsage,
			want:        true,
		},
		"custom below vcpu minimum": {
			diskType:    DiskHyperdiskExtreme,
			machineType: "n2-custom-8-16384",
			usage:       AnyUsage,
			want:        false,
		},
		"unknown disk type errors": {
			diskType:    "pd-turbo",
			machineType: "n2-standard-4",
			usage:       AnyUsage,
			wantErr:     true,
		},
		"unknown machine type errors": {
			diskType:    DiskPDBalanced,
			machineType: "n2-standard-7",
			usage:       AnyUsage,
			wantErr:     true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := SupportsDiskType(test.machineType, test.diskType, test.usage)

			if test.wantErr {
				if err == nil {
					t.Fatalf("SupportsDiskType(%q, %q) = %t, want error", test.machineType, test.diskType, got)
				}

				return
			}

			if err != nil {
				t.Fatalf("SupportsDiskType(%q, %q) returned unexpected error: %v", test.machineType, test.diskType, err)
			}

			if got != test.want {
				t.Errorf("SupportsDiskType(%q, %q) = %t, want %t", test.machineType, test.diskType, got, test.want)
			}
		})
	}
}

// TestSupportedDiskTypesIsSorted asserts the sort contract across the whole
// generated table, which the per-case exact lists above rely on.
func TestSupportedDiskTypesIsSorted(t *testing.T) {
	t.Parallel()

	for name := range machineTypes {
		for _, usage := range []Usage{AnyUsage, BootDisk} {
			got, err := SupportedDiskTypes(name, usage)
			if err != nil {
				t.Fatalf("SupportedDiskTypes(%q) returned unexpected error: %v", name, err)
			}

			if !slices.IsSorted(got) {
				t.Errorf("SupportedDiskTypes(%q, %v) = %v, want sorted", name, usage, got)
			}
		}
	}
}

// TestSeriesTableMatchesGeneratedTable guards against drift between the
// hand-written series table and the generated machine type table: a series
// added to one and not the other silently produces wrong answers.
func TestSeriesTableMatchesGeneratedTable(t *testing.T) {
	t.Parallel()

	generatedSeries := make(map[string]bool)

	for _, machineType := range machineTypes {
		generatedSeries[machineType.Series] = true
	}

	for series := range seriesDefs {
		if !generatedSeries[series] {
			t.Errorf("series %q is defined in seriesDefs but no generated machine type belongs to it", series)
		}
	}

	for series := range generatedSeries {
		if _, ok := seriesDefs[series]; !ok {
			t.Errorf("generated machine types belong to series %q, which has no seriesDefs entry", series)
		}
	}
}

// TestSeriesReferenceKnownDiskTypes catches a disk type misspelled in the
// series table, which would otherwise read as "not supported".
func TestSeriesReferenceKnownDiskTypes(t *testing.T) {
	t.Parallel()

	for series, definition := range seriesDefs {
		for diskType := range definition.Disks {
			if _, ok := diskTypes[diskType]; !ok {
				t.Errorf("series %q references disk type %q, which is not defined in diskTypes", series, diskType)
			}
		}
	}
}
