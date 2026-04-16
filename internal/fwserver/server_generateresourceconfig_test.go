// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestServerGenerateResourceConfig(t *testing.T) {
	t.Parallel()

	testNestedBlockType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_nested_block_attr": tftypes.String,
		},
	}

	testType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"test_computed":     tftypes.String,
			"test_optional":     tftypes.String,
			"test_required":     tftypes.String,
			"test_deprecated":   tftypes.List{ElementType: tftypes.String},
			"test_false_bool":   tftypes.Bool,
			"test_empty_string": tftypes.String,
			"test_deprecated_block": tftypes.List{
				ElementType: testNestedBlockType,
			},
		},
	}

	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"test_computed": schema.StringAttribute{
				Computed: true,
			},
			"test_optional": schema.StringAttribute{
				Optional: true,
			},
			"test_required": schema.StringAttribute{
				Required: true,
			},
			"test_deprecated": schema.ListAttribute{
				ElementType:        types.StringType,
				Optional:           true,
				DeprecationMessage: "deprecated",
			},
			"test_false_bool": schema.BoolAttribute{
				Optional: true,
			},
			"test_empty_string": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"test_deprecated_block": schema.ListNestedBlock{
				DeprecationMessage: "deprecated",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"test_nested_block_attr": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}

	validatorGroupType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"alpha": tftypes.String,
			"beta":  tftypes.String,
		},
	}

	conflictsWithSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alpha": schema.StringAttribute{
				Optional: true,
				Validators: []schemavalidator.String{
					testConflictsWithStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
				},
			},
			"beta": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	exactlyOneOfSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alpha": schema.StringAttribute{
				Optional: true,
				Validators: []schemavalidator.String{
					testExactlyOneOfStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
				},
			},
			"beta": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	alsoRequiresSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"alpha": schema.StringAttribute{
				Optional: true,
				Validators: []schemavalidator.String{
					testAlsoRequiresStringValidator{paths: path.Expressions{path.MatchRoot("beta")}},
				},
			},
			"beta": schema.StringAttribute{
				Optional: true,
			},
		},
	}

	timeoutsType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"create": tftypes.String,
		},
	}

	timeoutsAndNumberType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"test_optional_number": tftypes.Number,
			"test_nonzero_number":  tftypes.Number,
			"timeouts":             timeoutsType,
		},
	}

	timeoutsAndNumberSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"test_optional_number": schema.NumberAttribute{
				Optional: true,
			},
			"test_nonzero_number": schema.NumberAttribute{
				Optional: true,
			},
			"timeouts": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"create": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		server               *fwserver.Server
		request              *fwserver.GenerateResourceConfigRequest
		expectedResponse     *fwserver.GenerateResourceConfigResponse
		configureProviderReq *provider.ConfigureRequest
	}{
		"nil": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{},
		},
		"request-state-missing": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unexpected Generate Config Request",
						"An unexpected error was encountered when generating resource configuration. The current state was missing.\n\n"+
							"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
					),
				},
			},
		},
		"response-default-config": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: testSchema,
				State: &tfsdk.State{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"id":            tftypes.NewValue(tftypes.String, "test-id-val"),
						"test_computed": tftypes.NewValue(tftypes.String, "test-computed-val"),
						"test_optional": tftypes.NewValue(tftypes.String, "test-optional-val"),
						"test_required": tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_deprecated": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "test-deprecated-a"),
							tftypes.NewValue(tftypes.String, "test-deprecated-b"),
						}),
						"test_false_bool":   tftypes.NewValue(tftypes.Bool, false),
						"test_empty_string": tftypes.NewValue(tftypes.String, ""),
						"test_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: testNestedBlockType}, []tftypes.Value{
							tftypes.NewValue(testNestedBlockType, map[string]tftypes.Value{
								"test_nested_block_attr": tftypes.NewValue(tftypes.String, "test-nested-block-val-a"),
							}),
							tftypes.NewValue(testNestedBlockType, map[string]tftypes.Value{
								"test_nested_block_attr": tftypes.NewValue(tftypes.String, "test-nested-block-val-b"),
							}),
						}),
					}),
					Schema: testSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(testType, map[string]tftypes.Value{
						"id":                    tftypes.NewValue(tftypes.String, nil),
						"test_computed":         tftypes.NewValue(tftypes.String, nil),
						"test_optional":         tftypes.NewValue(tftypes.String, "test-optional-val"),
						"test_required":         tftypes.NewValue(tftypes.String, "test-config-value"),
						"test_deprecated":       tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
						"test_false_bool":       tftypes.NewValue(tftypes.Bool, false),
						"test_empty_string":     tftypes.NewValue(tftypes.String, nil),
						"test_deprecated_block": tftypes.NewValue(tftypes.List{ElementType: testNestedBlockType}, nil),
					}),
					Schema: testSchema,
				},
			},
		},
		"response-conflicts-with-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: conflictsWithSchema,
				State: &tfsdk.State{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: conflictsWithSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: conflictsWithSchema,
				},
			},
		},
		"response-exactly-one-of-group-all-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: exactlyOneOfSchema,
				State: &tfsdk.State{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: exactlyOneOfSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: exactlyOneOfSchema,
				},
			},
		},
		"response-also-requires-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: alsoRequiresSchema,
				State: &tfsdk.State{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: alsoRequiresSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: alsoRequiresSchema,
				},
			},
		},
		"response-resource-conflicts-with-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: conflictsWithSchema,
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceConflictsWithValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
					}),
					Schema: conflictsWithSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: conflictsWithSchema,
				},
			},
		},
		"response-resource-exactly-one-of-group-all-null": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: exactlyOneOfSchema,
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceExactlyOneOfValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: exactlyOneOfSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: exactlyOneOfSchema,
				},
			},
		},
		"response-resource-also-requires-group": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: alsoRequiresSchema,
				Resource: &testprovider.ResourceWithConfigValidators{
					Resource: &testprovider.Resource{},
					ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
						return []resource.ConfigValidator{&testResourceAlsoRequiresValidator{paths: path.Expressions{
							path.MatchRoot("alpha"),
							path.MatchRoot("beta"),
						}}}
					},
				},
				State: &tfsdk.State{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: alsoRequiresSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(validatorGroupType, map[string]tftypes.Value{
						"alpha": tftypes.NewValue(tftypes.String, nil),
						"beta":  tftypes.NewValue(tftypes.String, nil),
					}),
					Schema: alsoRequiresSchema,
				},
			},
		},
		"response-timeouts-and-optional-number": {
			server: &fwserver.Server{
				Provider: &testprovider.Provider{},
			},
			request: &fwserver.GenerateResourceConfigRequest{
				ResourceSchema: timeoutsAndNumberSchema,
				State: &tfsdk.State{
					Raw: tftypes.NewValue(timeoutsAndNumberType, map[string]tftypes.Value{
						"test_optional_number": tftypes.NewValue(tftypes.Number, big.NewFloat(0)),
						"test_nonzero_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(7)),
						"timeouts": tftypes.NewValue(timeoutsType, map[string]tftypes.Value{
							"create": tftypes.NewValue(tftypes.String, "30m"),
						}),
					}),
					Schema: timeoutsAndNumberSchema,
				},
			},
			expectedResponse: &fwserver.GenerateResourceConfigResponse{
				GeneratedConfig: &tfsdk.Config{
					Raw: tftypes.NewValue(timeoutsAndNumberType, map[string]tftypes.Value{
						"test_optional_number": tftypes.NewValue(tftypes.Number, nil),
						"test_nonzero_number":  tftypes.NewValue(tftypes.Number, big.NewFloat(7)),
						"timeouts":             tftypes.NewValue(timeoutsType, nil),
					}),
					Schema: timeoutsAndNumberSchema,
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.GenerateResourceConfigResponse{}
			testCase.server.GenerateResourceConfig(t.Context(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse, cmp.AllowUnexported(privatestate.ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

type testConflictsWithStringValidator struct {
	paths path.Expressions
}

func (v testConflictsWithStringValidator) Description(context.Context) string {
	return ""
}

func (v testConflictsWithStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testConflictsWithStringValidator) ValidateString(context.Context, schemavalidator.StringRequest, *schemavalidator.StringResponse) {
}

func (v testConflictsWithStringValidator) ConflictsWithPaths() path.Expressions {
	return v.paths
}

type testExactlyOneOfStringValidator struct {
	paths path.Expressions
}

func (v testExactlyOneOfStringValidator) Description(context.Context) string {
	return ""
}

func (v testExactlyOneOfStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testExactlyOneOfStringValidator) ValidateString(context.Context, schemavalidator.StringRequest, *schemavalidator.StringResponse) {
}

func (v testExactlyOneOfStringValidator) ExactlyOneOfPaths() path.Expressions {
	return v.paths
}

type testAlsoRequiresStringValidator struct {
	paths path.Expressions
}

func (v testAlsoRequiresStringValidator) Description(context.Context) string {
	return ""
}

func (v testAlsoRequiresStringValidator) MarkdownDescription(context.Context) string {
	return ""
}

func (v testAlsoRequiresStringValidator) ValidateString(context.Context, schemavalidator.StringRequest, *schemavalidator.StringResponse) {
}

func (v testAlsoRequiresStringValidator) AlsoRequiresPaths() path.Expressions {
	return v.paths
}

type testResourceConflictsWithValidator struct {
	testprovider.ResourceConfigValidator
	paths path.Expressions
}

func (v *testResourceConflictsWithValidator) ConflictsWithPaths() path.Expressions { return v.paths }

type testResourceExactlyOneOfValidator struct {
	testprovider.ResourceConfigValidator
	paths path.Expressions
}

func (v *testResourceExactlyOneOfValidator) ExactlyOneOfPaths() path.Expressions { return v.paths }

type testResourceAlsoRequiresValidator struct {
	testprovider.ResourceConfigValidator
	paths path.Expressions
}

func (v *testResourceAlsoRequiresValidator) AlsoRequiresPaths() path.Expressions { return v.paths }

var _ schemavalidator.String = testConflictsWithStringValidator{}
var _ schemavalidator.ConflictsWithValidator = testConflictsWithStringValidator{}
var _ schemavalidator.String = testExactlyOneOfStringValidator{}
var _ schemavalidator.ExactlyOneOfValidator = testExactlyOneOfStringValidator{}
var _ schemavalidator.String = testAlsoRequiresStringValidator{}
var _ schemavalidator.AlsoRequiresValidator = testAlsoRequiresStringValidator{}
var _ resource.ConfigValidator = &testResourceConflictsWithValidator{}
var _ resource.ConfigValidator = &testResourceExactlyOneOfValidator{}
var _ resource.ConfigValidator = &testResourceAlsoRequiresValidator{}
var _ schemavalidator.ConflictsWithValidator = &testResourceConflictsWithValidator{}
var _ schemavalidator.ExactlyOneOfValidator = &testResourceExactlyOneOfValidator{}
var _ schemavalidator.AlsoRequiresValidator = &testResourceAlsoRequiresValidator{}
