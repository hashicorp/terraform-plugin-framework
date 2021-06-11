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
		"boot_disk": {
			Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
				"id": {
					Type:     types.StringType,
					Required: true,
				},
				"delete_with_instance": {
					Type: types.BoolType,
				},
			}),
		},
		"scratch_disk": {
			Type: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"interface": types.StringType,
				},
			},
			Optional: true,
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
			"boot_disk": diskElementType,
			"scratch_disk": tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"interface": tftypes.String,
				},
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
		"boot_disk": tftypes.NewValue(diskElementType, map[string]tftypes.Value{
			"id":                   tftypes.NewValue(tftypes.String, "bootdisk"),
			"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
		}),
		"scratch_disk": tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"interface": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"interface": tftypes.NewValue(tftypes.String, "SCSI"),
		}),
	}),
	Schema: testSchema,
}

func TestStateGet(t *testing.T) {
	type testStateStructType struct {
		Foo   types.String `tfsdk:"foo"`
		Bar   types.List   `tfsdk:"bar"`
		Disks []struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"disks"`
		BootDisk struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"boot_disk"`
		ScratchDisk struct {
			Interface string `tfsdk:"interface"`
		} `tfsdk:"scratch_disk"`
	}
	var val testStateStructType
	err := testState.Get(context.Background(), &val)
	if err != nil {
		t.Fatalf("Error running Get: %s", err)
	}
	expected := testStateStructType{
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
		BootDisk: struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		}{
			ID:                 "bootdisk",
			DeleteWithInstance: true,
		},
		ScratchDisk: struct {
			Interface string `tfsdk:"interface"`
		}{
			Interface: "SCSI",
		},
	}
	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestStateGetAttribute_primitive(t *testing.T) {
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
}

func TestStateGetAttribute_list(t *testing.T) {
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

func TestStateGetAttribute_nestedlist(t *testing.T) {
	disksVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("disks"))
	if err != nil {
		t.Errorf("Error running GetAttribute for foo: %s", err)
	}

	disks, ok := disksVal.(types.List)
	if !ok {
		t.Errorf("expected disks to have type List, but it was %T", disksVal)
	}
	if disks.Unknown {
		t.Error("Expected disks to be known")
	}
	if disks.Null {
		t.Errorf("Expected disks to be non-null")
	}
	if len(disks.Elems) != 2 {
		t.Errorf("Expected disks to have 2 elements, had %d", len(disks.Elems))
	}

	disk0, ok := disks.Elems[0].(types.Object)
	if !ok {
		t.Errorf("expected disks[0] to have type Object, but it was %T", disks.Elems[0])
	}
	if disk0.Unknown {
		t.Errorf("Expected disks[0] to be known")
	}
	if disk0.Null {
		t.Errorf("expected disks[0] to be non-null")
	}

	disk0Id, ok := disk0.Attrs["id"].(types.String)
	if !ok {
		t.Errorf("expected disks[0].id to have type String, but it was %T", disk0.Attrs["id"])
	}
	if disk0Id.Unknown {
		t.Errorf("expected disks[0].id to be known")
	}
	if disk0Id.Null {
		t.Errorf("expected disks[0].id to be non-null")
	}
	if disk0Id.Value != "disk0" {
		t.Errorf("expected disks[0].id to be %q, got %q", "disk0", disk0Id.Value)
	}

	disk1, ok := disks.Elems[1].(types.Object)
	if !ok {
		t.Errorf("expected disks[1] to have type Object, but it was %T", disks.Elems[1])
	}
	if disk1.Unknown {
		t.Errorf("Expected disks[1] to be known")
	}
	if disk1.Null {
		t.Errorf("expected disks[1] to be non-null")
	}

	disk1Id, ok := disk0.Attrs["id"].(types.String)
	if !ok {
		t.Errorf("expected disks[1].id to have type String, but it was %T", disk1.Attrs["id"])
	}
	if disk1Id.Unknown {
		t.Errorf("expected disks[1].id to be known")
	}
	if disk1Id.Null {
		t.Errorf("expected disks[1].id to be non-null")
	}
	if disk1Id.Value != "disk0" {
		t.Errorf("expected disks[1].id to be %q, got %q", "disk0", disk1Id.Value)
	}
}

func TestStateGetAttribute_nestedsingle(t *testing.T) {
	bootDiskVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("boot_disk"))
	if err != nil {
		t.Errorf("Error running GetAttribute for foo: %s", err)
	}

	bootDisk, ok := bootDiskVal.(types.Object)
	if !ok {
		t.Errorf("expected boot_disk to have type Object, but it was %T", bootDiskVal)
	}
	if bootDisk.Unknown {
		t.Error("expected bootDisk to be known")
	}
	if bootDisk.Null {
		t.Errorf("expected bootDisk to be non-null")
	}

	bootDiskID, ok := bootDisk.Attrs["id"].(types.String)
	if !ok {
		t.Errorf("expected bootDisk.Attrs[\"id\"] to have type String, but it was %T", bootDisk.Attrs["id"])
	}
	if bootDiskID.Unknown {
		t.Errorf("expected bootDisk.Attrs[\"id\"] to be known")
	}
	if bootDiskID.Null {
		t.Errorf("expected bootDisk.Attrs[\"id\"] to be non-null")
	}
	if bootDiskID.Value != "bootdisk" {
		t.Errorf("expected bootDisk.Attrs[\"id\"] to be %q, got %q", "bootdisk", bootDiskID.Value)
	}

	bootDiskDelete, ok := bootDisk.Attrs["delete_with_instance"].(types.Bool)
	if !ok {
		t.Errorf("expected bootDisk.Attrs[\"delete_with_instance\"] to have type Bool, but it was %T", bootDisk.Attrs["delete_with_instance"])
	}
	if bootDiskDelete.Unknown {
		t.Errorf("expected bootDisk.Attrs[\"delete_with_instance\"] to be known")
	}
	if bootDiskDelete.Null {
		t.Errorf("expected bootDisk.Attrs[\"delete_with_instance\"] to be non-null")
	}
	if bootDiskDelete.Value != true {
		t.Errorf("expected bootDisk.Attrs[\"delete_with_instance\"] to be %t, got %t", true, bootDiskDelete.Value)
	}
}

