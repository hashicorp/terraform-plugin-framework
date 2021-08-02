package tfsdk

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConfigGet(t *testing.T) {
	testConfig := Config{
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

	type testConfigGetData struct {
		Name types.String `tfsdk:"name"`
	}

	var val testConfigGetData

	diags := testConfig.Get(context.Background(), &val)

	if diagsHasErrors(diags) {
		t.Fatalf("unexpected error: %s", diagsString(diags))
	}

	expected := testConfigGetData{
		Name: types.String{Value: "namevalue"},
	}

	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestConfigGet_AttrTypeWithValidate_Error(t *testing.T) {
	testConfig := Config{
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

	type testConfigGetData struct {
		Name types.String `tfsdk:"name"`
	}

	var val testConfigGetData

	diags := testConfig.Get(context.Background(), &val)

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !reflect.DeepEqual(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestErrorDiagnostic), diagString(diags[0]))
	}

	expected := testConfigGetData{
		Name: types.String{Value: ""},
	}

	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestConfigGet_AttrTypeWithValidate_Warning(t *testing.T) {
	testConfig := Config{
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

	type testConfigGetData struct {
		Name types.String `tfsdk:"name"`
	}

	var val testConfigGetData

	diags := testConfig.Get(context.Background(), &val)

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !reflect.DeepEqual(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestWarningDiagnostic), diagString(diags[0]))
	}

	expected := testConfigGetData{
		Name: types.String{Value: "namevalue"},
	}

	if diff := cmp.Diff(val, expected); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestConfigGetAttribute_AttrTypeWithValidate_Error(t *testing.T) {
	testConfig := Config{
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

	nameVal, diags := testConfig.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !reflect.DeepEqual(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestErrorDiagnostic), diagString(diags[0]))
	}

	if nameVal != nil {
		t.Fatal("expected name to be nil")
	}
}

func TestConfigGetAttribute_AttrTypeWithValidate_Warning(t *testing.T) {
	testConfig := Config{
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

	nameVal, diags := testConfig.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}

	if !reflect.DeepEqual(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestWarningDiagnostic), diagString(diags[0]))
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
