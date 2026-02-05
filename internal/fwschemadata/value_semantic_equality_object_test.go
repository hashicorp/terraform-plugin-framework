// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueSemanticEqualityObject(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		// Type and AttributeTypes without semantic equality
		"ObjectValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.StringType,
					},
					map[string]attr.Value{
						"test_attr": types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.StringType,
					},
					map[string]attr.Value{
						"test_attr": types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.StringType,
					},
					map[string]attr.Value{
						"test_attr": types.StringValue("new"),
					},
				),
			},
		},
		// AttributeTypes with semantic equality
		"ObjectValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
			},
		},
		"ObjectValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: false,
						},
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
		},
		"ObjectValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					map[string]attr.Value{
						"test_attr": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		// Nested AttributeTypes with semantic equality
		"ObjectValue-ObjectValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
		},
		"ObjectValue-ObjectValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: false,
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
		},
		"ObjectValue-ObjectValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						),
					},
				),
				ProposedNewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						},
					},
					map[string]attr.Value{
						"test_attr": types.ObjectValueMust(
							map[string]attr.Type{
								"test_attr": testtypes.StringTypeWithSemanticEquals{
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
							map[string]attr.Value{
								"test_attr": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		// Type with semantic equality
		"ObjectValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"ObjectValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"ObjectValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ObjectValueWithSemanticEquals{
					ObjectValue: types.ObjectValueMust(
						map[string]attr.Type{
							"test_attr": types.StringType,
						},
						map[string]attr.Value{
							"test_attr": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEqualityObject(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
