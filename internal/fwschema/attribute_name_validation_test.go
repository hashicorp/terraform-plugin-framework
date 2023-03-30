package fwschema_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestIsReservedProviderAttributeName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		name          string
		attributePath path.Path
		expected      diag.Diagnostics
	}{
		"empty-path": {
			name:          "",
			attributePath: path.Empty(),
			expected:      nil,
		},
		"invalid-path": {
			name:          "",
			attributePath: path.Empty().AtListIndex(0),
			expected:      nil,
		},
		"non-root-attribute-name": {
			name:          "alias",
			attributePath: path.Root("test").AtName("alias"),
			expected:      nil,
		},
		"alias": {
			name:          "alias",
			attributePath: path.Root("alias"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the provider schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"alias\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"version": {
			name:          "version",
			attributePath: path.Root("version"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the provider schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"version\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"other": {
			name:          "other",
			attributePath: path.Root("other"),
			expected:      nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.IsReservedProviderAttributeName(testCase.name, testCase.attributePath)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestIsReservedResourceAttributeName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		name          string
		attributePath path.Path
		expected      diag.Diagnostics
	}{
		"empty-path": {
			name:          "",
			attributePath: path.Empty(),
			expected:      nil,
		},
		"invalid-path": {
			name:          "",
			attributePath: path.Empty().AtListIndex(0),
			expected:      nil,
		},
		"non-root-attribute-name": {
			name:          "count",
			attributePath: path.Root("test").AtName("count"),
			expected:      nil,
		},
		"connection": {
			name:          "connection",
			attributePath: path.Root("connection"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"connection\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"count": {
			name:          "count",
			attributePath: path.Root("count"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"count\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"depends_on": {
			name:          "depends_on",
			attributePath: path.Root("depends_on"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"depends_on\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		// TODO: Validate for_each
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/704
		// "for_each": {
		// 	name: "for_each",
		//	attributePath: path.Root("for_each"),
		// 	expected: diag.Diagnostics{
		// 		diag.NewErrorDiagnostic(
		// 			"Reserved Root Attribute/Block Name",
		// 			"When validating the resource or data source schema, an implementation issue was found. "+
		// 				"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
		// 				"\"for_each\" is a reserved root attribute/block name. "+
		// 				"This is to prevent practitioners from needing special Terraform configuration syntax.",
		// 		),
		// 	},
		// },
		"lifecycle": {
			name:          "lifecycle",
			attributePath: path.Root("lifecycle"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"lifecycle\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"provider": {
			name:          "provider",
			attributePath: path.Root("provider"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"provider\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"provisioner": {
			name:          "provisioner",
			attributePath: path.Root("provisioner"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"provisioner\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
		"other": {
			name:          "other",
			attributePath: path.Root("other"),
			expected:      nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.IsReservedResourceAttributeName(testCase.name, testCase.attributePath)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
