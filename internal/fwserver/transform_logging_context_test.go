// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MarkComputedNilsAsUnknown and NullifyWriteOnlyAttributes return
// tftypes.Transform callbacks that log every visited value with its attribute
// path. The logging context for a value has to be derived from the context the
// callback was created with. Deriving it from the context of the previously
// visited value nests one more context.WithValue per value, and every log call
// walks that whole chain to find the logger, which makes a single transform
// quadratic in the number of values:
// https://github.com/hashicorp/terraform-plugin-framework/issues/1086

func TestMarkComputedNilsAsUnknown_DoesNotNestLoggingContexts(t *testing.T) {
	t.Parallel()

	for name, testCase := range transformLoggingContextTestCases(50) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			ctx := tfsdklogtest.RootLogger(context.Background(), &output)
			ctx = logging.InitContext(ctx)

			recorder := &contextRecordingSchema{Schema: testCase.schema}

			_, err := tftypes.Transform(testCase.value, fwserver.MarkComputedNilsAsUnknown(ctx, testCase.value, recorder))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			recorder.checkContextsNotNested(t)
			recorder.checkLoggedAttributePaths(t, &output)
		})
	}
}

func TestNullifyWriteOnlyAttributes_DoesNotNestLoggingContexts(t *testing.T) {
	t.Parallel()

	for name, testCase := range transformLoggingContextTestCases(50) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer

			ctx := tfsdklogtest.RootLogger(context.Background(), &output)
			ctx = logging.InitContext(ctx)

			recorder := &contextRecordingSchema{Schema: testCase.schema}

			_, err := tftypes.Transform(testCase.value, fwserver.NullifyWriteOnlyAttributes(ctx, recorder))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			recorder.checkContextsNotNested(t)
			recorder.checkLoggedAttributePaths(t, &output)
		})
	}
}

func BenchmarkMarkComputedNilsAsUnknown100(b *testing.B) {
	benchmarkMarkComputedNilsAsUnknown(b, 100)
}

func BenchmarkMarkComputedNilsAsUnknown1000(b *testing.B) {
	benchmarkMarkComputedNilsAsUnknown(b, 1000)
}

func BenchmarkMarkComputedNilsAsUnknown10000(b *testing.B) {
	benchmarkMarkComputedNilsAsUnknown(b, 10000)
}

func benchmarkMarkComputedNilsAsUnknown(b *testing.B, elements int) {
	testCase := transformLoggingContextTestCases(elements)["list-nested-attribute"]
	ctx := benchmarkLoggingContext()

	for n := 0; n < b.N; n++ {
		_, err := tftypes.Transform(testCase.value, fwserver.MarkComputedNilsAsUnknown(ctx, testCase.value, testCase.schema))
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
	}
}

func BenchmarkNullifyWriteOnlyAttributes100(b *testing.B) {
	benchmarkNullifyWriteOnlyAttributes(b, 100)
}

func BenchmarkNullifyWriteOnlyAttributes1000(b *testing.B) {
	benchmarkNullifyWriteOnlyAttributes(b, 1000)
}

func BenchmarkNullifyWriteOnlyAttributes10000(b *testing.B) {
	benchmarkNullifyWriteOnlyAttributes(b, 10000)
}

func benchmarkNullifyWriteOnlyAttributes(b *testing.B, elements int) {
	testCase := transformLoggingContextTestCases(elements)["list-nested-attribute"]
	ctx := benchmarkLoggingContext()

	for n := 0; n < b.N; n++ {
		_, err := tftypes.Transform(testCase.value, fwserver.NullifyWriteOnlyAttributes(ctx, testCase.schema))
		if err != nil {
			b.Fatalf("unexpected error: %s", err)
		}
	}
}

// benchmarkLoggingContext returns a context with framework logging set up as
// during an RPC, with the log output discarded.
func benchmarkLoggingContext() context.Context {
	ctx := tfsdklogtest.RootLogger(context.Background(), io.Discard)

	return logging.InitContext(ctx)
}

type transformLoggingContextTestCase struct {
	schema schema.Schema
	value  tftypes.Value
}

