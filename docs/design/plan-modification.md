# Modifying Plans

The Terraform [resource instance change lifecycle](https://github.com/hashicorp/terraform/blob/3b6c1ef156387e6652ed8516bce2af2422b7fcf1/docs/resource-instance-change-lifecycle.md) involves the following sequence of RPCs:
 1. `ValidateResourceTypeConfig`
 1. `PlanResourceChange`
 1. `ApplyResourceChange`

The plugin framework handles requests and responses for these RPCs, allowing providers to hook in to `ValidateResourceTypeConfig` via validation helpers (https://github.com/hashicorp/terraform-plugin-framework/issues/17), and to `ApplyResourceChange` via resource CRUD functions. This design document concerns the ways we allow provider developers to hook into the `PlanResourceChange` RPC, giving providers control over the plan rendered to users via the Terraform CLI.

## The `PlanResourceChange` RPC

```proto
message PlanResourceChange {
    message Request {
        string type_name = 1;
        DynamicValue prior_state = 2;
        DynamicValue proposed_new_state = 3;
        DynamicValue config = 4;
        bytes prior_private = 5; 
        DynamicValue provider_meta = 6;
    }

    message Response {
        DynamicValue planned_state = 1;
        repeated AttributePath requires_replace = 2;
        bytes planned_private = 3; 
        repeated Diagnostic diagnostics = 4;
    }
}
```

The `PlanResourceChange` RPC request contains `config`, `prior_state` and `proposed_new_state` values, from which the provider is required to determine and return the `planned_state` in the response.

Terraform CLI renders a diff between `prior_state` and `planned_state` to the user for confirmation.

The `PlanResourceChange` RPC response also contains a list of attribute paths: `repeated AttributePath requires_replace`. For each attribute path in this list, Terraform checks whether its value has changed, and if it has, then the user is shown a plan saying that the resource instance must be replaced, with the attribute in question marked with `# forces replacement`.

In allowing providers to control the `PlanResourceChange` response, i.e. "modify the plan", the plugin framework therefore enables providers not only to modify the diff that will be displayed to the user (and ultimately applied), but also to branch into a destroy-and-create lifecycle phase, triggering other RPCs.

Plan modification has two distinct use cases for providers:
  - Modifying plan values, and
  - Forcing resource replacement.

This design document therefore distinguishes "ModifyPlan" from "RequiresReplace" behaviour, the former being a superset of the latter.

## History: `ForceNew` and `CustomizeDiff`

In `helper/schema`, there are two ways, both somewhat indirect, that a provider can customise the plan.

### `ForceNew`

`ForceNew` is a boolean schema behaviour that marks a resource field as requiring the resource to be destroyed and recreated if the value of the field is changed. Schema example:
```go
"base_image": {
  Type:     schema.TypeString,
  Required: true,
  ForceNew: true,
},
```
After receiving the `PlanResourceChange` request, the SDK determines whether any change was made to a field marked as `ForceNew`, and if so, adds it to the returned `RequiresNew` array.

The SDK also executes logic at resource validation time (`InternalValidate`) to enforce the condition that if a resource does not have an Update function, all non-Computed attributes must have ForceNew set; and that if all fields are ForceNew or Computed without Optional, Update must _not_ be defined.

### `CustomizeDiff`

Rather than exposing Terraform plans to provider developers, `helper/schema` has as a first-class concept the _resource diff_. Providers can optionally define a `CustomizeDiff` method on the `Resource` struct, which resembles a CRUD method, except that instead of `ResourceData` the function is supplied a `*ResourceDiff`, and is called during several points in the resource lifecycle, and must therefore be "resilient to support all scenarios".

A large proportion of the examples of `CustomizeDiff` in large cloud provider code involves conditionally setting `ForceNew` behaviour on an attribute, most often:
 - If certain conditions hold on the value of the attribute (e.g. if the bandwidth of an instance is reduced)
 - If certain conditions hold on the values of the attribute and other attributes.

`CustomizeDiff` is also used to implement multi-attribute validation, e.g. checking that at most one of two particular attributes is set. The present framework, on the other hand, separates plan modification from validation, and the solutions proposed below are concerned only with plan modification.

### `helper/customdiff`

The legacy SDK provides a set of reusable and composable helper functions in its `customdiff` package, intended to reduce the need for long and complex logic to reside in a single `CustomizeDiff` function in provider code. This small package is best summarised in its [documentation](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff).

## Solution options

### `tfsdk.Resource.ModifyPlan()`

An extension to the `tfsdk.Resource` interface could add an optional `ModifyPlan()` function to resource implementations:

```go
// ResourceWithModifyPlan represents a resource instance with a ModifyPlan
// function.
type ResourceWithModifyPlan interface {
 	Resource
 
 	ModifyPlan(context.Context, ModifyResourcePlanRequest, *ModifyResourcePlanResponse)
}
```

Following the request and response parameter pattern used for resource CRUD functions, with the familiar types for `Config`, `State`, and `Plan`, minimises cognitive overhead for provider developers:

```go
type ModifyResourcePlanRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config Config

	// State is the current state of the resource.
	State State

	// Plan is the planned new state for the resource.
	Plan Plan

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta Config
}

type ModifyResourcePlanResponse struct {
	// Plan is the planned new state for the resource.
	Plan Plan

	// RequiresReplace is a list of tftypes.AttributePaths that require the
	// resource to be replaced. They should point to the specific field
	// that changed that requires the resource to be destroyed and
	// recreated.
	RequiresReplace []*tftypes.AttributePath

	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested resource. Returning an empty slice
	// indicates a successful validation with no warnings or errors
	// generated.
	Diagnostics []*tfprotov6.Diagnostic
}

func (r ModifyResourcePlanResponse) AppendRequiresReplace(attrPath *tftypes.AttributePath) {
  r.RequiresReplace = append(r.RequiresReplace, attrPath)
}
```

The only field unique to the `ModifyPlan` request or response types is `RequiresReplace` (whose name is copied directly from the protocol, but which could very well be called `ForceNewAttributes` or similar). 

In provider code:

```go
type myFileResource struct{}

func (r myFileResource) ModifyPlan(ctx context.Context, req ModifyResourcePlanRequest, resp *ModifyResourcePlanResponse) {
	var state fileData
	err := req.State.Get(ctx, &state)
	if err != nil {
		// diags
	}
	var plan fileData
	err = req.Plan.Get(ctx, &plan)
	if err != nil {
		// diags
	}

	// force resource recreation if the new favourite number is larger than the old
	if plan.FavoriteNumber > state.FavoriteNumber {
		resp.AppendRequiresReplace(tftypes.NewAttributePath.WithAttributeName("favorite_number"))
	}
}
```

The framework would call this `ModifyPlan` method during the `PlanResourceChange` RPC, populating the appropriate values of Config, Plan, and State in the `ModifyResourcePlanRequest`.

#### Tradeoffs

A `ModifyPlan` method on a Resource is as unit testable as any CRUD method, and has similar compatibility and discoverability properties. The code feels perfectly Go-native.

The main tradeoff here is verbosity. The actual work done by the function is the selection of attribute path(s) whose old and new values should be compared, the comparison condition, and the selection of attribute path(s) to mark as RequiresReplace. In the `FavoriteNumber` example above in particular, a less verbose option is illustrated below with the use of `schema.Attribute.ModifyPlanFunc`. For complex cases of plan modification involving multiple attributes, reading config, or making API calls, the `Resource.ModifyPlan` method has an appropriate amount of verbosity. We anticipate that most use cases for plan modification will not be this complex.

### `schema.Attribute.RequiresReplace`

Like `helper/schema`, we could add a `ForceNew bool`, here called `RequiresReplace` to match the protocol, to the framework's `schema.Attribute` struct, enabling provider developers to take advantage of this simple schema behaviour with one line of code.

Precedent for expressing RequiresReplace in a declarative manner is found in the AWS API, which has a concept of _create-only properties_: "properties that are only able to be specified by the customer when creating a resource" (see [CloudFormation Resource Semantics](https://github.com/aws-cloudformation/cloudformation-resource-schema#resource-semantics). For this API at least, marking such properties with ForceNew aligns with the [provider design principle](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html#resource-and-attribute-schema-should-closely-match-the-underlying-api): _Resource and attribute schema should closely match the underlying API_.

#### Tradeoffs

A single `RequiresReplace` schema attribute is easily discoverable, Go-native, minimally verbose, and a reasonably transparent representation of the protocol's concept of `RequiresReplace`.

A cursory survey of existing provider code finds `ForceNew` very widely used, and seldom the subject of individual acceptance tests, likely because provider developers do not feel the need to test a behaviour whose implementation resides in the SDK itself. For this reason the question of unit testability is also moot. 

One possible disadvantage of this approach is that it is atomic with respect to compatibility - unlike provider-defined plan modification functions like those described below, the framework must deprecate the field or undergo a breaking change in order to modify the behaviour of `schema.Attribute.RequiresReplace`.

### `schema.Attribute.RequiresReplaceIf`

An field on the `schema.Attribute` struct could add an optional `RequiresReplaceIf` function to schema attributes:

```go
type Attribute struct {
  // ...
  
  RequiresReplaceIf RequiresReplaceIfFunc 
}

type RequiresReplaceIfFunc func(context.Context, old, new attr.Value) bool
```

Provider code:

```go
type fileResourceType struct{}

// GetSchema returns the schema for this resource.
func (f fileResourceType) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"favorite_number": {
				Type:     types.NumberType,
				Required: true,
				RequiresReplaceIf: func(ctx context.Context, old, new attr.Value) bool {
				  oldVal := old.(types.Number)
				  newVal := new.(types.Number)
				  
				  if !oldVal.Unknown && !oldVal.Null && !newVal.Unknown && !newVal.Null {
				    if newVal.Value.Cmp(oldVal.Value) > 0 {
				      return true
				    }
				  }
				  return false
				},
			},
		},
	}, nil
}
```

While this provider implementation of the RequiresReplace function is unfortunately more verbose than its `helper/schema` equivalent, since any value comparison must be preceded by a check for null and unknown values, we can easily imagine a `RequiresReplaceIfSet` helper implemented in the framework, illustrated in the _Composition and other helpers_ section below. The ability to work with null and unknown values is a new feature not available in `helper/schema`.

Since `RequiresReplaceIf` must be implemented on the resource type, not the resource instance, it has no access to the provider's API client. It is intended to be used for the simple old and new value comparisons that make up the majority of observed usages of `CustomizeDiff` in current provider code.

#### Tradeoffs

With easily mocked arguments, `RequiresReplaceIf` functions are unit testable, and, being a field on the `schema.Attribute` struct, easily discoverable. The provider code is reasonably Go-native, as far as the complexities of the various type systems involved allow.

As indicated above, helper functions can be introduced to reduce verbosity.

Compatibility may be an issue if a common use case emerges for RequiresReplace conditions a little more complex than the signature of `RequiresReplaceIfFunc` allows - for example, requiring reading config values. This could be anticipated by having the signature instead be:
```go
type RequiresReplaceIfFunc (context.Context, ModifyPlanRequest) bool
```
By doing this, however, we would lose the benefits of the `old, new` parameters in reducing verbosity - the provider code would have to repeat the attribute path in order to retrieve the values from `req.Plan` and `req.State`.

### `attr.TypeWithRequiresReplace`

A `ModifyPlan()` or `RequiresReplaceIf()` function could be added to an extension of the `attr.Type` interface:

```go
type TypeWithRequiresReplace interface {
  Type
  
  RequiresReplaceIf(context.Context, old, new Value) bool
```

This would allow bundling reusable RequiresReplace behaviour up with a custom type's validation and other behaviours.

Without knowing how custom types will be used by provider developers, this option seems premature, and makes less sense than bundling validation functions with custom types.

### Composition and other helpers

#### `All()`

Similarly to `helper/schema`, the `All` composition helper runs all `RequiresReplaceFunc`s and returns true only if all funcs return true.

```go
func All(funcs ...RequiresReplaceFunc) bool {}
```

#### `Sequence()`

Similarly to `helper/schema`, the `Sequence` composition helper runs all `RequiresReplaceFunc`s in sequence, stopping at the first that returns false. 

```go
func Sequence(funcs ...RequiresReplaceFunc) bool {}
```

#### `RequiresReplaceIfSet()`

This helper reduces the verbosity of `RequiresReplaceFunc`s that always return false if either old or new value is null or unknown.

Framework implementation:

```go
func RequiresReplaceIfSet(f RequiresReplaceFunc) RequiresReplaceFunc {
  return func (ctx context.Context, old, new attr.Value) bool {
    if old.IsNull() || old.IsUnknown() || new.IsNull() || new.IsUnknown() {
      return false
    }
    return f(ctx, old, new)
  }
}
```

Note that this would require the addition of `IsNull()` and `IsUnknown()` functions to the `attr.Value` interface, since there is at present no way to determine whether a generic `attr.Value` is null or unknown.

## Recommendations

We recommend implementing `schema.Attribute.RequiresReplace`, `schema.Attribute.RequiresReplaceIf`, and the `ResourceWithModifyPlan` interface. Composition and other helpers can be implemented as required.
