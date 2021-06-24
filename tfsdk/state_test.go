package tfsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var allowAllUnexported = cmp.Exporter(func(reflect.Type) bool { return true })

// schema used for all tests
var testSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": {
			Type:     types.StringType,
			Required: true,
		},
		"machine_type": {
			Type: types.StringType,
		},
		"tags": {
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

// element type for the "disks" attribute, which is a list of disks.
// only used in "disks"
var diskElementType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":                   tftypes.String,
		"delete_with_instance": tftypes.Bool,
	},
}

// state used for all tests
func makeTestState() State {
	return State{
		Raw: tftypes.NewValue(tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":         tftypes.String,
				"machine_type": tftypes.String,
				"tags":         tftypes.List{ElementType: tftypes.String},
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
			"name":         tftypes.NewValue(tftypes.String, "hello, world"),
			"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
			"tags": tftypes.NewValue(tftypes.List{
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
}

// struct type used for Get() calls. note the mix of framework types and
// native Go types.
type testStateStructType struct {
	Name        types.String `tfsdk:"name"`
	MachineType string       `tfsdk:"machine_type"`
	Tags        types.List   `tfsdk:"tags"`
	Disks       []struct {
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

func TestStateGet(t *testing.T) {
	testState := makeTestState()
	var val testStateStructType
	err := testState.Get(context.Background(), &val)
	if err != nil {
		t.Fatalf("Error running Get: %s", err)
	}
	expected := testStateStructType{
		Name:        types.String{Value: "hello, world"},
		MachineType: "e2-medium",
		Tags: types.List{
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
	testState := makeTestState()
	nameVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))
	if err != nil {
		t.Errorf("Error running GetAttribute for name: %s", err)
	}
	name, ok := nameVal.(types.String)
	if !ok {
		t.Errorf("expected name to have type String, but it was %T", nameVal)
	}
	if name.Unknown {
		t.Error("Expected Name to be known")
	}
	if name.Null {
		t.Error("Expected Name to be non-null")
	}
	if name.Value != "hello, world" {
		t.Errorf("Expected Name to be %q, got %q", "hello, world", name.Value)
	}
}

func TestStateGetAttribute_list(t *testing.T) {
	testState := makeTestState()
	tagsVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("tags"))
	if err != nil {
		t.Errorf("Error running GetAttribute for tags: %s", err)
	}
	tags, ok := tagsVal.(types.List)
	if !ok {
		t.Errorf("expected tags to have type List, but it was %T", tagsVal)
	}
	if tags.Unknown {
		t.Error("Expected Tags to be known")
	}
	if tags.Null {
		t.Errorf("Expected Tags to be non-null")
	}
	if len(tags.Elems) != 3 {
		t.Errorf("Expected Tags to have 3 elements, had %d", len(tags.Elems))
	}
	if tags.Elems[0].(types.String).Value != "red" {
		t.Errorf("Expected Tags's first element to be %q, got %q", "red", tags.Elems[0].(types.String).Value)
	}
	if tags.Elems[1].(types.String).Value != "blue" {
		t.Errorf("Expected Tags's second element to be %q, got %q", "blue", tags.Elems[1].(types.String).Value)
	}
	if tags.Elems[2].(types.String).Value != "green" {
		t.Errorf("Expected Tags's third element to be %q, got %q", "green", tags.Elems[2].(types.String).Value)
	}

	tags0Val, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("tags").WithElementKeyInt(0))
	if err != nil {
		t.Errorf("Error running GetAttribute for tags[0]: %s", err)
	}
	tags0, ok := tags0Val.(types.String)
	if !ok {
		t.Errorf("expected tags[0] to have type String, but it was %T", tags0Val)
	}
	if tags0.Unknown {
		t.Error("expected tags[0] to be known")
	}
	if tags0.Null {
		t.Error("expected tags[0] to be non-null")
	}
	if tags0.Value != "red" {
		t.Errorf("Expected tags[0] to be %q, got %q", "red", tags0.Value)
	}

	tags1Val, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("tags").WithElementKeyInt(1))
	if err != nil {
		t.Errorf("Error running GetAttribute for tags[1]: %s", err)
	}
	tags1, ok := tags1Val.(types.String)
	if !ok {
		t.Errorf("expected tags[1] to have type String, but it was %T", tags1Val)
	}
	if tags1.Unknown {
		t.Error("expected tags[1] to be known")
	}
	if tags1.Null {
		t.Error("expected tags[1] to be non-null")
	}
	if tags1.Value != "blue" {
		t.Errorf("Expected tags[1] to be %q, got %q", "red", tags1.Value)
	}

	tags2Val, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("tags").WithElementKeyInt(2))
	if err != nil {
		t.Errorf("Error running GetAttribute for tags[2]: %s", err)
	}
	tags2, ok := tags2Val.(types.String)
	if !ok {
		t.Errorf("expected tags[2] to have type String, but it was %T", tags2Val)
	}
	if tags2.Unknown {
		t.Error("expected tags[2] to be known")
	}
	if tags2.Null {
		t.Error("expected tags[2] to be non-null")
	}
	if tags2.Value != "green" {
		t.Errorf("Expected tags[2] to be %q, got %q", "red", tags2.Value)
	}
}

func TestStateGetAttribute_nestedlist(t *testing.T) {
	testState := makeTestState()
	disksVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("disks"))
	if err != nil {
		t.Errorf("Error running GetAttribute for name: %s", err)
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
	testState := makeTestState()
	bootDiskVal, err := testState.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("boot_disk"))
	if err != nil {
		t.Errorf("Error running GetAttribute for name: %s", err)
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
	testState := makeTestState()
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
		Raw:    tftypes.Value{},
		Schema: testSchema,
	}

	type newStateType struct {
		Name        string   `tfsdk:"name"`
		MachineType string   `tfsdk:"machine_type"`
		Tags        []string `tfsdk:"tags"`
		Disks       []struct {
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
		Name:        "hello, world",
		MachineType: "e2-medium",
		Tags:        []string{"red", "blue", "green"},
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
	testState := makeTestState()
	expected := testState.Raw

	if !expected.Equal(actual) {
		t.Fatalf("unexpected diff in state.Raw (+wanted, -got): %s", cmp.Diff(actual, expected))
	}
}

// test that Get and Set are inverses of each other
func TestStateGetSetInverse(t *testing.T) {
	testState := makeTestState()
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

	if diff := cmp.Diff(testState, newState, allowAllUnexported); diff != "" {
		t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestStateSetAttribute(t *testing.T) {
	testState := makeTestState()

	// set a simple string attribute
	err := testState.SetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), "newname")
	if err != nil {
		t.Fatal(err)
	}

	// set an entire list
	err = testState.SetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("tags"), []string{"one", "two"})
	if err != nil {
		t.Fatal(err)
	}

	// set a list item
	err = testState.SetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(1), struct {
		ID                 string `tfsdk:"id"`
		DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
	}{
		ID:                 "mynewdisk",
		DeleteWithInstance: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// set an object attribute
	err = testState.SetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"), "NVME")
	if err != nil {
		t.Fatal(err)
	}

	expectedRawState := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name":         tftypes.String,
			"machine_type": tftypes.String,
			"tags":         tftypes.List{ElementType: tftypes.String},
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
		"name":         tftypes.NewValue(tftypes.String, "newname"),
		"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
		"tags": tftypes.NewValue(tftypes.List{
			ElementType: tftypes.String,
		}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "one"),
			tftypes.NewValue(tftypes.String, "two"),
		}),
		"disks": tftypes.NewValue(tftypes.List{
			ElementType: diskElementType,
		}, []tftypes.Value{
			tftypes.NewValue(diskElementType, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "disk0"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
			}),
			tftypes.NewValue(diskElementType, map[string]tftypes.Value{
				"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
				"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
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
			"interface": tftypes.NewValue(tftypes.String, "NVME"),
		}),
	})

	if diff := cmp.Diff(expectedRawState, testState.Raw, allowAllUnexported); diff != "" {
		t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
	}
}
