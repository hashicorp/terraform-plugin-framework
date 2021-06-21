package proto6

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestRequestType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		req         *tfprotov6.ApplyResourceChangeRequest
		typ         tftypes.Type
		isCreate    bool
		isUpdate    bool
		isDestroy   bool
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
			typ:       reqType,
			isCreate:  true,
			isUpdate:  false,
			isDestroy: false,
		},
		"set-prior-state-set-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &setDV,
			},
			typ:       reqType,
			isCreate:  false,
			isUpdate:  true,
			isDestroy: false,
		},
		"set-prior-state-null-planned-state": {
			req: &tfprotov6.ApplyResourceChangeRequest{
				TypeName:     "foo",
				PriorState:   &setDV,
				PlannedState: &nullDV,
			},
			typ:       reqType,
			isCreate:  false,
			isUpdate:  false,
			isDestroy: true,
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
			if tc.isCreate != got {
				t.Errorf("Expected IsCreate to return %v, got %v", tc.isCreate, got)
				return
			}

			got, err = IsUpdate(context.Background(), tc.req, tc.typ)
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
			if tc.isUpdate != got {
				t.Errorf("Expected IsUpdate to return %v, got %v", tc.isUpdate, got)
				return
			}

			got, err = IsDestroy(context.Background(), tc.req, tc.typ)
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
			if tc.isDestroy != got {
				t.Errorf("Expected IsDestroy to return %v, got %v", tc.isDestroy, got)
				return
			}
		})
	}
}
