package reflect_test

import (
	"context"
	"reflect"
	"testing"

	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBuildValue_unhandledNull(t *testing.T) {
	t.Parallel()

	var s string
	_, diags := refl.BuildValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if !diagsHasErrors(diags) {
		t.Error("Expected error, didn't get one")
	}
	if expected := `unhandled null value`; !diagsContainsDetail(diags, expected) {
		t.Errorf("Expected error to be %q, got %s", expected, diagsString(diags))
	}
}

func TestBuildValue_unhandledUnknown(t *testing.T) {
	t.Parallel()

	var s string
	_, diags := refl.BuildValue(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if !diagsHasErrors(diags) {
		t.Error("Expected error, didn't get one")
	}
	if expected := `unhandled unknown value`; !diagsContainsDetail(diags, expected) {
		t.Errorf("Expected error to be %q, got %s", expected, diagsString(diags))
	}
}
