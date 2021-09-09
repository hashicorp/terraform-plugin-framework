package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStateGet(t *testing.T) {
	t.Parallel()

	type testStateGetData struct {
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

	type testCase struct {
		state         State
		expected      testStateGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"complex": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":         tftypes.String,
						"machine_type": tftypes.String,
						"tags":         tftypes.List{ElementType: tftypes.String},
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"boot_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
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
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"boot_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
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
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"boot_disk": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
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
				},
			},
			expected: testStateGetData{
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
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testStateGetData

			diags := tc.state.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateGet_testTypes(t *testing.T) {
	t.Parallel()

	type testStateGetDataTestTypes struct {
		Name        testtypes.String `tfsdk:"name"`
		MachineType string           `tfsdk:"machine_type"`
		Tags        types.List       `tfsdk:"tags"`
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

	type testCase struct {
		state         State
		expected      testStateGetDataTestTypes
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"AttrTypeWithValidateError": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":         tftypes.String,
						"machine_type": tftypes.String,
						"tags":         tftypes.List{ElementType: tftypes.String},
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"boot_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"name":         tftypes.NewValue(tftypes.String, "namevalue"),
					"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
					"tags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"boot_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
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
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"boot_disk": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
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
				},
			},
			expected: testStateGetDataTestTypes{
				Name:        testtypes.String{String: types.String{Value: ""}, CreatedBy: testtypes.StringTypeWithValidateError{}},
				MachineType: "",
				Tags:        types.List{},
				Disks:       nil,
				BootDisk: struct {
					ID                 string `tfsdk:"id"`
					DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
				}{
					ID:                 "",
					DeleteWithInstance: false,
				},
				ScratchDisk: struct {
					Interface string `tfsdk:"interface"`
				}{
					Interface: "",
				},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":         tftypes.String,
						"machine_type": tftypes.String,
						"tags":         tftypes.List{ElementType: tftypes.String},
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"boot_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"name":         tftypes.NewValue(tftypes.String, "namevalue"),
					"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
					"tags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"boot_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
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
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
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
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"boot_disk": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
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
				},
			},
			expected: testStateGetDataTestTypes{
				Name:        testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
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
			},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testStateGetDataTestTypes

			diags := tc.state.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				for _, diag := range diags {
					t.Log(diag)
				}
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         State
		path          *tftypes.AttributePath
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"primitive": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "hello, world"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path:     tftypes.NewAttributePath().WithAttributeName("name"),
			expected: types.String{Value: "hello, world"},
		},
		"list": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags": tftypes.List{ElementType: tftypes.String},
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags"),
			expected: types.List{
				Elems: []attr.Value{
					types.String{Value: "red"},
					types.String{Value: "blue"},
					types.String{Value: "green"},
				},
				ElemType: types.StringType,
			},
		},
		"nested-list": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks"),
			expected: types.List{
				Elems: []attr.Value{
					types.Object{
						Attrs: map[string]attr.Value{
							"delete_with_instance": types.Bool{Value: true},
							"id":                   types.String{Value: "disk0"},
						},
						AttrTypes: map[string]attr.Type{
							"delete_with_instance": types.BoolType,
							"id":                   types.StringType,
						},
					},
					types.Object{
						Attrs: map[string]attr.Value{
							"delete_with_instance": types.Bool{Value: false},
							"id":                   types.String{Value: "disk1"},
						},
						AttrTypes: map[string]attr.Type{
							"delete_with_instance": types.BoolType,
							"id":                   types.StringType,
						},
					},
				},
				ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"delete_with_instance": types.BoolType,
						"id":                   types.StringType,
					},
				},
			},
		},
		"nested-single": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"boot_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
				}, map[string]tftypes.Value{
					"boot_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "bootdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"boot_disk": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type: types.BoolType,
								},
							}),
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("boot_disk"),
			expected: types.Object{
				Attrs: map[string]attr.Value{
					"delete_with_instance": types.Bool{Value: true},
					"id":                   types.String{Value: "bootdisk"},
				},
				AttrTypes: map[string]attr.Type{
					"delete_with_instance": types.BoolType,
					"id":                   types.StringType,
				},
			},
		},
		"object": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
					},
				}, map[string]tftypes.Value{
					"scratch_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"interface": tftypes.NewValue(tftypes.String, "SCSI"),
					}),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
								},
							},
							Optional: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("scratch_disk"),
			expected: types.Object{
				Attrs: map[string]attr.Value{
					"interface": types.String{Value: "SCSI"},
				},
				AttrTypes: map[string]attr.Type{
					"interface": types.StringType,
				},
			},
		},
		"AttrTypeWithValidateError": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("name"),
			expected:      nil,
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			path:          tftypes.NewAttributePath().WithAttributeName("name"),
			expected:      testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, diags := tc.state.GetAttribute(context.Background(), tc.path)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         State
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"write": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"machine_type": {
							Type:     types.StringType,
							Required: true,
						},
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				MachineType string `tfsdk:"machine_type"`
				Name        string `tfsdk:"name"`
			}{
				MachineType: "e2-medium",
				Name:        "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"machine_type": tftypes.String,
					"name":         tftypes.String,
				},
			}, map[string]tftypes.Value{
				"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
				"name":         tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"overwrite": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"machine_type": tftypes.String,
						"name":         tftypes.String,
					},
				}, map[string]tftypes.Value{
					"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
					"name":         tftypes.NewValue(tftypes.String, "oldvalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"machine_type": {
							Type:     types.StringType,
							Required: true,
						},
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				MachineType string `tfsdk:"machine_type"`
				Name        string `tfsdk:"name"`
			}{
				MachineType: "e2-medium",
				Name:        "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"machine_type": tftypes.String,
					"name":         tftypes.String,
				},
			}, map[string]tftypes.Value{
				"machine_type": tftypes.NewValue(tftypes.String, "e2-medium"),
				"name":         tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"list": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Tags []string `tfsdk:"tags"`
			}{
				Tags: []string{"red", "blue", "green"},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags": tftypes.List{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "blue"),
					tftypes.NewValue(tftypes.String, "green"),
				}),
			}),
		},
		"nested-list": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			val: struct {
				Disks []struct {
					ID                 string `tfsdk:"id"`
					DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
				} `tfsdk:"disks"`
			}{
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
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk1"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
		},
		"nested-single": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"boot_disk": {
							Attributes: SingleNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type: types.BoolType,
								},
							}),
						},
					},
				},
			},
			val: struct {
				BootDisk struct {
					ID                 string `tfsdk:"id"`
					DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
				} `tfsdk:"boot_disk"`
			}{
				BootDisk: struct {
					ID                 string `tfsdk:"id"`
					DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
				}{
					ID:                 "bootdisk",
					DeleteWithInstance: true,
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"boot_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				},
			}, map[string]tftypes.Value{
				"boot_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"id":                   tftypes.String,
						"delete_with_instance": tftypes.Bool,
					},
				}, map[string]tftypes.Value{
					"id":                   tftypes.NewValue(tftypes.String, "bootdisk"),
					"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
				}),
			}),
		},
		"object": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
								},
							},
							Optional: true,
						},
					},
				},
			},
			val: struct {
				ScratchDisk struct {
					Interface string `tfsdk:"interface"`
				} `tfsdk:"scratch_disk"`
			}{
				ScratchDisk: struct {
					Interface string `tfsdk:"interface"`
				}{
					Interface: "SCSI",
				},
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					},
				},
			}, map[string]tftypes.Value{
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "SCSI"),
				}),
			}),
		},
		"AttrTypeWithValidateError": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected:      tftypes.Value{},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			state: State{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.state.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.state.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestStateSetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		state         State
		path          *tftypes.AttributePath
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name":  tftypes.String,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name":  tftypes.NewValue(tftypes.String, "originalname"),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name":  tftypes.String,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name":  tftypes.NewValue(tftypes.String, "newname"),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"list": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags":  tftypes.List{ElementType: tftypes.String},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"tags": {
							Type: types.ListType{
								ElemType: types.StringType,
							},
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.List{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"list-element": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.List{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk0"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"id":                   tftypes.NewValue(tftypes.String, "disk1"),
							"delete_with_instance": tftypes.NewValue(tftypes.Bool, false),
						}),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"disks": {
							Attributes: ListNestedAttributes(map[string]Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}, ListNestedAttributesOptions{}),
							Optional: true,
							Computed: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(1),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "disk0"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":                   tftypes.String,
							"delete_with_instance": tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"object-attribute": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"scratch_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"interface": tftypes.NewValue(tftypes.String, "SCSI"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
								},
							},
							Optional: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"),
			val:  "NVME",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "NVME"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"AttrTypeWithValidateError": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "originalname"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			state: State{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name")),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.state.SetAttribute(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.state.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
