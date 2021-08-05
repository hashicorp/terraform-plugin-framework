package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/diagnostics"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPlanGet(t *testing.T) {
	testPlan := Plan{
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
					Type:     types.StringType,
					Required: true,
				},
			},
		},
	}

	type testPlanData struct {
		Name types.String `tfsdk:"name"`
	}

	var val testPlanData

	diags := testPlan.Get(context.Background(), &val)

	if diagnostics.DiagsHasErrors(diags) {
		t.Fatalf("unexpected error: %s", diagnostics.DiagsString(diags))
	}

	expected := testPlanData{
		Name: types.String{Value: "namevalue"},
	}

	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestPlanGet_AttrTypeWithValidate_Error(t *testing.T) {
	testPlan := Plan{
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
	}

	type testPlanData struct {
		Name types.String `tfsdk:"name"`
	}

	var val testPlanData

	diags := testPlan.Get(context.Background(), &val)

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestErrorDiagnostic), diagnostics.DiagString(diags[0]))
	}

	expected := testPlanData{
		Name: types.String{Value: ""},
	}

	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestPlanGet_AttrTypeWithValidate_Warning(t *testing.T) {
	testPlan := Plan{
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
	}

	type testPlanData struct {
		Name types.String `tfsdk:"name"`
	}

	var val testPlanData

	diags := testPlan.Get(context.Background(), &val)

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestWarningDiagnostic), diagnostics.DiagString(diags[0]))
	}

	expected := testPlanData{
		Name: types.String{Value: "namevalue"},
	}

	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestPlanGetAttribute_AttrTypeWithValidate_Error(t *testing.T) {
	testPlan := Plan{
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
	}

	nameVal, diags := testPlan.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestErrorDiagnostic), diagnostics.DiagString(diags[0]))
	}

	if nameVal != nil {
		t.Fatal("expected name to be nil")
	}
}

func TestPlanGetAttribute_AttrTypeWithValidate_Warning(t *testing.T) {
	testPlan := Plan{
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
	}

	nameVal, diags := testPlan.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestWarningDiagnostic), diagnostics.DiagString(diags[0]))
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

	if expected := "namevalue"; name.Value != expected {
		t.Errorf("Expected Name to be %q, got %q", expected, name.Value)
	}
}

func TestPlanSet(t *testing.T) {
	testPlan := Plan{
		Raw: tftypes.Value{},
		Schema: Schema{
			Attributes: map[string]Attribute{
				"name": {
					Type:     types.StringType,
					Required: true,
				},
			},
		},
	}

	type testPlanData struct {
		Name string `tfsdk:"name"`
	}

	diags := testPlan.Set(context.Background(), testPlanData{
		Name: "newvalue",
	})

	if diagnostics.DiagsHasErrors(diags) {
		t.Fatalf("error setting plan: %s", diagnostics.DiagsString(diags))
	}

	actual := testPlan.Raw
	expected := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "newvalue"),
	})

	if !expected.Equal(actual) {
		t.Fatalf("unexpected diff in testPlan.Raw (+wanted, -got): %s", cmp.Diff(actual, expected))
	}
}

func TestPlanSet_AttrTypeWithValidate_Error(t *testing.T) {
	testPlan := Plan{
		Raw: tftypes.Value{},
		Schema: Schema{
			Attributes: map[string]Attribute{
				"name": {
					Type:     testtypes.StringTypeWithValidateError{},
					Required: true,
				},
			},
		},
	}

	type testPlanData struct {
		Name string `tfsdk:"name"`
	}

	diags := testPlan.Set(context.Background(), testPlanData{
		Name: "newvalue",
	})

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestErrorDiagnostic), diagnostics.DiagString(diags[0]))
	}

	actual := testPlan.Raw
	expected := tftypes.Value{}

	if !expected.Equal(actual) {
		t.Fatalf("unexpected diff in testPlan.Raw (+wanted, -got): %s", cmp.Diff(actual, expected))
	}
}

func TestPlanSet_AttrTypeWithValidate_Warning(t *testing.T) {
	testPlan := Plan{
		Raw: tftypes.Value{},
		Schema: Schema{
			Attributes: map[string]Attribute{
				"name": {
					Type:     testtypes.StringTypeWithValidateWarning{},
					Required: true,
				},
			},
		},
	}

	type testPlanData struct {
		Name string `tfsdk:"name"`
	}

	diags := testPlan.Set(context.Background(), testPlanData{
		Name: "newvalue",
	})

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestWarningDiagnostic), diagnostics.DiagString(diags[0]))
	}

	actual := testPlan.Raw
	expected := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "newvalue"),
	})

	if !expected.Equal(actual) {
		t.Fatalf("unexpected diff in testPlan.Raw (+wanted, -got): %s", cmp.Diff(actual, expected))
	}
}

func TestPlanSetAttribute_AttrTypeWithValidate_Error(t *testing.T) {
	testPlan := Plan{
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
	}

	diags := testPlan.SetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), "newname")

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestErrorDiagnostic), diagnostics.DiagString(diags[0]))
	}

	expectedRawState := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "originalname"),
	})

	if diff := cmp.Diff(expectedRawState, testPlan.Raw, allowAllUnexported); diff != "" {
		t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestPlanSetAttribute_AttrTypeWithValidate_Warning(t *testing.T) {
	testPlan := Plan{
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
	}

	diags := testPlan.SetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), "newname")

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !cmp.Equal(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagnostics.DiagString(testtypes.TestWarningDiagnostic), diagnostics.DiagString(diags[0]))
	}

	expectedRawState := tftypes.NewValue(tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name": tftypes.String,
		},
	}, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, "newname"),
	})

	if diff := cmp.Diff(expectedRawState, testPlan.Raw, allowAllUnexported); diff != "" {
		t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
	}
}
