package tfsdk

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfReflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var attributeValueReflectType = reflect.TypeOf(new(attr.Value)).Elem()

type State struct {
	Raw    tftypes.Value
	Schema schema.Schema
}

func isValidFieldName(name string) bool {
	re := regexp.MustCompile("^[a-z][a-z0-9_]*$")
	return re.MatchString(name)
}

// Get populates the struct passed as `target` with the entire state. No type assertion necessary.
func (s State) Get(ctx context.Context, target interface{}) error {
	return tfReflect.Into(ctx, s.Raw, target, tfReflect.Options{}, tftypes.NewAttributePath())
}

// GetAttribute retrieves the attribute found at `path` and returns it as an attr.Value,
// which provider developers need to assert the type of
func (s State) GetAttribute(ctx context.Context, path tftypes.AttributePath) (attr.Value, error) {

}

// MustGetAttribute retrieves the attribute as GetAttribute does, but populates target using As,
// using the simplified representation without Unknown. Errors if Unknown present
func (s State) MustGetAttribute(ctx context.Context, path tftypes.AttributePath, target interface{}) error {
	return nil
}
