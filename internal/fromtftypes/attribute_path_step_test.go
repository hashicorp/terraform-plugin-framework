// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromtftypes_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfType        tftypes.AttributePathStep
		attrType      attr.Type
		expected      path.PathStep
		expectedError error
	}{
		"nil": {
			tfType:        nil,
			expected:      nil,
			expectedError: fmt.Errorf("unknown tftypes.AttributePathStep: <nil>"),
		},
		"PathStepAttributeName": {
			tfType:   tftypes.AttributeName("test"),
			expected: path.PathStepAttributeName("test"),
		},
		"PathStepElementKeyInt": {
			tfType:   tftypes.ElementKeyInt(1),
			expected: path.PathStepElementKeyInt(1),
		},
		"PathStepElementKeyString": {
			tfType:   tftypes.ElementKeyString("test"),
			expected: path.PathStepElementKeyString("test"),
		},
		"PathStepElementKeyValue": {
			tfType:   tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			attrType: types.StringType,
			expected: path.PathStepElementKeyValue{Value: types.StringValue("test")},
		},
		"PathStepElementKeyValue-error": {
			tfType:        tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
			attrType:      types.BoolType,
			expected:      nil,
			expectedError: fmt.Errorf("unable to create PathStepElementKeyValue from tftypes.Value: unable to convert tftypes.Value (tftypes.String<\"test\">) to attr.Value: can't unmarshal tftypes.String into *bool, expected boolean"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := fromtftypes.AttributePathStep(context.Background(), testCase.tfType, testCase.attrType)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, tfType: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
