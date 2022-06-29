package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// attributePath returns the path.Path equivalent of a *tftypes.AttributePath.
//
// TODO: This function should be exported as internal/fromtftypes.AttributePath
// except that doing so would currently introduce an import cycle due to the
// tfsdk.Schema parameter here and Config/Plan/State.PathMatches needing to
// call this function until the schema data is migrated to attr.Value.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/172
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func attributePath(ctx context.Context, tfType *tftypes.AttributePath, schema Schema) (path.Path, diag.Diagnostics) {
	fwPath := path.Empty()

	for tfTypeStepIndex, tfTypeStep := range tfType.Steps() {
		currentTfTypeSteps := tfType.Steps()[:tfTypeStepIndex+1]
		currentTfTypePath := tftypes.NewAttributePathWithSteps(currentTfTypeSteps)
		attrType, err := schema.AttributeTypeAtPath(currentTfTypePath)

		if err != nil {
			return path.Empty(), diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						// Since this is an error with the attribute path
						// conversion, we cannot return a protocol path-based
						// diagnostic. Returning a framework human-readable
						// representation seems like the next best thing to do.
						fmt.Sprintf("Attribute Path: %s\n", currentTfTypePath.String())+
						fmt.Sprintf("Original Error: %s", err),
				),
			}
		}

		fwStep, err := fromtftypes.AttributePathStep(ctx, tfTypeStep, attrType)

		if err != nil {
			return path.Empty(), diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is either an error in terraform-plugin-framework or a custom attribute type used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						// Since this is an error with the attribute path
						// conversion, we cannot return a protocol path-based
						// diagnostic. Returning a framework human-readable
						// representation seems like the next best thing to do.
						fmt.Sprintf("Attribute Path: %s\n", currentTfTypePath.String())+
						fmt.Sprintf("Original Error: %s", err),
				),
			}
		}

		// In lieu of creating a path.NewPathFromSteps function, this path
		// building logic is inlined to not expand the path package API.
		switch fwStep := fwStep.(type) {
		case path.PathStepAttributeName:
			fwPath = fwPath.AtName(string(fwStep))
		case path.PathStepElementKeyInt:
			fwPath = fwPath.AtListIndex(int(fwStep))
		case path.PathStepElementKeyString:
			fwPath = fwPath.AtMapKey(string(fwStep))
		case path.PathStepElementKeyValue:
			fwPath = fwPath.AtSetValue(fwStep.Value)
		default:
			return fwPath, diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Convert Attribute Path",
					"An unexpected error occurred while trying to convert an attribute path. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						// Since this is an error with the attribute path
						// conversion, we cannot return a protocol path-based
						// diagnostic. Returning a framework human-readable
						// representation seems like the next best thing to do.
						fmt.Sprintf("Attribute Path: %s\n", currentTfTypePath.String())+
						fmt.Sprintf("Original Error: unknown path.PathStep type: %#v", fwStep),
				),
			}
		}
	}

	return fwPath, nil
}
