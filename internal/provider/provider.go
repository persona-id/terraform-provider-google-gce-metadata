// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure GoogleGCEMetadata satisfies various provider interfaces.
var (
	_ provider.Provider              = &GoogleGCEMetadata{}
	_ provider.ProviderWithFunctions = &GoogleGCEMetadata{}
)

// GoogleGCEMetadata defines the provider implementation.
type GoogleGCEMetadata struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GoogleGCEMetadata{
			version: version,
		}
	}
}

func (p *GoogleGCEMetadata) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *GoogleGCEMetadata) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *GoogleGCEMetadata) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewBootDiskTypesFunction,
		NewDiskTypesFunction,
		NewIsValidDiskTypeFunction,
		NewIsValidMachineTypeFunction,
		NewMachineTypeInfoFunction,
		NewSupportsBootDiskTypeFunction,
		NewSupportsDiskTypeFunction,
	}
}

func (p *GoogleGCEMetadata) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "google-gce-metadata"
	resp.Version = p.version
}

func (p *GoogleGCEMetadata) Resources(ctx context.Context) []func() resource.Resource {
	return nil
}

func (p *GoogleGCEMetadata) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Static GCE metadata as provider-defined functions: which disk types each machine " +
			"type supports, and which of those it can boot from. The provider makes no API calls and needs no " +
			"credentials - the machine type table is generated from the Compute Engine API at build time and " +
			"committed, and the disk support matrix is transcribed from Google's documentation.",
	}
}
