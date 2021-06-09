package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var testSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"foo": {
			Type:     types.StringType,
			Required: true,
		},
		"bar": {
			Type: types.ListType{
				ElemType: types.StringType,
			},
			Required: true,
		},
		"disks": {
			Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
				"id": {
					Type:     types.StringType,
					Required: true,
				},
				"delete_with_instance": {
					Type:     types.BoolType,
					Optional: true,
				},
			}, schema.ListNestedAttributesOptions{}),
			Optional: true,
			Computed: true,
		},
	},
}

var diskElementType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":                   tftypes.String,
		"delete_with_instance": tftypes.Bool,
	},
}

var testState = State{
	Raw: tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"foo": tftypes.String,
			"bar": tftypes.List{ElementType: tftypes.String},
			"disks": tftypes.List{
				ElementType: diskElementType,
			},
		},
	}, map[string]tftypes.Value{
		"foo": tftypes.NewValue(tftypes.String, "hello, world"),
		"bar": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.String,
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "red"),
			tftypes.NewValue(tftypes.String, "blue"),
			tftypes.NewValue(tftypes.String, "green"),
		}),
		"disks": tftypes.NewValue(tftypes.List{
			ElementType: diskElementType,
		}, []tftypes.Value{
			tftypes.NewValue(diskElementType, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "disk0"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			}),
			tftypes.NewValue(diskElementType, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "disk1"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
			}),
		}),
	}),
	Schema: testSchema,
}

func TestStateGet(t *testing.T) {
	t.Logf("+%v", testSchema.AttributeType())
	type myType struct {
		Foo   types.String `tfsdk:"foo"`
		Bar   types.List   `tfsdk:"bar"`
		Disks []struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"disks"`
	}
	var val myType
	err := testState.Get(context.Background(), &val)
	if err != nil {
		t.Fatalf("Error running Get: %s", err)
	}
	expected := myType{
		Foo: types.String{Value: "hello, world"},
		Bar: types.List{
			ElemType: types.StringType,
			Elems: []attr.Value{
				types.String{Value: "red"},
				types.String{Value: "blue"},
				types.String{Value: "green"},
			},
		},
		Disks: []struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		}{
			{
				ID:                 "disk0",
				DeleteWithInstance: true,
			},
			{
				ID:                 "disk1",
				DeleteWithInstance: false,
			},
		},
	}
	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestStateGetAttribute(t *testing.T) {
	fooVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("foo"))
	if err != nil {
		t.Errorf("Error running GetAttribute for foo: %s", err)
	}
	foo, ok := fooVal.(types.String)
	if !ok {
		t.Errorf("expected foo to have type String, but it was %T", fooVal)
	}
	if foo.Unknown {
		t.Error("Expected Foo to be known")
	}
	if foo.Null {
		t.Error("Expected Foo to be non-null")
	}
	if foo.Value != "hello, world" {
		t.Errorf("Expected Foo to be %q, got %q", "hello, world", foo.Value)
	}

	barVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("bar"))
	if err != nil {
		t.Errorf("Error running GetAttribute for bar: %s", err)
	}
	bar, ok := barVal.(types.List)
	if !ok {
		t.Errorf("expected bar to have type List, but it was %T", barVal)
	}
	if bar.Unknown {
		t.Error("Expected Bar to be known")
	}
	if bar.Null {
		t.Errorf("Expected Bar to be non-null")
	}
	if len(bar.Elems) != 3 {
		t.Errorf("Expected Bar to have 3 elements, had %d", len(bar.Elems))
	}
	if bar.Elems[0].(types.String).Value != "red" {
		t.Errorf("Expected Bar's first element to be %q, got %q", "red", bar.Elems[0].(types.String).Value)
	}
	if bar.Elems[1].(types.String).Value != "blue" {
		t.Errorf("Expected Bar's second element to be %q, got %q", "blue", bar.Elems[1].(types.String).Value)
	}
	if bar.Elems[2].(types.String).Value != "green" {
		t.Errorf("Expected Bar's third element to be %q, got %q", "green", bar.Elems[2].(types.String).Value)
	}

	bar0Val, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("bar").WithElementKeyInt(0))
	if err != nil {
		t.Errorf("Error running GetAttribute for bar[0]: %s", err)
	}
	bar0, ok := bar0Val.(types.String)
	if !ok {
		t.Errorf("expected bar[0] to have type String, but it was %T", bar0Val)
	}
	if bar0.Unknown {
		t.Error("expected bar[0] to be known")
	}
	if bar0.Null {
		t.Error("expected bar[0] to be non-null")
	}
	if bar0.Value != "red" {
		t.Errorf("Expected bar[0] to be %q, got %q", "red", bar0.Value)
	}

	bar1Val, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("bar").WithElementKeyInt(1))
	if err != nil {
		t.Errorf("Error running GetAttribute for bar[1]: %s", err)
	}
	bar1, ok := bar1Val.(types.String)
	if !ok {
		t.Errorf("expected bar[1] to have type String, but it was %T", bar1Val)
	}
	if bar1.Unknown {
		t.Error("expected bar[1] to be known")
	}
	if bar1.Null {
		t.Error("expected bar[1] to be non-null")
	}
	if bar1.Value != "blue" {
		t.Errorf("Expected bar[1] to be %q, got %q", "red", bar1.Value)
	}

	bar2Val, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("bar").WithElementKeyInt(2))
	if err != nil {
		t.Errorf("Error running GetAttribute for bar[2]: %s", err)
	}
	bar2, ok := bar2Val.(types.String)
	if !ok {
		t.Errorf("expected bar[2] to have type String, but it was %T", bar2Val)
	}
	if bar2.Unknown {
		t.Error("expected bar[2] to be known")
	}
	if bar2.Null {
		t.Error("expected bar[2] to be non-null")
	}
	if bar2.Value != "green" {
		t.Errorf("Expected bar[2] to be %q, got %q", "red", bar2.Value)
	}
}
