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

func TestValueSemanticEqualityList(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		// Type and ElementType without semantic equality
		"ListValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
		},
		// ElementType with semantic equality
		"ListValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
				ProposedNewValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
			},
		},
		"ListValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: false,
						},
					},
				),
				ProposedNewValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
		},
		"ListValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
				ProposedNewValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
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
				NewValue: types.ListValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
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
		"ListValue-ListValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
				ProposedNewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
		},
		"ListValue-ListValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
				ProposedNewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
		},
		"ListValue-ListValue-StringValuableWithSemanticEquals-NewValueElementsGreaterThanPriorValueElements": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
				ProposedNewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new1"),
									SemanticEquals: true,
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new2"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new2"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
		},
		"ListValue-ListValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
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
				ProposedNewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
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
				NewValue: types.ListValueMust(
					types.ListType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.ListValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
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
		"ListValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"ListValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"ListValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
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
				NewValue: testtypes.ListValueWithSemanticEquals{
					ListValue: types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
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
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEqualityList(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
