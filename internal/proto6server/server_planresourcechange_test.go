package proto6server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		priorState       tftypes.Value
		proposedNewState tftypes.Value
		config           tftypes.Value
		priorPrivate     []byte
		providerMeta     tftypes.Value
		resource         string
		resourceType     tftypes.Type

		modifyPlanFunc func(context.Context, tfsdk.ModifyResourcePlanRequest, *tfsdk.ModifyResourcePlanResponse)

		// response expectations
		expectedPlannedState    tftypes.Value
		expectedRequiresReplace []*tftypes.AttributePath
		expectedPlannedPrivate  []byte
		expectedDiags           []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"one_changed": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
		"one_not_changed": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
		},
		"one_nil_state_and_config": {
			priorState:           tftypes.NewValue(testServeResourceTypeOneType, nil),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeOneType, nil),
			config:               tftypes.NewValue(testServeResourceTypeOneType, nil),
			resource:             "test_one",
			resourceType:         testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"two_nil_state_and_config": {
			priorState:           tftypes.NewValue(testServeResourceTypeTwoType, nil),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeTwoType, nil),
			config:               tftypes.NewValue(testServeResourceTypeTwoType, nil),
			resource:             "test_two",
			resourceType:         testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
		"two_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeTwoType, nil),
			config:               tftypes.NewValue(testServeResourceTypeTwoType, nil),
			resource:             "test_two",
			resourceType:         testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
		"three_nested_computed_no_changes": {
			resource:     "test_three",
			resourceType: testServeResourceTypeThreeType,
			priorState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
		},
		"three_nested_computed_configuration_change": {
			resource:     "test_three",
			resourceType: testServeResourceTypeThreeType,
			priorState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "newname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "newname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "newname"),
				"last_updated":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"first_updated": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
		},
		"three_nested_computed_nested_configuration_change": {
			resource:     "test_three",
			resourceType: testServeResourceTypeThreeType,
			priorState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, "yesterday"),
				"first_updated": tftypes.NewValue(tftypes.String, "last year"),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, "mycompstring"),
						"string":          tftypes.NewValue(tftypes.String, "mystring"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, nil),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, nil),
				"first_updated": tftypes.NewValue(tftypes.String, nil),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, nil),
						"string":          tftypes.NewValue(tftypes.String, nil),
					}),
				}),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeThreeType, map[string]tftypes.Value{
				"name":          tftypes.NewValue(tftypes.String, "myname"),
				"last_updated":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"first_updated": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"map_nested": tftypes.NewValue(tftypes.Map{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					},
				}, map[string]tftypes.Value{
					"one": tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"computed_string": tftypes.String,
							"string":          tftypes.String,
						},
					}, map[string]tftypes.Value{
						"computed_string": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
						"string":          tftypes.NewValue(tftypes.String, nil),
					}),
				}),
			}),
		},
		"one_add": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, nil),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
		"two_modifyplan_add_list_elem": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "auto-boot-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 1),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 456),
						"required_string": tftypes.NewValue(tftypes.String, "newvalue"),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
				resp.Plan.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "123456"),
					"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						}}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 10),
							"boot":    tftypes.NewValue(tftypes.Bool, false),
						}),
						tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						}}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "auto-boot-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 1),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
					"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_bool":   tftypes.Bool,
								"required_number": tftypes.Number,
								"required_string": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_bool":   tftypes.NewValue(tftypes.Bool, true),
							"required_number": tftypes.NewValue(tftypes.Number, 123),
							"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
						}),
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"required_bool":   tftypes.Bool,
								"required_number": tftypes.Number,
								"required_string": tftypes.String,
							},
						}, map[string]tftypes.Value{
							"required_bool":   tftypes.NewValue(tftypes.Bool, true),
							"required_number": tftypes.NewValue(tftypes.Number, 456),
							"required_string": tftypes.NewValue(tftypes.String, "newvalue"),
						}),
					}),
				})
			},
		},
		"two_modifyplan_requires_replace": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "1234567"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "1234567"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "1234567"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
				resp.RequiresReplace = []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")}
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")},
		},
		"two_modifyplan_diags_warning": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
				resp.RequiresReplace = []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")}
				resp.Diagnostics.AddWarning("I'm warning you", "You have been warned")
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")},
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "I'm warning you",
					Detail:   "You have been warned",
				},
			},
		},
		"two_modifyplan_diags_error": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			config: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
				"list_nested_blocks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"required_bool":   tftypes.Bool,
						"required_number": tftypes.Number,
						"required_string": tftypes.String,
					},
				}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"required_bool":   tftypes.Bool,
							"required_number": tftypes.Number,
							"required_string": tftypes.String,
						},
					}, map[string]tftypes.Value{
						"required_bool":   tftypes.NewValue(tftypes.Bool, true),
						"required_number": tftypes.NewValue(tftypes.Number, 123),
						"required_string": tftypes.NewValue(tftypes.String, "statevalue"),
					}),
				}),
			}),
			modifyPlanFunc: func(ctx context.Context, req tfsdk.ModifyResourcePlanRequest, resp *tfsdk.ModifyResourcePlanResponse) {
				resp.RequiresReplace = []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")}
				resp.Diagnostics.AddError("This is an error", "More details about the error")
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("id")},
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityError,
					Summary:  "This is an error",
					Detail:   "More details about the error",
				},
			},
		},
		"attr_plan_modifiers_nil_state_and_config": {
			priorState:           tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
			config:               tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
			resource:             "test_attribute_plan_modifiers",
			resourceType:         testServeResourceTypeAttributePlanModifiersType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, nil),
		},
		"attr_plan_modifiers_requiresreplace": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext3"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, map[string]tftypes.Value{
						"size":   tftypes.NewValue(tftypes.Number, 1),
						"format": tftypes.NewValue(tftypes.String, "ext4"),
					}),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedRequiresReplace: []*tftypes.AttributePath{

				tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("filesystem").WithAttributeName("format"),
				tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"),
			},
		},
		"attr_plan_modifiers_requiresreplaceif_true": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 999),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 999),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 999),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface"), tftypes.NewAttributePath().WithAttributeName("size")},
		},
		"attr_plan_modifiers_requiresreplaceif_false": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 1),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 1),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 1),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		"attr_plan_modifiers_diags": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "TESTDIAG"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		"attr_plan_modifiers_chained_modifiers": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		"attr_plan_modifiers_default_value_modifier": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, nil),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, nil),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, nil),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
		// TODO: Attribute plan modifiers should run before plan unknown marking.
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/183
		// "attr_plan_modifiers_trigger_computed_unknown": {
		// 	resource:     "test_attribute_plan_modifiers",
		// 	resourceType: testServeResourceTypeAttributePlanModifiersType,
		// 	priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
		// 		"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
		// 		"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
		// 		"name":                         tftypes.NewValue(tftypes.String, "TESTATTRONE"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	expectedDiags: []*tfprotov6.Diagnostic{
		// 		{
		// 			Severity: tfprotov6.DiagnosticSeverityWarning,
		// 			Summary:  "Warning diag",
		// 			Detail:   "This is a warning",
		// 		},
		// 	},
		// 	expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
		// 		"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		// 		"name":                         tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
		// 		"size":                         tftypes.NewValue(tftypes.Number, 3),
		// 		"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 			"id":        tftypes.String,
		// 			"interface": tftypes.String,
		// 			"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}},
		// 		}}, map[string]tftypes.Value{
		// 			"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
		// 			"interface": tftypes.NewValue(tftypes.String, "scsi"),
		// 			"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		// 				"size":   tftypes.Number,
		// 				"format": tftypes.String,
		// 			}}, nil),
		// 		}),
		// 		"region": tftypes.NewValue(tftypes.String, "DEFAULTVALUE"),
		// 	}),
		// 	expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		// },
		"attr_plan_modifiers_nested_modifier": {
			priorState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "my-scr-disk"),
					"interface": tftypes.NewValue(tftypes.String, "something-else"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, "statevalue"),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "TESTATTRTWO"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			config: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, nil),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "TESTATTRTWO"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeAttributePlanModifiersType, map[string]tftypes.Value{
				"computed_string_no_modifiers": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"name":                         tftypes.NewValue(tftypes.String, "name1"),
				"size":                         tftypes.NewValue(tftypes.Number, 3),
				"scratch_disk": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"id":        tftypes.String,
					"interface": tftypes.String,
					"filesystem": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}},
				}}, map[string]tftypes.Value{
					"id":        tftypes.NewValue(tftypes.String, "MODIFIED_TWO"),
					"interface": tftypes.NewValue(tftypes.String, "scsi"),
					"filesystem": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"size":   tftypes.Number,
						"format": tftypes.String,
					}}, nil),
				}),
				"region": tftypes.NewValue(tftypes.String, "region1"),
			}),
			resource:     "test_attribute_plan_modifiers",
			resourceType: testServeResourceTypeAttributePlanModifiersType,
			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Severity: tfprotov6.DiagnosticSeverityWarning,
					Summary:  "Warning diag",
					Detail:   "This is a warning",
				},
			},
			expectedRequiresReplace: []*tftypes.AttributePath{tftypes.NewAttributePath().WithAttributeName("scratch_disk").WithAttributeName("interface")},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := &testServeProvider{
				modifyPlanFunc: tc.modifyPlanFunc,
			}
			testServer := &Server{
				FrameworkServer: fwserver.Server{
					Provider: s,
				},
			}

			priorStateDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.priorState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			proposedStateDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.proposedNewState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			configDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.PlanResourceChangeRequest{
				TypeName:         tc.resource,
				PriorPrivate:     tc.priorPrivate,
				PriorState:       &priorStateDV,
				ProposedNewState: &proposedStateDV,
				Config:           &configDV,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.PlanResourceChange(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			gotPlannedState, err := got.PlannedState.Unmarshal(tc.resourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotPlannedState, tc.expectedPlannedState); diff != "" {
				t.Errorf("Unexpected diff in planned state (+wanted, -got): %s", diff)
				return
			}
			if string(got.PlannedPrivate) != string(tc.expectedPlannedPrivate) {
				t.Errorf("Expected planned private to be %q, got %q", tc.expectedPlannedPrivate, got.PlannedPrivate)
				return
			}
			if diff := cmp.Diff(got.RequiresReplace, tc.expectedRequiresReplace, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unexpected diff in requires replace (+wanted, -got): %s", diff)
				return
			}
		})
	}
}
