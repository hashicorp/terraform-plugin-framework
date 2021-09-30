package reflect

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func toTerraform5ValueErrorDiag(err error, path *tftypes.AttributePath) *diag.GenericDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to convert into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

func toTerraformValueErrorDiag(err error, path *tftypes.AttributePath) *diag.GenericDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to convert the Attribute value into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

func validateValueErrorDiag(err error, path *tftypes.AttributePath) *diag.GenericDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to validate the Terraform value type. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

func valueFromTerraformErrorDiag(err error, path *tftypes.AttributePath) *diag.GenericDiagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		"Value Conversion Error",
		"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
	)
}

var _ diag.Diagnostic = &DiagIntoIncompatibleType{}

type DiagIntoIncompatibleType struct {
	*diag.GenericDiagnostic

	Val        tftypes.Value
	TargetType reflect.Type
	Err        error
}

func NewDiagIntoIncompatibleType(path *tftypes.AttributePath, val tftypes.Value, targetType reflect.Type, err error) *DiagIntoIncompatibleType {
	diagnostic := &DiagIntoIncompatibleType{
		GenericDiagnostic: diag.NewAttributeErrorDiagnostic(
			path,
			"Value Conversion Error",
			fmt.Sprintf("An unexpected error was encountered trying to convert %T into %s. This is always an error in the provider. Please report the following to the provider developer:\n\n%s", val, targetType, err.Error()),
		),

		Val:        val,
		TargetType: targetType,
		Err:        err,
	}

	return diagnostic
}

func (d *DiagIntoIncompatibleType) Equal(o diag.Diagnostic) bool {
	if d == nil && o == nil {
		return true
	}
	od, ok := o.(*DiagIntoIncompatibleType)
	if !ok {
		return false
	}
	if !d.Val.Equal(od.Val) {
		return false
	}
	if d.TargetType != od.TargetType {
		return false
	}
	if d.Err.Error() != od.Err.Error() {
		return false
	}
	return d.GenericDiagnostic.Equal(od.GenericDiagnostic)
}

type DiagNewAttributeValueIntoWrongType struct {
	*diag.GenericDiagnostic

	ValType    reflect.Type
	TargetType reflect.Type
	SchemaType attr.Type
}

func NewDiagNewAttributeValueIntoWrongType(path *tftypes.AttributePath, valType reflect.Type, targetType reflect.Type, schemaType attr.Type) *DiagNewAttributeValueIntoWrongType {
	diagnostic := &DiagNewAttributeValueIntoWrongType{
		GenericDiagnostic: diag.NewAttributeErrorDiagnostic(
			path,
			"Value Conversion Error",
			fmt.Sprintf("An unexpected error was encountered trying to convert into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\nCannot use attr.Value %s, only %s is supported because %T is the type in the schema", targetType, valType, schemaType),
		),

		ValType:    valType,
		TargetType: targetType,
		SchemaType: schemaType,
	}

	return diagnostic
}

func (d *DiagNewAttributeValueIntoWrongType) Equal(o diag.Diagnostic) bool {
	if d == nil && o == nil {
		return true
	}
	od, ok := o.(*DiagNewAttributeValueIntoWrongType)
	if !ok {
		return false
	}
	if d.ValType != od.ValType {
		return false
	}
	if d.TargetType != od.TargetType {
		return false
	}
	if !d.SchemaType.Equal(od.SchemaType) {
		return false
	}
	return d.GenericDiagnostic.Equal(od.GenericDiagnostic)
}
