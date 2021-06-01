package reflect

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBuildReflectValue_unhandledNull(t *testing.T) {
	t.Parallel()

	var s string
	_, err := buildReflectValue(context.Background(), tftypes.NewValue(tftypes.String, nil), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: unhandled null value`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestBuildReflectValue_unhandledUnknown(t *testing.T) {
	t.Parallel()

	var s string
	_, err := buildReflectValue(context.Background(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := `: unhandled unknown value`; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}
