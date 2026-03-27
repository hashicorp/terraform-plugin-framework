// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildValidatorGroupsIncludesNestedBlockMembers(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"settings": tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"alpha": tftypes.String,
			"beta":  tftypes.String,
		}},
	}}

	testSchema := testschema.Schema{Blocks: map[string]fwschema.Block{
		"settings": testschema.BlockWithObjectValidators{
			Attributes: map[string]fwschema.Attribute{
				"alpha": testschema.Attribute{Optional: true, Type: types.StringType},
				"beta":  testschema.Attribute{Optional: true, Type: types.StringType},
			},
			Validators: []validator.Object{
				testConflictsWithObjectValidator{Object: testvalidator.Object{}, paths: path.Expressions{
					path.MatchRelative().AtName("alpha"),
					path.MatchRelative().AtName("beta"),
				}},
			},
		},
	}}

	config := tftypes.NewValue(testType, map[string]tftypes.Value{
		"settings": tftypes.NewValue(testType.AttributeTypes["settings"], map[string]tftypes.Value{
			"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
			"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
		}),
	})

	groups := buildValidatorGroups(context.Background(), config, testSchema, nil, getConflictsWithPaths)
	if len(groups) != 1 {
		t.Fatalf("expected one group, got %d", len(groups))
	}

	for _, members := range groups {
		expected := path.Paths{path.Root("settings").AtName("alpha"), path.Root("settings").AtName("beta")}
		if diff := cmp.Diff(expected, members); diff != "" {
			t.Fatalf("unexpected members diff: %s", diff)
		}
	}
}

func TestBuildValidatorGroupsIgnoresNonConfigValidators(t *testing.T) {
	t.Parallel()

	testType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"alpha": tftypes.String,
		"beta":  tftypes.String,
	}}

	testSchema := testschema.Schema{Attributes: map[string]fwschema.Attribute{
		"alpha": testschema.Attribute{Optional: true, Type: types.StringType},
		"beta":  testschema.Attribute{Optional: true, Type: types.StringType},
	}}

	res := &testprovider.ResourceWithConfigValidators{
		Resource: &testprovider.Resource{},
		ConfigValidatorsMethod: func(context.Context) []resource.ConfigValidator {
			return []resource.ConfigValidator{
				&testResourceExactlyOneOfValidator{paths: path.Expressions{
					path.MatchRoot("alpha"),
					path.MatchRoot("beta"),
				}},
			}
		},
	}

	config := tftypes.NewValue(testType, map[string]tftypes.Value{
		"alpha": tftypes.NewValue(tftypes.String, "configured-alpha"),
		"beta":  tftypes.NewValue(tftypes.String, "configured-beta"),
	})

	groups := buildValidatorGroups(context.Background(), config, testSchema, res, getConflictsWithPaths)
	if len(groups) != 0 {
		t.Fatalf("expected no groups, got %d", len(groups))
	}
}
