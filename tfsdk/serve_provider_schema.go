package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func serveGetProviderSchema(ctx context.Context, provider Provider) (*tfprotov6.Schema, diag.Diagnostics) {
	// get the provider schema
	providerSchema, diags := provider.GetSchema(ctx)
	if diags.HasError() {
		return nil, diags
	}
	// convert the provider schema to a *tfprotov6.Schema
	provider6Schema, err := providerSchema.tfprotov6Schema(ctx)
	if err != nil {
		diags.AddError(
			"Error converting provider schema",
			"The provider schema couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}
	return provider6Schema, nil
}

func serveGetProviderMetaSchema(ctx context.Context, provider Provider) (*tfprotov6.Schema, diag.Diagnostics) {
	pm, ok := provider.(ProviderWithProviderMeta)
	if !ok {
		return nil, nil
	}
	pmSchema, diags := pm.GetMetaSchema(ctx)
	if diags.HasError() {
		return nil, diags
	}

	pm6Schema, err := pmSchema.tfprotov6Schema(ctx)
	if err != nil {
		diags.AddError(
			"Error converting provider_meta schema",
			"The provider_meta schema couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}
	return pm6Schema, nil
}

func serveGetResourceSchemas(ctx context.Context, provider Provider) (map[string]*tfprotov6.Schema, diag.Diagnostics) {
	resourceSchemas, diags := provider.GetResources(ctx)
	if diags.HasError() {
		return nil, diags
	}
	resource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range resourceSchemas {
		schema, ds := v.GetSchema(ctx)
		diags.Append(ds...)
		if diags.HasError() {
			return nil, diags
		}
		schema6, err := schema.tfprotov6Schema(ctx)
		if err != nil {
			diags.AddError(
				"Error converting resource schema",
				"The schema for the resource \""+k+"\" couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return nil, diags
		}
		resource6Schemas[k] = schema6
	}
	return resource6Schemas, nil
}

func serveGetDataSourceSchemas(ctx context.Context, provider Provider) (map[string]*tfprotov6.Schema, diag.Diagnostics) {
	dataSourceSchemas, diags := provider.GetDataSources(ctx)
	if diags.HasError() {
		return nil, diags
	}
	dataSource6Schemas := map[string]*tfprotov6.Schema{}
	for k, v := range dataSourceSchemas {
		schema, ds := v.GetSchema(ctx)
		diags.Append(ds...)
		if diags.HasError() {
			return nil, diags
		}
		schema6, err := schema.tfprotov6Schema(ctx)
		if err != nil {
			diags.AddError(
				"Error converting data sourceschema",
				"The schema for the data source \""+k+"\" couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return nil, diags
		}
		dataSource6Schemas[k] = schema6
	}
	return dataSource6Schemas, nil
}
