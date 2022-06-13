package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// GetProviderSchemaResponse returns the *tfprotov5.GetProviderSchemaResponse
// equivalent of a *fwserver.GetProviderSchemaResponse.
func GetProviderSchemaResponse(ctx context.Context, fw *fwserver.GetProviderSchemaResponse) *tfprotov5.GetProviderSchemaResponse {
	if fw == nil {
		return nil
	}

	protov6 := &tfprotov5.GetProviderSchemaResponse{
		DataSourceSchemas: map[string]*tfprotov5.Schema{},
		Diagnostics:       Diagnostics(fw.Diagnostics),
		ResourceSchemas:   map[string]*tfprotov5.Schema{},
	}

	var err error

	protov6.Provider, err = Schema(ctx, fw.Provider)

	if err != nil {
		protov6.Diagnostics = append(protov6.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error converting provider schema",
			Detail:   "The provider schema couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		})

		return protov6
	}

	protov6.ProviderMeta, err = Schema(ctx, fw.ProviderMeta)

	if err != nil {
		protov6.Diagnostics = append(protov6.Diagnostics, &tfprotov5.Diagnostic{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Error converting provider_meta schema",
			Detail:   "The provider_meta schema couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n" + err.Error(),
		})

		return protov6
	}

	for dataSourceType, dataSourceSchema := range fw.DataSourceSchemas {
		protov6.DataSourceSchemas[dataSourceType], err = Schema(ctx, dataSourceSchema)

		if err != nil {
			protov6.Diagnostics = append(protov6.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Error converting data source schema",
				Detail:   "The schema for the data source \"" + dataSourceType + "\" couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			})

			return protov6
		}
	}

	for resourceType, resourceSchema := range fw.ResourceSchemas {
		protov6.ResourceSchemas[resourceType], err = Schema(ctx, resourceSchema)

		if err != nil {
			protov6.Diagnostics = append(protov6.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Error converting resource schema",
				Detail:   "The schema for the resource \"" + resourceType + "\" couldn't be converted into a usable type. This is always a problem with the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			})

			return protov6
		}
	}

	return protov6
}
