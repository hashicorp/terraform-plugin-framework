// Copyright (c) HashiCorp, Inc.
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

func TestValueSemanticEqualityMap(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		// Type and ElementType without semantic equality
		"MapValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"testkey": types.StringValue("new"),
					},
				),
			},
		},
		// ElementType with semantic equality
		"MapValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
				ProposedNewValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
			},
		},
		"MapValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: false,
						},
					},
				),
				ProposedNewValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
		},
		"MapValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
				ProposedNewValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
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
				NewValue: types.MapValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					map[string]attr.Value{
						"testkey": testtypes.StringValueWithSemanticEquals{
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
		// Nested ElementType with semantic equality
		"MapValue-MapValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
				ProposedNewValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
		},
		"MapValue-MapValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
				ProposedNewValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
		},
		"MapValue-MapValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
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
				ProposedNewValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
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
				NewValue: types.MapValueMust(
					types.MapType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					map[string]attr.Value{
						"testkey": types.MapValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							map[string]attr.Value{
								"testkey": testtypes.StringValueWithSemanticEquals{
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
		"MapValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"MapValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"MapValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
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
				NewValue: testtypes.MapValueWithSemanticEquals{
					MapValue: types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"testkey": types.StringValue("new"),
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

			fwschemadata.ValueSemanticEqualityMap(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
