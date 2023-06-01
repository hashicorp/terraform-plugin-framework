// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package privatestate

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestData_Bytes(t *testing.T) {
	// 1 x 1 transparent gif pixel.
	const transPixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"

	t.Parallel()

	testCases := map[string]struct {
		data          *Data
		expected      []byte
		expectedDiags diag.Diagnostics
	}{
		"nil": {
			data: nil,
		},
		"empty": {
			data: &Data{},
		},
		"uninitialized-provider-data-data": {
			data: &Data{
				Provider: &ProviderData{},
			},
		},
		"empty-initialized-provider-data-data": {
			data: &Data{
				Provider: &ProviderData{
					data: nil,
				},
			},
		},
		"framework-data-value-invalid-utf-8": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(fmt.Sprintf(`{"fwKeyOne": "%s"}`, transPixel)),
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Encoding Private State",
					"An error was encountered when validating private state value."+
						"The value associated with key \".frameworkKeyOne\" is is not valid UTF-8.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-data-value-invalid-json": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`}`),
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Encoding Private State",
					"An error was encountered when validating private state value."+
						"The value associated with key \".frameworkKeyOne\" is is not valid JSON.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-data": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
					".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
				},
			},
			expected: []byte(`{` +
				`".frameworkKeyOne":"eyJmd0tleU9uZSI6IHsiazAiOiAiemVybyIsICJrMSI6IDF9fQ==",` +
				`".frameworkKeyTwo":"eyJmd0tleVR3byI6IHsiazIiOiAidHdvIiwgImszIjogM319"` +
				`}`),
		},
		"framework-data-value-nil": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
					".frameworkKeyTwo": nil,
				},
			},
			expected: []byte(`{` +
				`".frameworkKeyOne":"eyJmd0tleU9uZSI6IHsiazAiOiAiemVybyIsICJrMSI6IDF9fQ=="` +
				`}`),
		},
		"framework-data-value-zero-len": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
					".frameworkKeyTwo": {},
				},
			},
			expected: []byte(`{` +
				`".frameworkKeyOne":"eyJmd0tleU9uZSI6IHsiazAiOiAiemVybyIsICJrMSI6IDF9fQ=="` +
				`}`),
		},
		"provider-data-data-value-invalid-utf-8": {
			data: &Data{
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(fmt.Sprintf(`{"key": "%s"}`, transPixel)),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Encoding Private State",
					"An error was encountered when validating private state value."+
						"The value associated with key \"providerKeyOne\" is is not valid UTF-8.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"provider-data-data-value-invalid-json": {
			data: &Data{
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`}`),
					},
				},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Encoding Private State",
					"An error was encountered when validating private state value."+
						"The value associated with key \"providerKeyOne\" is is not valid JSON.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"provider-data": {
			data: &Data{
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
						"providerKeyTwo": []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
					},
				},
			},
			expected: []byte(`{` +
				`"providerKeyOne":"eyJwS2V5T25lIjogeyJrMCI6ICJ6ZXJvIiwgImsxIjogMX19",` +
				`"providerKeyTwo":"eyJwS2V5VHdvIjogeyJrMiI6ICJ0d28iLCAiazMiOiAzfX0="` +
				`}`),
		},
		"provider-data-data-value-nil": {
			data: &Data{
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
						"providerKeyTwo": nil,
					},
				},
			},
			expected: []byte(`{"providerKeyOne":"eyJwS2V5T25lIjogeyJrMCI6ICJ6ZXJvIiwgImsxIjogMX19"}`),
		},
		"provider-data-data-value-zero-len": {
			data: &Data{
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
						"providerKeyTwo": {},
					},
				},
			},
			expected: []byte(`{"providerKeyOne":"eyJwS2V5T25lIjogeyJrMCI6ICJ6ZXJvIiwgImsxIjogMX19"}`),
		},
		"framework-provider-data": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
					".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
				},
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
						"providerKeyTwo": []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
					},
				},
			},
			expected: []byte(`{` +
				`".frameworkKeyOne":"eyJmd0tleU9uZSI6IHsiazAiOiAiemVybyIsICJrMSI6IDF9fQ==",` +
				`".frameworkKeyTwo":"eyJmd0tleVR3byI6IHsiazIiOiAidHdvIiwgImszIjogM319",` +
				`"providerKeyOne":"eyJwS2V5T25lIjogeyJrMCI6ICJ6ZXJvIiwgImsxIjogMX19",` +
				`"providerKeyTwo":"eyJwS2V5VHdvIjogeyJrMiI6ICJ0d28iLCAiazMiOiAzfX0="` +
				`}`),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, actualDiags := testCase.data.Bytes(context.Background())

			if diff := cmp.Diff(actual, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(actualDiags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewData(t *testing.T) {
	t.Parallel()

	// 1 x 1 transparent gif pixel.
	const transPixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"

	frameworkInvalidUTF8 := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(transPixel),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	frameworkValueInvalidUTF8 := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(fmt.Sprintf(`{"fwKeyOne": "%s"}`, transPixel)),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	frameworkInvalidJSON := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	frameworkValueInvalidJSON := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{"fwKeyOne": { }`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	providerInvalidUTF8 := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(transPixel),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	providerValueInvalidUTF8 := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(fmt.Sprintf(`{"fwKeyOne": "%s"}`, transPixel)),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	providerInvalidJSON := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	providerValueInvalidJSON := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{"pKeyOne": { }`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	frameworkProviderData := MustMarshalToJson(map[string][]byte{
		".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
		".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
		"providerKeyOne":   []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
		"providerKeyTwo":   []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
	})

	sdkJSON, err := json.Marshal(map[string]any{
		"schema_version": "2",
	})

	if err != nil {
		t.Fatalf("unexpected error marshaling SDK JSON: %s", err)
	}

	testCases := map[string]struct {
		data          []byte
		expected      *Data
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			data: []byte{},
		},
		"invalid-json": {
			data: []byte(`{`),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Decoding Private State",
					"An error was encountered when decoding private state: unexpected end of JSON input.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"empty-json": {
			data: []byte(`{}`),
			expected: &Data{
				Framework: map[string][]byte{},
				Provider: &ProviderData{
					data: map[string][]byte{},
				},
			},
		},
		"framework-invalid-utf-8": {
			data: frameworkInvalidUTF8,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \".frameworkKeyOne\" is is not valid UTF-8.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-value-invalid-utf-8": {
			data: frameworkValueInvalidUTF8,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \".frameworkKeyOne\" is is not valid UTF-8.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-invalid-json": {
			data: frameworkInvalidJSON,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \".frameworkKeyOne\" is is not valid JSON.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-value-invalid-json": {
			data: frameworkValueInvalidJSON,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \".frameworkKeyOne\" is is not valid JSON.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"provider-invalid-utf-8": {
			data: providerInvalidUTF8,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \"providerKeyOne\" is is not valid UTF-8.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"provider-value-invalid-utf-8": {
			data: providerValueInvalidUTF8,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \"providerKeyOne\" is is not valid UTF-8.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"provider-invalid-json": {
			data: providerInvalidJSON,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \"providerKeyOne\" is is not valid JSON.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"provider-value-invalid-json": {
			data: providerValueInvalidJSON,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Error Decoding Private State",
					"An error was encountered when validating private state value.\n"+
						"The value being supplied for key \"providerKeyOne\" is is not valid JSON.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-provider-data": {
			data: frameworkProviderData,
			expected: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"fwKeyOne": {"k0": "zero", "k1": 1}}`),
					".frameworkKeyTwo": []byte(`{"fwKeyTwo": {"k2": "two", "k3": 3}}`),
				},
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
						"providerKeyTwo": []byte(`{"pKeyTwo": {"k2": "two", "k3": 3}}`),
					},
				},
			},
		},
		"sdk-ignore": {
			data:     sdkJSON,
			expected: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, actualDiags := NewData(context.Background(), testCase.data)

			if diff := cmp.Diff(actual, testCase.expected, cmp.AllowUnexported(ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(actualDiags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestNewProviderData(t *testing.T) {
	t.Parallel()

	// 1 x 1 transparent gif pixel.
	const transPixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"

	invalidKey := MustMarshalToJson(map[string][]byte{
		".providerKeyOne": {},
	})

	invalidUTF8Value := MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(transPixel),
	})

	invalidJSONValue := MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{`),
	})

	validKeyValue := MustMarshalToJson(map[string][]byte{
		"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
	})

	testCases := map[string]struct {
		data          []byte
		expected      *ProviderData
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			data: []byte{},
			expected: &ProviderData{
				data: map[string][]byte{},
			},
		},
		"invalid-json": {
			data: []byte(`{`),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Decoding Provider Data",
					"An error was encountered when decoding provider data: unexpected end of JSON input.\n\n"+
						"Please check that the data you are supplying is a byte representation of valid JSON.",
				),
			},
		},
		"empty-json": {
			data: []byte(`{}`),
			expected: &ProviderData{
				data: map[string][]byte{},
			},
		},
		"invalid-key": {
			data: invalidKey,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Restricted Resource Private State Namespace",
					"Using a period ('.') as a prefix for a key used in private state is not allowed.\n\n"+
						"The key \".providerKeyOne\" is invalid. Please check the key you are supplying does not use a a period ('.') as a prefix.",
				),
			},
		},
		"invalid-utf-8-value": {
			data: invalidUTF8Value,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"UTF-8 Invalid",
					"Values stored in private state must be valid UTF-8.\n\n"+
						"The value being supplied for key \"providerKeyOne\" is invalid. Please verify that the value is valid UTF-8.",
				),
			},
		},
		"invalid-json-value": {
			data: invalidJSONValue,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"JSON Invalid",
					"Values stored in private state must be valid JSON.\n\n"+
						"The value being supplied for key \"providerKeyOne\" is invalid. Please verify that the value is valid JSON.",
				),
			},
		},
		"valid-key-value": {
			data: validKeyValue,
			expected: &ProviderData{
				data: map[string][]byte{
					"providerKeyOne": []byte(`{"pKeyOne": {"k0": "zero", "k1": 1}}`),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, actualDiags := NewProviderData(context.Background(), testCase.data)

			if diff := cmp.Diff(actual, testCase.expected, cmp.AllowUnexported(ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(actualDiags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestProviderDataEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		providerData *ProviderData
		other        *ProviderData
		expected     bool
	}{
		"nil-nil": {
			providerData: nil,
			other:        nil,
			expected:     true,
		},
		"nil-empty": {
			providerData: nil,
			other:        EmptyProviderData(context.Background()),
			expected:     false,
		},
		"empty-nil": {
			providerData: EmptyProviderData(context.Background()),
			other:        nil,
			expected:     false,
		},
		"empty-data": {
			providerData: EmptyProviderData(context.Background()),
			other: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test": []byte(`{}`)}),
			),
			expected: false,
		},
		"data-empty": {
			providerData: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test": []byte(`{}`)}),
			),
			other:    EmptyProviderData(context.Background()),
			expected: false,
		},
		"data-data-different-keys": {
			providerData: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test1": []byte(`{}`)}),
			),
			other: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test2": []byte(`{}`)}),
			),
			expected: false,
		},
		"data-data-different-values": {
			providerData: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test": []byte(`{"subtest":true}`)}),
			),
			other: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test": []byte(`{"subtest":false}`)}),
			),
			expected: false,
		},
		"data-data-equal": {
			providerData: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test": []byte(`{}`)}),
			),
			other: MustProviderData(
				context.Background(),
				MustMarshalToJson(map[string][]byte{"test": []byte(`{}`)}),
			),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.providerData.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestProviderData_GetKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		providerData  *ProviderData
		key           string
		expected      []byte
		expectedDiags diag.Diagnostics
	}{
		"nil": {
			providerData: &ProviderData{},
			key:          "key",
		},
		"key-invalid": {
			providerData: &ProviderData{
				data: map[string][]byte{
					"providerKeyOne": []byte(`{"pKeyOne": "provider value one"}`),
				},
			},
			key: ".key",
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Restricted Resource Private State Namespace",
					"Using a period ('.') as a prefix for a key used in private state is not allowed.\n\n"+
						`The key ".key" is invalid. Please check the key you are supplying does not use a a period ('.') as a prefix.`,
				),
			},
		},
		"key-not-found": {
			providerData: &ProviderData{
				data: map[string][]byte{
					"providerKeyOne": []byte(`{"pKeyOne": "provider value one"}`),
				},
			},
			key: "key-not-found",
		},
		"key-found": {
			providerData: &ProviderData{
				data: map[string][]byte{
					"key": []byte("value"),
				},
			},
			key:      "key",
			expected: []byte("value"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, actualDiags := testCase.providerData.GetKey(context.Background(), testCase.key)

			if diff := cmp.Diff(actual, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(actualDiags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestProviderData_SetKey(t *testing.T) {
	t.Parallel()

	// 1 x 1 transparent gif pixel.
	const transPixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"

	testCases := map[string]struct {
		providerData  *ProviderData
		key           string
		value         []byte
		expected      *ProviderData
		expectedDiags diag.Diagnostics
	}{
		"nil": {
			providerData: nil,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Uninitialized ProviderData",
					"ProviderData must be initialized before it is used.\n\n"+
						"Call privatestate.NewProviderData to obtain an initialized instance of ProviderData."),
			},
		},
		"key-invalid-data-uninitialized": {
			providerData: &ProviderData{},
			key:          ".key",
			expected: &ProviderData{
				data: map[string][]byte{},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Restricted Resource Private State Namespace",
					"Using a period ('.') as a prefix for a key used in private state is not allowed.\n\n"+
						`The key ".key" is invalid. Please check the key you are supplying does not use a a period ('.') as a prefix.`,
				),
			},
		},
		"key-invalid-data-initialized": {
			providerData: &ProviderData{
				data: map[string][]byte{},
			},
			key: ".key",
			expected: &ProviderData{
				data: map[string][]byte{},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Restricted Resource Private State Namespace",
					"Using a period ('.') as a prefix for a key used in private state is not allowed.\n\n"+
						`The key ".key" is invalid. Please check the key you are supplying does not use a a period ('.') as a prefix.`,
				),
			},
		},
		"value-utf8-invalid-data-uninitialized": {
			providerData: &ProviderData{},
			key:          "key",
			value:        []byte(fmt.Sprintf(`{"key": "%s"}`, transPixel)),
			expected: &ProviderData{
				data: map[string][]byte{},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"UTF-8 Invalid",
					"Values stored in private state must be valid UTF-8.\n\n"+
						`The value being supplied for key "key" is invalid. Please verify that the value is valid UTF-8.`,
				),
			},
		},
		"value-utf8-invalid-data-initialized": {
			providerData: &ProviderData{
				data: map[string][]byte{},
			},
			key:   "key",
			value: []byte(fmt.Sprintf(`{"key": "%s"}`, transPixel)),
			expected: &ProviderData{
				data: map[string][]byte{},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"UTF-8 Invalid",
					"Values stored in private state must be valid UTF-8.\n\n"+
						`The value being supplied for key "key" is invalid. Please verify that the value is valid UTF-8.`,
				),
			},
		},
		"value-json-invalid-data-uninitialized": {
			providerData: &ProviderData{},
			key:          "key",
			value:        []byte("{"),
			expected: &ProviderData{
				data: map[string][]byte{},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"JSON Invalid",
					"Values stored in private state must be valid JSON.\n\n"+
						`The value being supplied for key "key" is invalid. Please verify that the value is valid JSON.`,
				),
			},
		},
		"value-json-invalid-data-initialized": {
			providerData: &ProviderData{
				data: map[string][]byte{},
			},
			key:   "key",
			value: []byte("{"),
			expected: &ProviderData{
				data: map[string][]byte{},
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"JSON Invalid",
					"Values stored in private state must be valid JSON.\n\n"+
						`The value being supplied for key "key" is invalid. Please verify that the value is valid JSON.`,
				),
			},
		},
		"key-value-ok-data-uninitialized": {
			providerData: &ProviderData{},
			key:          "key",
			value:        []byte(`{"key": {"k0": "zero", "k1": 1}}`),
			expected: &ProviderData{
				data: map[string][]byte{
					"key": []byte(`{"key": {"k0": "zero", "k1": 1}}`),
				},
			},
		},
		"key-value-ok-data-initialized": {
			providerData: &ProviderData{
				data: map[string][]byte{},
			},
			key:   "key",
			value: []byte(`{"key": {"k0": "zero", "k1": 1}}`),
			expected: &ProviderData{
				data: map[string][]byte{
					"key": []byte(`{"key": {"k0": "zero", "k1": 1}}`),
				},
			},
		},
		"key-value-added": {
			providerData: &ProviderData{
				data: map[string][]byte{
					"keyOne": []byte(`{"foo": "bar"}`),
				},
			},
			key:   "keyTwo",
			value: []byte(`{"buzz": "bazz"}`),
			expected: &ProviderData{
				data: map[string][]byte{
					"keyOne": []byte(`{"foo": "bar"}`),
					"keyTwo": []byte(`{"buzz": "bazz"}`),
				},
			},
		},
		"key-value-updated": {
			providerData: &ProviderData{
				data: map[string][]byte{
					"keyOne": []byte(`{"foo": "bar"}`),
				},
			},
			key:   "keyOne",
			value: []byte(`{"buzz": "bazz"}`),
			expected: &ProviderData{
				data: map[string][]byte{
					"keyOne": []byte(`{"buzz": "bazz"}`),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := testCase.providerData.SetKey(context.Background(), testCase.key, testCase.value)

			if diff := cmp.Diff(testCase.expected, testCase.providerData, cmp.AllowUnexported(ProviderData{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(actual, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestValidateProviderDataKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		key           string
		expectedDiags diag.Diagnostics
	}{
		"namespace-restricted": {
			key: ".restricted",
			expectedDiags: diag.Diagnostics{diag.NewErrorDiagnostic(
				"Restricted Resource Private State Namespace",
				"Using a period ('.') as a prefix for a key used in private state is not allowed.\n\n"+
					`The key ".restricted" is invalid. Please check the key you are supplying does not use a a period ('.') as a prefix.`,
			)},
		},
		"namespace-ok": {
			key: "unrestricted",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := ValidateProviderDataKey(context.Background(), testCase.key)

			if diff := cmp.Diff(actual, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
