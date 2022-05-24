package proto6server

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/emptyprovider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"
)

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	s := new(testServeProvider)
	testServer := &Server{
		FrameworkServer: fwserver.Server{
			Provider: s,
		},
	}
	got, err := testServer.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
		return
	}
	expected := &tfprotov6.GetProviderSchemaResponse{
		Provider: testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":                           testServeResourceTypeOneSchema,
			"test_two":                           testServeResourceTypeTwoSchema,
			"test_three":                         testServeResourceTypeThreeSchema,
			"test_attribute_plan_modifiers":      testServeResourceTypeAttributePlanModifiersSchema,
			"test_config_validators":             testServeResourceTypeConfigValidatorsSchema,
			"test_import_state":                  testServeResourceTypeImportStateSchema,
			"test_import_state_not_implemented":  testServeResourceTypeImportStateNotImplementedSchema,
			"test_upgrade_state":                 testServeResourceTypeUpgradeStateSchema,
			"test_upgrade_state_empty":           testServeResourceTypeUpgradeStateEmptySchema,
			"test_upgrade_state_not_implemented": testServeResourceTypeUpgradeStateNotImplementedSchema,
			"test_validate_config":               testServeResourceTypeValidateConfigSchema,
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":               testServeDataSourceTypeOneSchema,
			"test_two":               testServeDataSourceTypeTwoSchema,
			"test_config_validators": testServeDataSourceTypeConfigValidatorsSchema,
			"test_validate_config":   testServeDataSourceTypeValidateConfigSchema,
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}

func TestServerGetProviderSchema_logging(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)
	ctx = logging.InitContext(ctx)

	testServer := &Server{
		FrameworkServer: fwserver.Server{
			Provider: &emptyprovider.Provider{},
		},
	}

	_, err := testServer.GetProviderSchema(ctx, new(tfprotov6.GetProviderSchemaRequest))

	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":   "trace",
			"@message": "Checking ProviderSchema lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Calling provider defined Provider GetSchema",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider GetSchema",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ResourceSchemas lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking ResourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Calling provider defined Provider GetResources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider GetResources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking DataSourceSchemas lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "trace",
			"@message": "Checking DataSourceTypes lock",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Calling provider defined Provider GetDataSources",
			"@module":  "sdk.framework",
		},
		{
			"@level":   "debug",
			"@message": "Called provider defined Provider GetDataSources",
			"@module":  "sdk.framework",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}

func TestServerGetProviderSchemaWithProviderMeta(t *testing.T) {
	t.Parallel()

	s := new(testServeProviderWithMetaSchema)
	testServer := &Server{
		FrameworkServer: fwserver.Server{
			Provider: s,
		},
	}
	got, err := testServer.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
		return
	}
	expected := &tfprotov6.GetProviderSchemaResponse{
		Provider: testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":                           testServeResourceTypeOneSchema,
			"test_two":                           testServeResourceTypeTwoSchema,
			"test_three":                         testServeResourceTypeThreeSchema,
			"test_attribute_plan_modifiers":      testServeResourceTypeAttributePlanModifiersSchema,
			"test_config_validators":             testServeResourceTypeConfigValidatorsSchema,
			"test_import_state":                  testServeResourceTypeImportStateSchema,
			"test_import_state_not_implemented":  testServeResourceTypeImportStateNotImplementedSchema,
			"test_upgrade_state":                 testServeResourceTypeUpgradeStateSchema,
			"test_upgrade_state_empty":           testServeResourceTypeUpgradeStateEmptySchema,
			"test_upgrade_state_not_implemented": testServeResourceTypeUpgradeStateNotImplementedSchema,
			"test_validate_config":               testServeResourceTypeValidateConfigSchema,
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			"test_one":               testServeDataSourceTypeOneSchema,
			"test_two":               testServeDataSourceTypeTwoSchema,
			"test_config_validators": testServeDataSourceTypeConfigValidatorsSchema,
			"test_validate_config":   testServeDataSourceTypeValidateConfigSchema,
		},
		ProviderMeta: &tfprotov6.Schema{
			Version: 2,
			Block: &tfprotov6.SchemaBlock{
				Attributes: []*tfprotov6.SchemaAttribute{
					{
						Name:            "foo",
						Required:        true,
						Type:            tftypes.String,
						Description:     "A **string**",
						DescriptionKind: tfprotov6.StringKindMarkdown,
					},
				},
			},
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}
