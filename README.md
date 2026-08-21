# Google GCE Metadata Terraform Provider

Static Compute Engine metadata exposed as Terraform provider-defined functions: which disk types each machine type supports, and which of those it can boot from. User-facing documentation lives on the [Terraform Registry](https://registry.terraform.io/providers/persona-id/google-gce-metadata/latest/docs); this README covers developing the provider.

- **Supported:** general-purpose (C3, C3D, C4, C4A, C4D, E2, N1, N2, N2D, N4, N4A, N4D, T2A, T2D) and compute-optimized (C2, C2D, H3).
- **Unsupported:** accelerator-optimized (A\*, G\*), memory-optimized (M\*), storage-optimized (Z\*), and network- or HPC-optimized (C4N, H4D, X4).
- **Custom machine types:** resolved, including the legacy N1 `custom-8-16384` form and extended-memory `-ext` variants.
- **No credentials:** the provider makes no API calls. Both metadata tables are compiled in.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.8
- [Go](https://golang.org/doc/install) >= 1.26

## Building the Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the Provider

| Function | Returns | Purpose |
| --- | --- | --- |
| `boot_disk_types(machine_type)` | `list(string)` | Disk types usable as the machine type's **boot** disk |
| `disk_types(machine_type)` | `list(string)` | Disk types the machine type supports in any position |
| `is_valid_disk_type(disk_type)` | `bool` | Whether the name is a disk type this provider models |
| `is_valid_machine_type(machine_type)` | `bool` | Whether the name is a machine type this provider models |
| `machine_type_info(machine_type)` | `object` | Series, shape, vCPU count, memory, and custom/shared-core flags |
| `supports_boot_disk_type(machine_type, disk_type)` | `bool` | Whether the machine type can boot from the disk type |
| `supports_disk_type(machine_type, disk_type)` | `bool` | Whether the machine type supports the disk type at all |

Behaviours worth knowing before adding a function:

- **Boot and data disks are distinct.** A disk type can be attachable to a machine type while being unusable as its boot disk; Hyperdisk Extreme is the clearest case. A GKE node pool's `disk_type` is a boot disk, so the `*boot*` variants exist for that. New functions that answer a support question should take a `gce.Usage`.
- **Unrecognized names raise errors, except in `is_valid_*`.** On an unknown machine type or disk type, functions except `is_valid_*` will raise an error, so a typo fails the plan instead of quietly reading as `false`. Terraform cannot recover from a function error, so the `is_valid_*` functions report an unknown name as `false`, which is the only way a configuration can raise a bad name in its own words.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`. Function documentation is assembled from the schema plus the matching example under `examples/functions/`, so a new function needs an example file to get a docs page.

Unit tests over the metadata tables need no Terraform CLI:

```shell
go test ./internal/gce/
```

The acceptance tests drive a real Terraform binary against the provider's functions. As a static provider, they create no infrastructure and cost nothing to run:

```shell
make testacc
```

### Updating Metadata

There are two tables, kept deliberately separate.

`internal/gce/machinetypes_gen.go` holds every predefined machine type with its vCPU count and memory. It is generated, not hand-written, because machine type names are not a reliable grammar: the trailing number in `g2-standard-8` is a vCPU count while the trailing number in `a2-highgpu-1g` is a GPU count, so deriving these figures from the name is wrong for some series. Regenerate it with:

```shell
GCE_METADATA_GEN_PROJECT=my-project make generate-machine-types
```

Any project you can read machine types from will do, because machine type definitions are global, so the choice of project does not affect the output.

`internal/gce/series.go` holds the disk support matrix, hand-written and cited per entry. Where Google's documentation and observed behaviour disagree, this table follows observed behaviour and records the contradiction in a comment; the instance creation console is treated as the most current source of truth for attach rules. Some `MinVCPUs` figures are still only documented, not verified, and carry a `REVIEW` comment.

`TestSeriesTableMatchesGeneratedTable` fails if a series appears in one table and not the other, so the two cannot drift apart silently.
