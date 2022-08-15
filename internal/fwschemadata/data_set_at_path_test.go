package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataSetAtPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		data          fwschemadata.Data
		path          path.Path
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"add-List-Element-append": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
		"add-List-Element-append-length-error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(2),
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
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("disks"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 3 as list currently has 1 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"add-List-Element-first": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(0),
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
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-List-Element-first-length-error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
				}, nil),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("disks"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 2 as list currently has 0 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"add-Map-Element-append": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key1": tftypes.NewValue(tftypes.String, "key1value"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
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
			path: path.Root("test").AtMapKey("key2"),
			val:  "key2value",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key1": tftypes.NewValue(tftypes.String, "key1value"),
					"key2": tftypes.NewValue(tftypes.String, "key2value"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-Map-Element-first": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
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
			path: path.Root("test").AtMapKey("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"add-Set-Element-append": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
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
					"disks": tftypes.NewValue(tftypes.Set{
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
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
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
				"disks": tftypes.NewValue(tftypes.Set{
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
		"add-Set-Element-first": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
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
					"disks": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"id":                   tftypes.String,
								"delete_with_instance": tftypes.Bool,
							},
						},
					}, nil),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
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
				"disks": tftypes.NewValue(tftypes.Set{
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
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Bool": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Bool,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.Bool, true),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.BoolType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: path.Root("test"),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Bool,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Bool, false),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-List": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
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
			path: path.Root("tags"),
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
		"overwrite-List-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
		"overwrite-Map": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"originalkey": tftypes.NewValue(tftypes.String, "originalvalue"),
						"otherkey":    tftypes.NewValue(tftypes.String, "othervalue"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
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
			path: path.Root("test"),
			val: map[string]string{
				"newkey": "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"newkey": tftypes.NewValue(tftypes.String, "newvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Map-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test": tftypes.NewValue(tftypes.Map{
						ElementType: tftypes.String,
					}, map[string]tftypes.Value{
						"key":   tftypes.NewValue(tftypes.String, "originalvalue"),
						"other": tftypes.NewValue(tftypes.String, "should be untouched"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
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
			path: path.Root("test").AtMapKey("key"),
			val:  "newvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key":   tftypes.NewValue(tftypes.String, "newvalue"),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Number": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Number,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.Number, 1),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.NumberType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: path.Root("test"),
			val:  2,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Number,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Number, 2),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Object": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
								"other":     tftypes.String,
							},
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"scratch_disk": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
							"other":     tftypes.String,
						},
					}, map[string]tftypes.Value{
						"interface": tftypes.NewValue(tftypes.String, "SCSI"),
						"other":     tftypes.NewValue(tftypes.String, "originalvalue"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"scratch_disk": {
							Type: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"interface": types.StringType,
									"other":     types.StringType,
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
			path: path.Root("scratch_disk"),
			val: struct {
				Interface string `tfsdk:"interface"`
				Other     string `tfsdk:"other"`
			}{
				Interface: "NVME",
				Other:     "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"scratch_disk": tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"interface": tftypes.String,
							"other":     tftypes.String,
						},
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"scratch_disk": tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"interface": tftypes.String,
						"other":     tftypes.String,
					},
				}, map[string]tftypes.Value{
					"interface": tftypes.NewValue(tftypes.String, "NVME"),
					"other":     tftypes.NewValue(tftypes.String, "newvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Object-Attribute": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
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
			path: path.Root("scratch_disk").AtName("interface"),
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
		"overwrite-Set": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags":  tftypes.Set{ElementType: tftypes.String},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"tags": {
							Type: types.SetType{
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
			path: path.Root("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.Set{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-Set-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
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
					"disks": tftypes.NewValue(tftypes.Set{
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
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "disk1"},
					"delete_with_instance": types.Bool{Value: false},
				},
			}),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
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
				"disks": tftypes.NewValue(tftypes.Set{
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
		"overwrite-Set-Element-duplicate": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags": tftypes.Set{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"tags": tftypes.NewValue(tftypes.Set{
						ElementType: tftypes.String,
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "one"),
						tftypes.NewValue(tftypes.String, "two"),
						tftypes.NewValue(tftypes.String, "three"),
					}),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"tags": {
							Type: types.SetType{
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
			path: path.Root("tags").AtSetValue(types.String{Value: "three"}),
			val:  "three",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags": tftypes.Set{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
					tftypes.NewValue(tftypes.String, "three"),
				}),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"overwrite-String": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"test":  tftypes.NewValue(tftypes.String, "originalvalue"),
					"other": tftypes.NewValue(tftypes.String, "should be untouched"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
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
			path: path.Root("test"),
			val:  "newvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.String,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.String, "newvalue"),
				"other": tftypes.NewValue(tftypes.String, "should be untouched"),
			}),
		},
		"write-root": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Bool,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.BoolType,
							Required: true,
						},
					},
				},
			},
			path: path.Empty(),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Bool,
				},
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\nexpected tftypes.Object[\"test\":tftypes.Bool], got tftypes.Bool",
				),
			},
		},
		"write-Bool": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Bool,
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.BoolType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: path.Root("test"),
			val:  false,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Bool,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Bool, false),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-List": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
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
			path: path.Root("tags"),
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
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-List-AttrTypeWithValidateWarning-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.List{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: testtypes.ListTypeWithValidateWarning{
								ListType: types.ListType{
									ElemType: types.StringType,
								},
							},
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
			path: path.Root("test").AtListIndex(0),
			val:  "testvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.List{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("test")),
			},
		},
		"write-List-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(0),
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
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-List-Element-length-error": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
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
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(1),
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
			}, nil),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("disks"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to create a value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"Cannot add list element 2 as list currently has 0 length. To prevent ambiguity, only the next element can be added to a list. Add empty elements into the list prior to this call, if appropriate.",
				),
			},
		},
		"write-List-Element-AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     testtypes.StringTypeWithValidateWarning{},
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtListIndex(0),
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
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("disks").AtListIndex(0).AtName("id")),
			},
		},
		"write-Map": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
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
			path: path.Root("test"),
			val: map[string]string{
				"newkey": "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"newkey": tftypes.NewValue(tftypes.String, "newvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Map-AttrTypeWithValidateWarning-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: testtypes.MapTypeWithValidateWarning{
								MapType: types.MapType{
									ElemType: types.StringType,
								},
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
			path: path.Root("test").AtMapKey("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("test")),
			},
		},
		"write-Map-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
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
			path: path.Root("test").AtMapKey("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Map-Element-AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Map{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: types.MapType{
								ElemType: testtypes.StringTypeWithValidateWarning{},
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
			path: path.Root("test").AtMapKey("key"),
			val:  "keyvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Map{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.String,
				}, map[string]tftypes.Value{
					"key": tftypes.NewValue(tftypes.String, "keyvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("test").AtMapKey("key")),
			},
		},
		"write-Number": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.Number,
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type:     types.NumberType,
							Required: true,
						},
						"other": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: path.Root("test"),
			val:  1,
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.Number,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.Number, 1),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Object": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"scratch_disk": tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"interface": tftypes.String,
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
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
			path: path.Root("scratch_disk"),
			val: struct {
				Interface string `tfsdk:"interface"`
			}{
				Interface: "NVME",
			},
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
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Set": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"tags":  tftypes.Set{ElementType: tftypes.String},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"tags": {
							Type: types.SetType{
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
			path: path.Root("tags"),
			val:  []string{"one", "two"},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags":  tftypes.Set{ElementType: tftypes.String},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "one"),
					tftypes.NewValue(tftypes.String, "two"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Set-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     types.StringType,
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
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
				"disks": tftypes.NewValue(tftypes.Set{
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
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"write-Set-AttrTypeWithValidateWarning-Element": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test": tftypes.Set{
							ElementType: tftypes.String,
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
							Type: testtypes.SetTypeWithValidateWarning{
								SetType: types.SetType{
									ElemType: types.StringType,
								},
							},
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
			path: path.Root("test").AtSetValue(types.String{Value: "testvalue"}),
			val:  "testvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test": tftypes.Set{
						ElementType: tftypes.String,
					},
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test": tftypes.NewValue(tftypes.Set{
					ElementType: tftypes.String,
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "testvalue"),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("test")),
			},
		},
		"write-Set-Element-AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"disks": tftypes.Set{
							ElementType: tftypes.Object{
								AttributeTypes: map[string]tftypes.Type{
									"id":                   tftypes.String,
									"delete_with_instance": tftypes.Bool,
								},
							},
						},
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"disks": {
							Attributes: tfsdk.SetNestedAttributes(map[string]tfsdk.Attribute{
								"id": {
									Type:     testtypes.StringTypeWithValidateWarning{},
									Required: true,
								},
								"delete_with_instance": {
									Type:     types.BoolType,
									Optional: true,
								},
							}),
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
			path: path.Root("disks").AtSetValue(types.Object{
				AttrTypes: map[string]attr.Type{
					"id":                   types.StringType,
					"delete_with_instance": types.BoolType,
				},
				Attrs: map[string]attr.Value{
					"id":                   types.String{Value: "mynewdisk"},
					"delete_with_instance": types.Bool{Value: true},
				},
			}),
			val: struct {
				ID                 string `tfsdk:"id"`
				DeleteWithInstance bool   `tfsdk:"delete_with_instance"`
			}{
				ID:                 "mynewdisk",
				DeleteWithInstance: true,
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"disks": tftypes.Set{
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
				"disks": tftypes.NewValue(tftypes.Set{
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
						"id":                   tftypes.NewValue(tftypes.String, "mynewdisk"),
						"delete_with_instance": tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("disks").AtSetValue(types.Object{
					AttrTypes: map[string]attr.Type{
						"id":                   types.StringType,
						"delete_with_instance": types.BoolType,
					},
					Attrs: map[string]attr.Value{
						"id":                   types.String{Value: "mynewdisk"},
						"delete_with_instance": types.Bool{Value: true},
					},
				}).AtName("id")),
			},
		},
		"write-String": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"test":  tftypes.String,
						"other": tftypes.String,
					},
				}, nil),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"test": {
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
			path: path.Root("test"),
			val:  "newvalue",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"test":  tftypes.String,
					"other": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"test":  tftypes.NewValue(tftypes.String, "newvalue"),
				"other": tftypes.NewValue(tftypes.String, nil),
			}),
		},
		"AttrTypeWithValidateError": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			path: path.Root("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "originalname"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(path.Root("name"))},
		},
		"AttrTypeWithValidateWarning": {
			data: fwschemadata.Data{
				TerraformValue: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: tfsdk.Schema{
					Attributes: map[string]tfsdk.Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			path: path.Root("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Root("name")),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.data.SetAtPath(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.data.TerraformValue, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
