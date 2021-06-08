package reflect_test

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBuildValue_unhandledNull(t *testing.T) {
	t.Parallel()

	var s string
	_, err := refl.BuildValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: unhandled null value`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestBuildValue_unhandledUnknown(t *testing.T) {
	t.Parallel()

	var s string
	_, err := refl.BuildValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: unhandled unknown value`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestOutOfString(t *testing.T) {
	expectedVal := types.String{
		Value: "mystring",
	}
	actualVal, actualType, err := refl.OutOf(context.Background(), reflect.ValueOf("mystring"), refl.OutOfOptions{
		Strings: types.StringType,
	}, tftypes.NewAttributePath())
	if err != nil {
		t.Fatal(err)
	}
	expectedType := types.StringType
	if !expectedVal.Equal(actualVal) {
		t.Fatalf("fail: got %+v, wanted %+v", actualVal, expectedVal)
	}
	if actualType != expectedType {
		t.Fatalf("fail: got %+v, wanted %+v", actualType, expectedType)
	}
}

func TestOutOfStruct(t *testing.T) {
	type disk struct {
		Name string `tfsdk:"name"`
		// bool
	}
	disk1 := disk{
		Name: "myfirstdisk",
	}

	actualVal, actualType, err := refl.OutOf(context.Background(), reflect.ValueOf(disk1), refl.OutOfOptions{
		Structs: types.ObjectType{},
		Strings: types.StringType,
	}, tftypes.NewAttributePath())
	if err != nil {
		t.Fatal(err)
	}

	expectedVal := types.Object{
		Attrs: map[string]attr.Value{
			"name": types.String{Value: "myfirstdisk"},
		},
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
		},
	}
	expectedType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
		},
	}

	if !expectedVal.Equal(actualVal) {
		t.Fatalf("fail: got %+v, wanted %+v", actualVal, expectedVal)
	}

	if !reflect.DeepEqual(expectedType, actualType) {
		t.Fatalf("fail: got %+v, wanted %+v", actualType, expectedType)
	}
}

func TestOutOfBool(t *testing.T) {
	// the rare exhaustive test
	cases := []struct {
		val         reflect.Value
		expectedVal attr.Value
	}{
		{
			reflect.ValueOf(true),
			types.Bool{
				Value: true,
			},
		},
		{
			reflect.ValueOf(false),
			types.Bool{
				Value: false,
			},
		},
	}

	expectedType := types.BoolType

	for _, tc := range cases {
		actualVal, actualType, err := refl.OutOf(context.Background(), tc.val, refl.OutOfOptions{Bools: types.BoolType}, tftypes.NewAttributePath())
		if err != nil {
			t.Fatal(err)
		}

		if !tc.expectedVal.Equal(actualVal) {
			t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expectedVal)
		}

		if !reflect.DeepEqual(expectedType, actualType) {
			t.Fatalf("fail: got %+v, wanted %+v", actualType, expectedType)
		}
	}
}

func TestOutOfInteger(t *testing.T) {
	cases := []struct {
		val         reflect.Value
		expectedVal attr.Value
	}{
		{
			reflect.ValueOf(0),

			types.Number{
				Value: big.NewFloat(0),
			},
		},
		{
			reflect.ValueOf(1),
			types.Number{
				Value: big.NewFloat(1),
			},
		},
		{
			reflect.ValueOf(big.MaxExp),
			types.Number{
				Value: big.NewFloat(big.MaxExp),
			},
		},
		{
			reflect.ValueOf(big.MinExp),
			types.Number{
				Value: big.NewFloat(big.MinExp),
			},
		},
	}

	expectedType := types.NumberType

	for _, tc := range cases {
		actualVal, actualType, err := refl.OutOf(context.Background(), tc.val, refl.OutOfOptions{Integers: types.NumberType}, tftypes.NewAttributePath())
		if err != nil {
			t.Fatal(err)
		}

		if !tc.expectedVal.Equal(actualVal) {
			t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expectedVal)
		}

		if !reflect.DeepEqual(expectedType, actualType) {
			t.Fatalf("fail: got %+v, wanted %+v", actualType, expectedType)
		}
	}
}
