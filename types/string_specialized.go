package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// specializedStringType is a convenience helper for the common situation where
// an underlying API takes a value as a string but imposes additional
// validation and/or normalization rules on that string.
type specializedStringType struct {
	opts SpecializedStringOpts
}

type SpecializedStringOpts struct {
	// NormalizeFunc is an optional function to convert a given string to the
	// for that the underlying system would store and return it in.
	//
	// If you provide a NormalizeFunc then Value.Equal will return true for
	// any pair of values that normalize to the same string.
	NormalizeFunc func(given string) string
	ValidateFunc  func(ctx context.Context, val string, path *tftypes.AttributePath) diag.Diagnostics
	TypeString    string
}

func SpecializedStringType(opts SpecializedStringOpts) attr.Type {
	return &specializedStringType{opts}
}

func (t *specializedStringType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.String
}

func (t *specializedStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return SpecializedString{ty: t, Unknown: true}, nil
	}
	if in.IsNull() {
		return SpecializedString{ty: t, Null: true}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	return SpecializedString{ty: t, Value: s}, nil
}

func (t *specializedStringType) Equal(other attr.Type) bool {
	if other, ok := other.(*specializedStringType); ok {
		// Two specialized string types are equal only if they came from
		// the same call to SpecializedStringType.
		return t == other
	}
	return false
}

func (t *specializedStringType) Validate(ctx context.Context, val tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	var diags diag.Diagnostics
	strDiags := StringType.Validate(ctx, val, path)
	diags.Append(strDiags...)
	if strDiags.HasError() {
		return diags
	}
	if t.opts.ValidateFunc == nil {
		return diags
	}
	// We validated the value for StringType above, so we know it must be
	// a string.
	if val.IsNull() || !val.IsKnown() {
		return diags // no special validation for null or unknown strings
	}
	var s string
	err := val.As(&s)
	if err != nil {
		// We should not get here, because we already validated that we had
		// a known, non-null string above. This is just for robustness, then.
		diags.AddError("Invalid string", err.Error())
	}
	customDiags := t.opts.ValidateFunc(ctx, s, path)
	diags.Append(customDiags...)
	return diags
}

func (t *specializedStringType) String() string {
	if t.opts.TypeString != "" {
		return t.opts.TypeString
	}
	return "types.SpecializedStringType"
}

func (t *specializedStringType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return StringType.ApplyTerraform5AttributePathStep(step)
}

type SpecializedString struct {
	ty *specializedStringType

	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value string
}

var _ attr.Value = SpecializedString{}

// Equal returns true if the given value is a SpecializedString of the same
// type as the reciever and if the type's "NormalizeFunc" produces the same
// value for both the receiver and for the other given value.
//
// If the value's type does not have a NormalizeFunc then this is equivalent
// to String.Equal.
func (s SpecializedString) Equal(other attr.Value) bool {
	o, ok := other.(SpecializedString)
	if !ok {
		return false
	}
	if o.ty != s.ty {
		return false // different specializations of String
	}
	if s.Unknown != o.Unknown {
		return false
	}
	if s.Null != o.Null {
		return false
	}
	if norm := s.ty.opts.NormalizeFunc; !(s.Null || s.Unknown) && norm != nil {
		ss := norm(s.Value)
		os := norm(o.Value)
		return ss == os
	}
	return s.Value == o.Value
}

// Type returns the specialized string type for this value.
func (s SpecializedString) Type(_ context.Context) attr.Type {
	return s.ty
}

// ToTerraformValue returns the data contained in the SpecializedString as
// a string.
func (s SpecializedString) ToTerraformValue(_ context.Context) (interface{}, error) {
	if s.Null {
		return nil, nil
	}
	if s.Unknown {
		return tftypes.UnknownValue, nil
	}
	return s.Value, nil
}
