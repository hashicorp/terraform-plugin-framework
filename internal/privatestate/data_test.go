package privatestate_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
)

func TestData_Bytes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		data          privatestate.Data
		expected      []byte
		expectedDiags diag.Diagnostics
	}{
		"empty": {
			data:          privatestate.Data{},
			expected:      []byte(`{}`),
			expectedDiags: diag.Diagnostics{},
		},
		"framework-data": {
			data: privatestate.Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte("framework value one"),
					".frameworkKeyTwo": []byte("framework value two"),
				},
			},
			expected:      []byte(`{".frameworkKeyOne":"ZnJhbWV3b3JrIHZhbHVlIG9uZQ==",".frameworkKeyTwo":"ZnJhbWV3b3JrIHZhbHVlIHR3bw=="}`),
			expectedDiags: diag.Diagnostics{},
		},
		"provider-data": {
			data: privatestate.Data{
				Provider: map[string][]byte{
					"providerKeyOne": []byte("provider value one"),
					"providerKeyTwo": []byte("provider value two")},
			},
			expected:      []byte(`{"providerKeyOne":"cHJvdmlkZXIgdmFsdWUgb25l","providerKeyTwo":"cHJvdmlkZXIgdmFsdWUgdHdv"}`),
			expectedDiags: diag.Diagnostics{},
		},
		"framework-provider-data": {
			data: privatestate.Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte("framework value one"),
					".frameworkKeyTwo": []byte("framework value two"),
				},
				Provider: map[string][]byte{
					"providerKeyOne": []byte("provider value one"),
					"providerKeyTwo": []byte("provider value two")},
			},
			expected:      []byte(`{".frameworkKeyOne":"ZnJhbWV3b3JrIHZhbHVlIG9uZQ==",".frameworkKeyTwo":"ZnJhbWV3b3JrIHZhbHVlIHR3bw==","providerKeyOne":"cHJvdmlkZXIgdmFsdWUgb25l","providerKeyTwo":"cHJvdmlkZXIgdmFsdWUgdHdv"}`),
			expectedDiags: diag.Diagnostics{},
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
	frameworkProviderData, _ := json.Marshal(map[string][]byte{
		".frameworkKeyOne": []byte("framework value one"),
		".frameworkKeyTwo": []byte("framework value two"),
		"providerKeyOne":   []byte("provider value one"),
		"providerKeyTwo":   []byte("provider value two"),
	})

	testCases := map[string]struct {
		data          []byte
		expected      privatestate.Data
		expectedDiags diag.Diagnostics
	}{
		"invalid-json": {
			data:     []byte(`{`),
			expected: privatestate.Data{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("Error Decoding Private State",
					"An error was encountered when decoding private state: unexpected end of JSON input.\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"framework-provider-data": {
			data: frameworkProviderData,
			expected: privatestate.Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte("framework value one"),
					".frameworkKeyTwo": []byte("framework value two"),
				},
				Provider: map[string][]byte{
					"providerKeyOne": []byte("provider value one"),
					"providerKeyTwo": []byte("provider value two"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, actualDiags := privatestate.NewData(context.Background(), testCase.data)

			if diff := cmp.Diff(actual, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(actualDiags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestProviderData_GetKey(t *testing.T) {
	testCases := map[string]struct {
		providerData  privatestate.ProviderData
		key           string
		expected      []byte
		expectedDiags diag.Diagnostics
	}{
		"key-invalid": {
			providerData: map[string][]byte{
				"key": []byte("value"),
			},
			key: ".key",
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Restricted Resource Private State Namespace",
					"Using a period ('.') as a prefix for a key used in private state is not allowed\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"key-not-found": {
			providerData: map[string][]byte{
				"key": []byte("value"),
			},
			key: "key-not-found",
		},
		"key-found": {
			providerData: map[string][]byte{
				"key": []byte("value"),
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
	// 1 x 1 transparent gif pixel.
	const transPixel = "\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00\x21\xF9\x04\x01\x00\x00\x00\x00\x2C\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3B"

	testCases := map[string]struct {
		key      string
		value    []byte
		expected diag.Diagnostics
	}{
		"key-invalid": {
			key: ".key",
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Restricted Resource Private State Namespace",
					"Using a period ('.') as a prefix for a key used in private state is not allowed\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"utf8-invalid": {
			key:   "key",
			value: []byte(fmt.Sprintf(`{"key": "%s"}`, transPixel)),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"UTF-8 Invalid",
					"Values stored in private state must be valid UTF-8\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"value-json-invalid": {
			key:   "key",
			value: []byte("{"),
			expected: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"JSON Invalid",
					"Values stored in private state must be valid JSON\n\n"+
						"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
				),
			},
		},
		"key-value-ok": {
			key:   "key",
			value: []byte(`{"key": "value"}`),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := privatestate.ProviderData{}.SetKey(context.Background(), testCase.key, testCase.value)

			if diff := cmp.Diff(actual, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestValidateProviderDataKey(t *testing.T) {
	testCases := map[string]struct {
		key      string
		expected diag.Diagnostics
	}{
		"namespace-restricted": {
			key: ".restricted",
			expected: diag.Diagnostics{diag.NewErrorDiagnostic(
				"Restricted Resource Private State Namespace",
				"Using a period ('.') as a prefix for a key used in private state is not allowed\n\n"+
					"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
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

			actual := privatestate.ValidateProviderDataKey(context.Background(), testCase.key)

			if diff := cmp.Diff(actual, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
