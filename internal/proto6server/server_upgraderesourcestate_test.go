package proto6server

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerUpgradeResourceState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema, _ := testServeResourceTypeUpgradeState{}.GetSchema(ctx)
	schemaType := schema.TerraformType(ctx)

	testCases := map[string]struct {
		request          *tfprotov6.UpgradeResourceStateRequest
		expectedResponse *tfprotov6.UpgradeResourceStateResponse
		expectedError    error
	}{
		"nil": {
			request:          nil,
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{},
		},
		"RawState-missing": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				TypeName: "test_upgrade_state_not_implemented",
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{},
		},
		"TypeName-missing": {
			request: &tfprotov6.UpgradeResourceStateRequest{},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"\" was found in the provider.",
					},
				},
			},
		},
		"TypeName-UpgradeState-not-implemented": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state_not_implemented",
				Version:  0,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Upgrade Resource State",
						Detail: "This resource was implemented without an UpgradeState() method, " +
							"however Terraform was expecting an implementation for version 0 upgrade.\n\n" +
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					},
				},
			},
		},
		"TypeName-UpgradeState-empty": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state_empty",
				Version:  0,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Upgrade Resource State",
						Detail: "This resource was implemented with an UpgradeState() method, " +
							"however Terraform was expecting an implementation for version 0 upgrade.\n\n" +
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					},
				},
			},
		},
		"TypeName-unknown": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				TypeName: "unknown",
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Resource Type Not Found",
						Detail:   "No resource type named \"unknown\" was found in the provider.",
					},
				},
			},
		},
		"Version-0": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state",
				Version:  0,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"Version-1": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state",
				Version:  1,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"Version-2": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state",
				Version:  2,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"Version-3": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state",
				Version:  3,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Read Previously Saved State for UpgradeResourceState",
						Detail: "There was an error reading the saved resource state using the prior resource schema defined for version 3 upgrade.\n\n" +
							"Please report this to the provider developer:\n\n" +
							"AttributeName(\"required_attribute\"): unsupported type bool sent as tftypes.Number",
					},
				},
			},
		},
		"Version-4": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state",
				Version:  4,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Missing Upgraded Resource State",
						Detail: "After attempting a resource state upgrade to version 4, the provider did not return any state data. " +
							"Preventing the unexpected loss of resource state data. " +
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					},
				},
			},
		},
		"Version-current-flatmap": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: &tfprotov6.RawState{
					Flatmap: map[string]string{
						"flatmap": "is not supported",
					},
				},
				TypeName: "test_upgrade_state_not_implemented", // Framework should allow non-ResourceWithUpgradeState
				Version:  1,                                    // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Read Previously Saved State for UpgradeResourceState",
						Detail: "There was an error reading the saved resource state using the current resource schema.\n\n" +
							"If this resource state was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. " +
							"If you manually modified the resource state, you will need to manually modify it to match the current resource schema. " +
							"Otherwise, please report this to the provider developer:\n\n" +
							"flatmap states cannot be unmarshaled, only states written by Terraform 0.12 and higher can be unmarshaled",
					},
				},
			},
		},
		"Version-current-json-match": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": "true",
				}),
				TypeName: "test_upgrade_state_not_implemented", // Framework should allow non-ResourceWithUpgradeState
				Version:  1,                                    // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				UpgradedState: testNewDynamicValue(t, schemaType, map[string]tftypes.Value{
					"id":                 tftypes.NewValue(tftypes.String, "test-id-value"),
					"optional_attribute": tftypes.NewValue(tftypes.String, nil),
					"required_attribute": tftypes.NewValue(tftypes.String, "true"),
				}),
			},
		},
		"Version-current-json-mismatch": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: &tfprotov6.RawState{
					JSON: []byte(`{"nonexistent_attribute":"value"}`),
				},
				TypeName: "test_upgrade_state_not_implemented", // Framework should allow non-ResourceWithUpgradeState
				Version:  1,                                    // Must match current tfsdk.Schema version to trigger framework implementation
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Read Previously Saved State for UpgradeResourceState",
						Detail: "There was an error reading the saved resource state using the current resource schema.\n\n" +
							"If this resource state was last refreshed with Terraform CLI 0.11 and earlier, it must be refreshed or applied with an older provider version first. " +
							"If you manually modified the resource state, you will need to manually modify it to match the current resource schema. " +
							"Otherwise, please report this to the provider developer:\n\n" +
							"ElementKeyValue(tftypes.String<unknown>): unsupported attribute \"nonexistent_attribute\"",
					},
				},
			},
		},
		"Version-not-implemented": {
			request: &tfprotov6.UpgradeResourceStateRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"id":                 "test-id-value",
					"required_attribute": true,
				}),
				TypeName: "test_upgrade_state",
				Version:  999,
			},
			expectedResponse: &tfprotov6.UpgradeResourceStateResponse{
				Diagnostics: []*tfprotov6.Diagnostic{
					{
						Severity: tfprotov6.DiagnosticSeverityError,
						Summary:  "Unable to Upgrade Resource State",
						Detail: "This resource was implemented with an UpgradeState() method, " +
							"however Terraform was expecting an implementation for version 999 upgrade.\n\n" +
							"This is always an issue with the Terraform Provider and should be reported to the provider developer.",
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			testProvider := &testServeProvider{}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: testProvider,
				},
			}

			got, err := testServer.UpgradeResourceState(ctx, testCase.request)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if testCase.request != nil && testCase.request.TypeName == "test_upgrade_state" && testProvider.upgradeResourceStateCalledResourceType != testCase.request.TypeName {
				t.Errorf("expected to call resource %q, called: %s", testCase.request.TypeName, testProvider.upgradeResourceStateCalledResourceType)
				return
			}

			if diff := cmp.Diff(got, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference in response: %s", diff)
			}
		})
	}
}
