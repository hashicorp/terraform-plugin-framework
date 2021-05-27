package reflect_test

import (
	"context"
	goReflect "reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestOutOfString(t *testing.T) {
	expectedVal := types.String{
		Value: "mystring",
	}
	actualVal, actualType, err := reflect.OutOf(context.Background(), "mystring", reflect.OutOfOptions{
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

	actualVal, actualType, err := reflect.OutOf(context.Background(), disk1, reflect.OutOfOptions{
		Structs: types.ObjectType{},
		Strings: types.StringType,
	}, tftypes.NewAttributePath())
	if err != nil {
		t.Fatal(err)
	}

	expectedVal := types.Object{
		Attributes: map[string]attr.Value{
			"name": types.String{Value: "myfirstdisk"},
		},
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}
	expectedType := types.ObjectType{
		AttributeTypes: map[string]attr.Type{
			"name": types.StringType,
		},
	}

	if !expectedVal.Equal(actualVal) {
		t.Fatalf("fail: got %+v, wanted %+v", actualVal, expectedVal)
	}

	if !goReflect.DeepEqual(expectedType, actualType) {

		t.Fatalf("fail: got %+v, wanted %+v", actualType, expectedType)
	}
}