func TestStateGetAttribute_object(t *testing.T) {
	scratchDiskVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("scratch_disk"))
	if err != nil {
		t.Errorf("error running GetAttribute for scratch_disk: %s", err)
	}
	scratchDisk, ok := scratchDiskVal.(types.Object)
	if !ok {
		t.Errorf("expected scratchDisk to have type Object, but it was %T", scratchDiskVal)
	}
	if scratchDisk.Unknown {
		t.Error("expected scratchDisk to be known")
	}
	if scratchDisk.Null {
		t.Error("expected scratchDisk to be non-null")
	}

	scratchDiskInterface, ok := scratchDisk.Attrs["interface"].(types.String)
	if !ok {
		t.Errorf("expected scratchDisk[\"interface\"] to have type String, but it was %T", scratchDisk.Attrs["interface"])
	}
	if scratchDiskInterface.Unknown {
		t.Error("expected scratchDiskInterface to be known")
	}
	if scratchDiskInterface.Null {
		t.Error("expected scratchDiskInterface to be non-null")
	}
	if scratchDiskInterface.Value != "SCSI" {
		t.Errorf("expected scratchDiskInterface to be %q, got %q", "SCSI", scratchDiskInterface.Value)
	}

	// now get the value directly
	sdInterfaceVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"))
	if err != nil {
		t.Errorf("error running GetAttribute for scratch_disk.interface: %s", err)
	}
	sdInterface, ok := sdInterfaceVal.(types.String)
	if !ok {
		t.Errorf("expected scratchDiskInterface to have type String, but it was %T", sdInterfaceVal)
	}
	if sdInterface.Unknown {
		t.Error("expected scratchDiskInterface to be known")
	}
	if sdInterface.Null {
		t.Error("expected scratchDiskInterface to be non-null")
	}
	if sdInterface.Value != "SCSI" {
		t.Errorf("expected scratchDiskInterface to be %q, got %q", "SCSI", sdInterface.Value)
	}
}

func TestStateSet(t *testing.T) {
	state := State{
		// Raw:    tftypes.Value{},
		Schema: testSchema,
	}

	type newStateType struct {
		Foo   string   `tfsdk:"foo"`
		Bar   []string `tfsdk:"bar"`
		Disks []struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"disks"`
		BootDisk struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"boot_disk"`
		ScratchDisk struct {
			Interface string `tfsdk:"interface"`
		} `tfsdk:"scratch_disk"`
	}

	err := state.Set(context.Background(), newStateType{
		Foo: "hello, world",
		Bar: []string{"red", "blue", "green"},
		Disks: []struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		}{{
			ID:                 "disk0",
			DeleteWithInstance: true,
		},
			{
				ID:                 "disk1",
				DeleteWithInstance: false,
			}},
		BootDisk: struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		}{
			ID:                 "bootdisk",
			DeleteWithInstance: true,
		},
		ScratchDisk: struct {
			Interface string `tfsdk:"interface"`
		}{
			Interface: "SCSI",
		},
	})
	if err != nil {
		t.Fatalf("error setting state: %s", err)
	}

	actual := state.Raw
	expected := testState.Raw
	t.Logf("actual: %+v", actual.String())

	if !expected.Equal(actual) {
		t.Fatalf("unexpected diff in state.Raw (+wanted, -got): %s", cmp.Diff(actual, expected))
	}
}

func TestStateGetSetInverse(t *testing.T) {
	t.Skip()
	// test that Get and Set are inverses of each other
	type testStateStructType struct {
		Foo   types.String `tfsdk:"foo"`
		Bar   types.List   `tfsdk:"bar"`
		Disks []struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"disks"`
		BootDisk struct {
			ID                 string `tfsdk:"id"`
			DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
		} `tfsdk:"boot_disk"`
		ScratchDisk struct {
			Interface string `tfsdk:"interface"`
		} `tfsdk:"scratch_disk"`
	}
	var val testStateStructType
	err := testState.Get(context.Background(), &val)
	if err != nil {
		t.Fatalf("Error running Get: %s", err)
	}

	newState := State{
		Schema: testSchema,
	}

	err = newState.Set(context.Background(), val)
	if err != nil {
		t.Fatalf("error setting state: %s", err)
	}

	t.Log(newState)

	if diff := cmp.Diff(testState, newState); diff != "" {
		t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
	}
}