// transformLoggingContextTestCases returns resources whose values hold the
// given number of elements, in the shapes the transforms handle differently:
// nested attributes with their own schema, a dynamic attribute and a
// collection attribute without schema for its elements.
func transformLoggingContextTestCases(elements int) map[string]transformLoggingContextTestCase {
	ruleType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"name":   tftypes.String,
			"id":     tftypes.String,
			"secret": tftypes.String,
		},
	}
	rules := make([]tftypes.Value, elements)
	manifestTypes := make(map[string]tftypes.Type, elements)
	manifest := make(map[string]tftypes.Value, elements)
	tags := make([]tftypes.Value, elements)

	for i := 0; i < elements; i++ {
		rules[i] = tftypes.NewValue(ruleType, map[string]tftypes.Value{
			"name":   tftypes.NewValue(tftypes.String, "rule-"+strconv.Itoa(i)),
			"id":     tftypes.NewValue(tftypes.String, nil),
			"secret": tftypes.NewValue(tftypes.String, "secret-"+strconv.Itoa(i)),
		})
		manifestTypes["key-"+strconv.Itoa(i)] = tftypes.String
		manifest["key-"+strconv.Itoa(i)] = tftypes.NewValue(tftypes.String, "value-"+strconv.Itoa(i))
		tags[i] = tftypes.NewValue(tftypes.String, "tag-"+strconv.Itoa(i))
	}

	return map[string]transformLoggingContextTestCase{
		"list-nested-attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"rules": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
								},
								"id": schema.StringAttribute{
									Computed: true,
								},
								"secret": schema.StringAttribute{
									Optional:  true,
									WriteOnly: true,
								},
							},
						},
						Required: true,
					},
				},
			},
			value: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"rules": tftypes.List{ElementType: ruleType},
				},
			}, map[string]tftypes.Value{
				"rules": tftypes.NewValue(tftypes.List{ElementType: ruleType}, rules),
			}),
		},
		"dynamic-attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"manifest": schema.DynamicAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			value: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"manifest": tftypes.DynamicPseudoType,
				},
			}, map[string]tftypes.Value{
				"manifest": tftypes.NewValue(tftypes.Object{AttributeTypes: manifestTypes}, manifest),
			}),
		},
		"list-attribute": {
			schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"tags": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			value: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"tags": tftypes.List{ElementType: tftypes.String},
				},
			}, map[string]tftypes.Value{
				"tags": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tags),
			}),
		},
	}
}

// contextRecordingSchema is a schema that records, for every attribute the
// transform looks up, the attribute path and how many values the context it
// was given carries.
type contextRecordingSchema struct {
	fwschema.Schema

	paths  []string
	depths []int
}

func (s *contextRecordingSchema) AttributeAtTerraformPath(ctx context.Context, p *tftypes.AttributePath) (fwschema.Attribute, error) {
	s.paths = append(s.paths, p.String())
	s.depths = append(s.depths, contextDepth(ctx))

	return s.Schema.AttributeAtTerraformPath(ctx, p)
}

// checkContextsNotNested fails when the contexts handed to the schema get
// deeper from one visited value to the next.
func (s *contextRecordingSchema) checkContextsNotNested(t *testing.T) {
	t.Helper()

	if len(s.depths) < 2 {
		t.Fatalf("expected the schema to be consulted for several values, got %d", len(s.depths))
	}

	if s.depths[0] == 0 {
		t.Fatal("unable to measure context nesting: the context does not print its context.WithValue chain")
	}

	for i, depth := range s.depths {
		if depth != s.depths[0] {
			t.Fatalf("the logging context nests deeper with every visited value: "+
				"value 1 of %d (%s) carries %d context values, value %d (%s) %d, value %d (%s) %d",
				len(s.depths), s.paths[0], s.depths[0],
				i+1, s.paths[i], depth,
				len(s.depths), s.paths[len(s.paths)-1], s.depths[len(s.depths)-1])
		}
	}
}

// checkLoggedAttributePaths fails when a log entry is not attributed to one of
// the visited values, or when no entry is logged at all.
func (s *contextRecordingSchema) checkLoggedAttributePaths(t *testing.T, output *bytes.Buffer) {
	t.Helper()

	entries, err := tfsdklogtest.MultilineJSONDecode(output)
	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected log entries for the visited values, got none")
	}

	visited := make(map[string]bool, len(s.paths))

	for _, path := range s.paths {
		visited[path] = true
	}

	for _, entry := range entries {
		path, ok := entry[logging.KeyAttributePath].(string)
		if !ok || !visited[path] {
			t.Fatalf("log entry not attributed to a visited value: %v", entry)
		}
	}
}

// contextDepth returns how many values were added to ctx with
// context.WithValue, counted in the chain the context prints as its string.
func contextDepth(ctx context.Context) int {
	return strings.Count(fmt.Sprint(ctx), ".WithValue(")
}
