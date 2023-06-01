// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
		"for_each": {
			name:          "for_each",
			attributePath: path.Root("for_each"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Reserved Root Attribute/Block Name",
					"When validating the resource or data source schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"for_each\" is a reserved root attribute/block name. "+
						"This is to prevent practitioners from needing special Terraform configuration syntax.",
				),
			},
		},
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

func TestIsValidAttributeName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		name     string
		expected diag.Diagnostics
	}{
		"empty": {
			name: "",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"\" at schema path \"\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"ascii-lowercase-alphabet": {
			name:     "test",
			expected: nil,
		},
		"ascii-lowercase-alphabet-leading-underscore": {
			name:     "_test",
			expected: nil,
		},
		"ascii-lowercase-alphabet-middle-hyphens": {
			name: "test-me",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"test-me\" at schema path \"test-me\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"ascii-lowercase-alphabet-middle-underscore": {
			name:     "test_me",
			expected: nil,
		},
		"ascii-lowercase-alphanumeric": {
			name:     "test123",
			expected: nil,
		},
		"ascii-lowercase-alphanumeric-leading-numeric": {
			name: "123test",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"123test\" at schema path \"123test\" is an invalid attribute/block name. "+
						"Names must begin with a lowercase alphabet character (a-z) or underscore (_) and "+
						"must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"ascii-uppercase-alphabet": {
			name: "TEST",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"TEST\" at schema path \"TEST\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"ascii-uppercase-alphanumeric": {
			name: "TEST123",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"TEST123\" at schema path \"TEST123\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"invalid-bytes": {
			name: "\xff\xff",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"\\xff\\xff\" at schema path \"\\xff\\xff\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
		"unicode": {
			name: `tést`, // t, latin small letter e with acute (00e9), s, t
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Attribute/Block Name",
					"When validating the schema, an implementation issue was found. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"\"tést\" at schema path \"tést\" is an invalid attribute/block name. "+
						"Names must only contain lowercase alphanumeric characters (a-z, 0-9) and underscores (_).",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fwschema.IsValidAttributeName(testCase.name, path.Root(testCase.name))

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
