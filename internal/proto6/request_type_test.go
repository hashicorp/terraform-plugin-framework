package proto6

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestIsCreate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		req         *tfprotov6.ApplyResourceChangeRequest
		typ         tftypes.Type
		expected    bool
		expectedErr string
	}

	reqType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}

	setDV, err := tfprotov6.NewDynamicValue(reqType, tftypes.NewValue(reqType, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello, world"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}))
	if err != nil {
		t.Errorf("Unexpected error creating set dynamic value: %s", err)
		return
	}

	nullDV, err := tfprotov6.NewDynamicValue(reqType, tftypes.NewValue(reqType, nil))
	if err != nil {
		t.Errorf("Unexpected error creating null dynamic value: %s", err)
		return
	}

	tests := map[string]testCase{
		"null-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &nullDV,
				PlannedState: &setDV,
			},
			typ:      reqType,
			expected: true,
		},
		"set-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &setDV,
			},
			typ:      reqType,
			expected: false,
		},
		"set-prior-state-null-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &nullDV,
			},
			typ:      reqType,
			expected: false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := IsCreate(context.Background(), tc.req, tc.typ)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}
			if err == nil && tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}
			if tc.expected != got {
				t.Errorf("Expected %v, got %v", tc.expected, got)
				return
			}
		})
	}
}

func TestIsUpdate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		req         *tfprotov6.ApplyResourceChangeRequest
		typ         tftypes.Type
		expected    bool
		expectedErr string
	}

	reqType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}

	setDV, err := tfprotov6.NewDynamicValue(reqType, tftypes.NewValue(reqType, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello, world"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}))
	if err != nil {
		t.Errorf("Unexpected error creating set dynamic value: %s", err)
		return
	}

	nullDV, err := tfprotov6.NewDynamicValue(reqType, tftypes.NewValue(reqType, nil))
	if err != nil {
		t.Errorf("Unexpected error creating null dynamic value: %s", err)
		return
	}

	tests := map[string]testCase{
		"null-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &nullDV,
				PlannedState: &setDV,
			},
			typ:      reqType,
			expected: false,
		},
		"set-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &setDV,
			},
			typ:      reqType,
			expected: true,
		},
		"set-prior-state-null-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &nullDV,
			},
			typ:      reqType,
			expected: false,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := IsUpdate(context.Background(), tc.req, tc.typ)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}
			if err == nil && tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}
			if tc.expected != got {
				t.Errorf("Expected %v, got %v", tc.expected, got)
				return
			}
		})
	}
}

func TestIsDestroy(t *testing.T) {
	t.Parallel()

	type testCase struct {
		req         *tfprotov6.ApplyResourceChangeRequest
		typ         tftypes.Type
		expected    bool
		expectedErr string
	}

	reqType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"a": tftypes.String,
			"b": tftypes.Number,
			"c": tftypes.Bool,
		},
	}

	setDV, err := tfprotov6.NewDynamicValue(reqType, tftypes.NewValue(reqType, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "hello, world"),
		"b": tftypes.NewValue(tftypes.Number, 123),
		"c": tftypes.NewValue(tftypes.Bool, true),
	}))
	if err != nil {
		t.Errorf("Unexpected error creating set dynamic value: %s", err)
		return
	}

	nullDV, err := tfprotov6.NewDynamicValue(reqType, tftypes.NewValue(reqType, nil))
	if err != nil {
		t.Errorf("Unexpected error creating null dynamic value: %s", err)
		return
	}

	tests := map[string]testCase{
		"null-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &nullDV,
				PlannedState: &setDV,
			},
			typ:      reqType,
			expected: false,
		},
		"set-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &setDV,
			},
			typ:      reqType,
			expected: false,
		},
		"set-prior-state-null-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &nullDV,
			},
			typ:      reqType,
			expected: true,
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := IsDestroy(context.Background(), tc.req, tc.typ)
			if err != nil {
				if tc.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if err.Error() != tc.expectedErr {
					t.Errorf("Expected error to be %q, got %q", tc.expectedErr, err.Error())
					return
				}
				// got expected error
				return
			}
			if err == nil && tc.expectedErr != "" {
				t.Errorf("Expected error to be %q, got nil", tc.expectedErr)
				return
			}
			if tc.expected != got {
				t.Errorf("Expected %v, got %v", tc.expected, got)
				return
			}
		})
	}
}
