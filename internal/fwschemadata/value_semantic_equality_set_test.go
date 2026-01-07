// Copyright IBM Corp. 2021, 2025
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

func TestValueSemanticEqualitySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		// Type and ElementType without semantic equality
		"SetValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
		},
		"SetValue-diff-order": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("prior"),
						types.StringValue("value"),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("value"),
						types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("value"),
						types.StringValue("new"),
					},
				),
			},
		},
		// ElementType with semantic equality
		"SetValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
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
				NewValue: types.SetValueMust(
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
		"SetValue-StringValuableWithSemanticEquals-true-diff-order": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue: types.StringValue("keep-lowercase-123"),
							SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
								StringValue: types.StringValue("KEEP-LOWERCASE-123"),
							},
						},
						testtypes.StringValueWithSemanticEquals{
							StringValue: types.StringValue("keep-lowercase-456"),
							SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
								StringValue: types.StringValue("KEEP-LOWERCASE-456"),
							},
						},
					},
				),
				ProposedNewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue: types.StringValue("KEEP-LOWERCASE-456"),
							SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
								StringValue: types.StringValue("keep-lowercase-456"),
							},
						},
						testtypes.StringValueWithSemanticEquals{
							StringValue: types.StringValue("KEEP-LOWERCASE-123"),
							SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
								StringValue: types.StringValue("keep-lowercase-123"),
							},
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue: types.StringValue("keep-lowercase-123"),
							SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
								StringValue: types.StringValue("KEEP-LOWERCASE-123"),
							},
						},
						testtypes.StringValueWithSemanticEquals{
							StringValue: types.StringValue("keep-lowercase-456"),
							SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
								StringValue: types.StringValue("KEEP-LOWERCASE-456"),
							},
						},
					},
				),
			},
		},
		"SetValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
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
				NewValue: types.SetValueMust(
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
		"SetValue-StringValuableWithSemanticEquals-false-diff-order": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: false,
						},
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("value"),
							SemanticEquals: false,
						},
					},
				),
				ProposedNewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("value"),
							SemanticEquals: false,
						},
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("value"),
							SemanticEquals: false,
						},
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
		},
		"SetValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
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
				NewValue: types.SetValueMust(
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
		"SetValue-SetValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
		"SetValue-SetValue-StringValuableWithSemanticEquals-true-diff-order": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-123"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-123"),
									},
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-456"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-456"),
									},
								},
							},
						),
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-789"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-789"),
									},
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-012"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-012"),
									},
								},
							},
						),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("KEEP-LOWERCASE-012"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("keep-lowercase-012"),
									},
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("KEEP-LOWERCASE-789"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("keep-lowercase-789"),
									},
								},
							},
						),
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("KEEP-LOWERCASE-456"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("keep-lowercase-456"),
									},
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("KEEP-LOWERCASE-123"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("keep-lowercase-123"),
									},
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-123"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-123"),
									},
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-456"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-456"),
									},
								},
							},
						),
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-789"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-789"),
									},
								},
								testtypes.StringValueWithSemanticEquals{
									StringValue: types.StringValue("keep-lowercase-012"),
									SemanticallyEqualTo: testtypes.StringValueWithSemanticEquals{
										StringValue: types.StringValue("KEEP-LOWERCASE-012"),
									},
								},
							},
						),
					},
				),
			},
		},
		"SetValue-SetValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
		"SetValue-SetValue-StringValuableWithSemanticEquals-false-diff-order": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("value"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("value"),
									SemanticEquals: false,
								},
							},
						),
						types.SetValueMust(
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
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("value"),
									SemanticEquals: false,
								},
							},
						),
						types.SetValueMust(
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
		"SetValue-SetValue-StringValuableWithSemanticEquals-NewValueElementsGreaterThanPriorValueElements": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
		"SetValue-SetValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.SetValueMust(
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
		"SetValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"SetValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"SetValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
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
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
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
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEqualitySet(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
