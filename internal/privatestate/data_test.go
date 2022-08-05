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
		"framework-data": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"frameworkKeyOne": "framework value one"}`),
					".frameworkKeyTwo": []byte(`{"frameworkKeyTwo": "framework value two"}`),
				},
			},
			expected: []byte(`{` +
				`".frameworkKeyOne":"eyJmcmFtZXdvcmtLZXlPbmUiOiAiZnJhbWV3b3JrIHZhbHVlIG9uZSJ9",` +
				`".frameworkKeyTwo":"eyJmcmFtZXdvcmtLZXlUd28iOiAiZnJhbWV3b3JrIHZhbHVlIHR3byJ9"` +
				`}`),
		},
		"provider-data": {
			data: &Data{
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"providerKeyOne": "provider value one"}`),
						"providerKeyTwo": []byte(`{"providerKeyTwo": "provider value two"}`),
					},
				},
			},
			expected: []byte(`{` +
				`"providerKeyOne":"eyJwcm92aWRlcktleU9uZSI6ICJwcm92aWRlciB2YWx1ZSBvbmUifQ==",` +
				`"providerKeyTwo":"eyJwcm92aWRlcktleVR3byI6ICJwcm92aWRlciB2YWx1ZSB0d28ifQ=="` +
				`}`),
		},
		"framework-provider-data": {
			data: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"frameworkKeyOne": "framework value one"}`),
					".frameworkKeyTwo": []byte(`{"frameworkKeyTwo": "framework value two"}`),
				},
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"providerKeyOne": "provider value one"}`),
						"providerKeyTwo": []byte(`{"providerKeyTwo": "provider value two"}`),
					},
				},
			},
			expected: []byte(`{` +
				`".frameworkKeyOne":"eyJmcmFtZXdvcmtLZXlPbmUiOiAiZnJhbWV3b3JrIHZhbHVlIG9uZSJ9",` +
				`".frameworkKeyTwo":"eyJmcmFtZXdvcmtLZXlUd28iOiAiZnJhbWV3b3JrIHZhbHVlIHR3byJ9",` +
				`"providerKeyOne":"eyJwcm92aWRlcktleU9uZSI6ICJwcm92aWRlciB2YWx1ZSBvbmUifQ==",` +
				`"providerKeyTwo":"eyJwcm92aWRlcktleVR3byI6ICJwcm92aWRlciB2YWx1ZSB0d28ifQ=="` +
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
	frameworkProviderData, err := json.Marshal(map[string][]byte{
		".frameworkKeyOne": []byte(`{"frameworkKeyOne": "framework value one"}`),
		".frameworkKeyTwo": []byte(`{"frameworkKeyTwo": "framework value two"}`),
		"providerKeyOne":   []byte(`{"providerKeyOne": "provider value one"}`),
		"providerKeyTwo":   []byte(`{"providerKeyTwo": "provider value two"}`),
	})
	if err != nil {
		t.Errorf("could not marshal JSON: %s", err)
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
			data:     []byte(`{`),
			expected: nil,
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
		"framework-provider-data": {
			data: frameworkProviderData,
			expected: &Data{
				Framework: map[string][]byte{
					".frameworkKeyOne": []byte(`{"frameworkKeyOne": "framework value one"}`),
					".frameworkKeyTwo": []byte(`{"frameworkKeyTwo": "framework value two"}`),
				},
				Provider: &ProviderData{
					data: map[string][]byte{
						"providerKeyOne": []byte(`{"providerKeyOne": "provider value one"}`),
						"providerKeyTwo": []byte(`{"providerKeyTwo": "provider value two"}`),
					},
				},
			},
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

func TestProviderData_GetKey(t *testing.T) {
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
					"providerKeyOne": []byte(`{"providerKeyOne": "provider value one"}`),
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
					"providerKeyOne": []byte(`{"providerKeyOne": "provider value one"}`),
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
		"utf8-invalid-data-uninitialized": {
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
						`The value being supplied for key "key" is invalid. Please check the value you are supplying is valid UTF-8.`,
				),
			},
		},
		"utf8-invalid-data-initialized": {
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
						`The value being supplied for key "key" is invalid. Please check the value you are supplying is valid UTF-8.`,
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
						`The value being supplied for key "key" is invalid. Please check the value you are supplying is valid JSON.`,
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
						`The value being supplied for key "key" is invalid. Please check the value you are supplying is valid JSON.`,
				),
			},
		},
		"key-value-ok-data-uninitialized": {
			providerData: &ProviderData{},
			key:          "key",
			value:        []byte(`{"key": "value"}`),
			expected: &ProviderData{
				data: map[string][]byte{
					"key": []byte(`{"key": "value"}`),
				},
			},
		},
		"key-value-ok-data-initialized": {
			providerData: &ProviderData{
				data: map[string][]byte{},
			},
			key:   "key",
			value: []byte(`{"key": "value"}`),
			expected: &ProviderData{
				data: map[string][]byte{
					"key": []byte(`{"key": "value"}`),
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
