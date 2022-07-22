package reflect_test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ diag.Diagnostic = expectedDiagnostic{}

type expectedDiagnostic struct {
	detail   string
	severity diag.Severity
	summary  string
	t        *testing.T
}

func (d expectedDiagnostic) Detail() string {
	return d.detail
}

func (d expectedDiagnostic) Equal(other diag.Diagnostic) bool {
	detailEqual := other.Detail() == d.Detail()
	severityEqual := other.Severity() == d.Severity()
	summaryEqual := other.Summary() == d.Summary()

	if !detailEqual {
		d.t.Logf("Details equality: %s != %s", other.Detail(), d.Detail())
	}
	if !severityEqual {
		d.t.Logf("Severity equality: %s != %s", other.Severity(), d.Severity())
	}
	if !summaryEqual {
		d.t.Logf("Summary equality: %s != %s", other.Summary(), d.Summary())
	}
	return detailEqual && severityEqual && summaryEqual
}

func (d expectedDiagnostic) Severity() diag.Severity {
	return d.severity
}

func (d expectedDiagnostic) Summary() string {
	return d.summary
}

func TestIntoPointerError(t *testing.T) {
	ctx := context.TODO()
	var target float64
	val, err := types.String{
		Value: "foo",
	}.ToTerraformValue(ctx)
	if err != nil {
		t.Fatal("Construct test value failed", err)
	}
	diags := reflect.Into(ctx, types.StringType, val, target, reflect.Options{})

	if !diags.HasError() {
		t.Fatal("Expected error")
	}

	if len(diags) != 1 {
		t.Fatal("Expected only one error")
	}

	t.Logf("Searching substring in `%s`", diags[0].Detail())
	if !strings.Contains(diags[0].Detail(), "target must be a pointer, got float64, which is a float64. Maybe change to *string") {
		t.Fatal("Details not found.")
	}
}
