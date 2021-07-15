# Modifying Plans

The Terraform [resource instance change lifecycle](https://github.com/hashicorp/terraform/blob/3b6c1ef156387e6652ed8516bce2af2422b7fcf1/docs/resource-instance-change-lifecycle.md) involves the following sequence of RPCs:
 1. `ValidateResourceTypeConfig`
 1. `PlanResourceChange`
 1. `ApplyResourceChange`

The plugin framework handles requests and responses for these RPCs, allowing providers to hook in to `ValidateResourceTypeConfig` via validation helpers (https://github.com/hashicorp/terraform-plugin-framework/issues/17), and to `ApplyResourceChange` via resource CRUD functions. This design document concerns the ways we allow provider developers to hook into the `PlanResourceChange` RPC, giving providers control over the plan rendered to users via the Terraform CLI, and which Terraform commits to apply.

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

Plan modification currently has two distinct use cases for providers:
  - Modifying plan values, and
  - Forcing resource replacement.

This design document therefore distinguishes "ModifyPlan" from "RequiresReplace" behaviour, the former being a superset of the latter.

## History: `ForceNew`, `DiffSuppressFunc`, and `CustomizeDiff`

In `helper/schema`, there are three ways, all somewhat indirect, that a provider can customise the plan.

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

### `DiffSuppressFunc`

```go
type SchemaDiffSuppressFunc func(k, old, new string, d *ResourceData) bool
```

Providers can use another schema behaviour, `DiffSuppressFunc`, to control whether a detected diff on a schema field should be considered semantically different. If this function returns true, any diff in the element values is ignored. This is commonly used to ignore differences in string capitalisation, or logically equivalent JSON values.

### `CustomizeDiff`

```go
type CustomizeDiffFunc func(context.Context, *ResourceDiff, interface{}) error
```

Rather than exposing Terraform plans to provider developers, `helper/schema` has as a first-class concept the _resource diff_. Providers can optionally define a `CustomizeDiff` method on the `Resource` struct, which resembles a CRUD method, except that instead of `ResourceData` the function is supplied a `*ResourceDiff`, and is called during several points in the resource lifecycle, and must therefore be "resilient to support all scenarios".

Unlike `DiffSuppressFunc`, `CustomizeDiff` is supplied the `meta` parameter, so API calls can be made.

A large proportion of the examples of `CustomizeDiff` in large cloud provider code involves conditionally setting `ForceNew` behaviour on an attribute, most often:
 - If certain conditions hold on the value of the attribute (e.g. if the bandwidth of an instance is reduced)
 - If certain conditions hold on the values of the attribute and other attributes.

`CustomizeDiff` is also used to implement multi-attribute validation, e.g. checking that at most one of two particular attributes is set. The present framework, on the other hand, separates plan modification from validation, and the solutions proposed below are concerned only with plan modification.

### `helper/customdiff`

The legacy SDK provides a set of reusable and composable helper functions in its `customdiff` package, intended to reduce the need for long and complex logic to reside in a single `CustomizeDiff` function in provider code. This small package is best summarised in its [documentation](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff).

## Solution options

### 1. `tfsdk.Resource.ModifyPlan()`

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
		resp.RequiresReplace = append(resp.RequiresReplace, tftypes.NewAttributePath.WithAttributeName("favorite_number"))
	}
}
```

The framework would call this `ModifyPlan` method during the `PlanResourceChange` RPC, populating the appropriate values of Config, Plan, and State in the `ModifyResourcePlanRequest`.

#### Tradeoffs

A `ModifyPlan` method on a Resource is as unit testable as any CRUD method, and has similar compatibility and discoverability properties. The code feels perfectly Go-native.

The main tradeoff here is verbosity. The actual work done by the function is the selection of attribute path(s) whose old and new values should be compared, the comparison condition, and the selection of attribute path(s) to mark as RequiresReplace. In the `FavoriteNumber` example above in particular, a less verbose option is illustrated below with the use of `schema.Attribute.ModifyPlanFunc`. For complex cases of plan modification involving multiple attributes, reading config, or making API calls, the `Resource.ModifyPlan` method has an appropriate amount of verbosity. We anticipate that most use cases for plan modification will not be this complex.

The inability to use a sequence of helper functions with more declarative syntax (see options 3 and 4) also makes this option more verbose.

### 2. `schema.Attribute.RequiresReplace`

Like `helper/schema`, we could add a `ForceNew bool`, here called `RequiresReplace` to match the protocol, to the framework's `schema.Attribute` struct, enabling provider developers to take advantage of this simple schema behaviour with one line of code.

Precedent for expressing `RequiresReplace` in a declarative manner is found in the AWS API, which has a concept of _create-only properties_: "properties that are only able to be specified by the customer when creating a resource" (see [CloudFormation Resource Semantics](https://github.com/aws-cloudformation/cloudformation-resource-schema#resource-semantics). For this API at least, marking such properties with ForceNew aligns with the [provider design principle](https://www.terraform.io/docs/extend/hashicorp-provider-design-principles.html#resource-and-attribute-schema-should-closely-match-the-underlying-api): _Resource and attribute schema should closely match the underlying API_.

Of course, the simplicity of a declarative `RequiresReplace` property means that there are many plan modification use cases it cannot cover, such as conditionally marking a field as requiring the resource to be replaced. This solution must therefore be combined with other solutions from this document.

#### Tradeoffs

A single `RequiresReplace` schema attribute is easily discoverable, Go-native, minimally verbose, and a reasonably transparent representation of the protocol's concept of `RequiresReplace`.

A cursory survey of existing provider code finds `ForceNew` very widely used, and seldom the subject of individual acceptance tests, likely because provider developers do not feel the need to test a behaviour whose implementation resides in the SDK itself. For this reason the question of unit testability is also moot. 

One possible disadvantage of this approach is that it is atomic with respect to compatibility - unlike provider-defined plan modification functions like those described below, the framework must deprecate the field or undergo a breaking change in order to modify the behaviour of `schema.Attribute.RequiresReplace`.

### 3. `schema.Attribute.RequiresReplaceIf`

An field on the `schema.Attribute` struct could add an optional `RequiresReplaceIf` function to schema attributes:

```go
type Attribute struct {
  // ...
  
  RequiresReplaceIf RequiresReplaceIfFunc 
}

type RequiresReplaceIfFunc func(context.Context, state, config attr.Value) bool
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
				RequiresReplaceIf: func(ctx context.Context, state, config attr.Value) bool {
				  stateVal := state.(types.Number)
				  configVal := config.(types.Number)
				  
				  if !stateVal.Unknown && !stateVal.Null && !configVal.Unknown && !configVal.Null {
				    if configVal.Value.Cmp(stateVal.Value) > 0 {
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
By doing this, however, we would lose the benefits of the `state, config` parameters in reducing verbosity - the provider code would have to repeat the attribute path in order to retrieve the values from `req.Plan` and `req.State`.


### 3a. Composition and other helpers

#### `All()`

Similarly to `helper/schema`, the `All` composition helper runs all `RequiresReplaceFunc`s and returns true only if all funcs return true.

```go
func All(funcs ...RequiresReplaceFunc) bool {}
```

#### `Sequence()`

Similarly to `helper/schema`, the `Sequence` composition helper runs all `RequiresReplaceFunc`s in sequence, stopping at the first that returns true. 

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

### 4. `schema.Attribute.PlanModifiers`

Extending the abstraction of `RequiresReplaceIf` one level higher, we can add a `PlanModifiers` field on the `schema.Attribute` struct, with the following framework code:

```go
type AttributePlanModifier interface {
  Description(context.Context) string
  MarkdownDescription(context.Context) string

  Modify(context.Context, ModifyAttributePlanRequest, *ModifyAttributePlanResponse)
}

type AttributePlanModifiers []AttributePlanModifier

type Attribute struct {
  // ...
  PlanModifiers AttributePlanModifiers
}

type ModifyAttributePlanRequest struct {
	// Config is the configuration the user supplied for the attribute.
	Config attr.Value

	// State is the current state of the attribute.
	State attr.Value

	// Plan is the planned new state for the attribute.
	Plan attr.Value

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta Config
}

type ModifyAttributePlanResponse struct {
	// Plan is the planned new state for the attribute.
	Plan attr.Value

	// RequiresReplace indicates whether a change in the attribute
	// requires replacement of the whole resource.
	RequiresReplace bool

	// Diagnostics report errors or warnings related to determining the
	// planned state of the requested resource. Returning an empty slice
	// indicates a successful validation with no warnings or errors
	// generated.
	Diagnostics []*tfprotov6.Diagnostic
}
```

This approach directly offers documentation hooks, so that plan modification behaviours can be documented alongside their definition and included in generated schema docs, and therefore also in tools such as the Terraform Language Server that consume the schema.

In this case, `RequiresReplace` and `RequiresReplaceIf` can be implemented as `AttributePlanModifier`s, e.g.:


```go
func RequiresReplace() AttributePlanModifier {
  return RequiresReplaceModifier{}
}

type RequiresReplaceModifier struct{}

func (r RequiresReplace) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
  resp.RequiresReplace = true
}

func (r RequiresReplace) Description(ctx context.Context) string {
  // ...
}

func (r RequiresReplace) MarkdownDescription(ctx context.Context) string {
  // ...
}
```

```go
func RequiresReplaceIf(f RequiresReplaceIfFunc, description markdownDescription string) AttributePlanModifier {
  return RequiresReplaceIfModifier{
    f: f, 
    description: description, 
    markdownDescription: markdownDescription
  }
}

type RequiresReplaceIfFunc func(context.Context, state, config attr.Value) (bool, error)

type RequiresReplaceIfModifier struct {
  f RequiresReplaceIfFunc
  description string
  markdownDescription string
}

func (r RequiresReplaceIfModifier) Modify(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
  resp.RequiresReplace = r.f(ctx, req.State, req.Config)
}

func (r RequiresReplaceIfModifier) Description(ctx context.Context) string {
  return r.description
}

func (r RequiresReplaceIfModifier) MarkdownDescription(ctx context.Context) string {
  return r.markdownDescription
}
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
				PlanModifiers: schema.AttributePlanModifiers{
				  RequiresReplaceIf(func(ctx context.Context, state, config attr.Value) (bool, error) {
				    stateVal := state.(types.Number)
				    configVal := config.(types.Number)
				  
				    if !stateVal.Unknown && !stateVal.Null && !configVal.Unknown && !configVal.Null {
				      if configVal.Value.Cmp(stateVal.Value) > 0 {
				        return true, nil
				      }
				    }
				    return false, nil
				  }),
				  CustomModifier,
				  // ...
				},
			},
		},
	}, nil
}
```

Here, `CustomModifier` is a user-defined `AttributePlanModifier`.

The `AttributePlanModifier`s in the slice of `PlanModifiers` are executed in order. Note that unlike the `customdiff.All` and `customdiff.Sequence` composition helpers in SDKv2, there is no choice to be made here between executing all helpers, and stopping at the first that "returns true", since the function could be setting the `resp.RequiresReplace` bool _or_ modifying the plan.

The fields in the `ModifyAttributePlanRequest` and `ModifyAttributePlanResponse` struct are not the same as those in `ModifyResourcePlanRequest` and `ModifyResourcePlanResponse` from option 1. In particular, it is not possible to change the planned value of _other_ attributes inside an attribute's `PlanModifier`. If it were, it would be possible for two or more attributes to have `PlanModifier`s modifying each other's planned values, with no clear indication of the order in which those operations would be performed. 

Framework documentation should make it clear that `schema.Attribute.PlanModifier` is scoped to a single attribute with no access to other attribute values or API requests. If either of these is needed, provider developers should use `tfsdk.Resource.ModifyPlan()`.

#### Tradeoffs

Similarly to option 3, the plan modifier functions are easily unit testable, and the `PlanModifier` field easily discoverable on the `schema.Attribute` struct. The code is reasonably Go-native.

This solution aims to improve on the compatibility of option 3 by ensuring that no further fields need be added to `schema.Attribute` in future for the purposes of plan modification.

If we were to change the fields of `ModifyResourcePlanRequest` to allow access to the full plan, state, and config (so that the result of the modify plan function could depend on the planned value of another attribute, for example), it would come at the cost of verbosity in the user-defined `RequiresIfFunc`, since the user must now retrieve the attribute values from the full `Config` and `State` rather than having them supplied as `attr.Value`s in the function arguments.

### 4a. `attr.ValueAs()`

This option is to be considered an extension to option 4.

Plan modifier functions (`AttributePlanModifier.Modify()`), and helper functions such as `RequiresReplaceIf`, are supplied `attr.Value` arguments from which the provider code must derive attribute values. In the above example, this is done with type assertions:

```go
RequiresReplaceIf(func(ctx context.Context, state, config attr.Value) (bool, error) {
  stateVal := state.(types.Number)
  configVal := config.(types.Number)
  // ...
}
```

Avoiding such type assertions is one of the design goals of the new framework.

Solution 4 should therefore be considered in combination with a way of avoiding such type assertions in provider code. One such proposal, which will be included in a separate design doc, is an `attr.ValueAs()` function:

```go
func ValueAs(ctx context.Context, val attr.Value, target interface{}) error {}
```

Similar to functions such as `state.Get(ctx, val, target)`, this function could use reflection to populate `target` with the value in `val`, returning an error if `target` is not of a compatible type. The provider code could then be:


```go
RequiresReplaceIf(func(ctx context.Context, state, config attr.Value) (bool, error) {
  var stateInt int
  var configInt int
  err: = attr.ValueAs(ctx, state, stateInt)
  if err != nil {
  	return false, err
  }
    err: = attr.ValueAs(ctx, config, configInt)
  if err != nil {
  	return false, err
  }
  // ...
}
```

### 5. `attr.TypeWithModifyPlan`

A `ModifyPlan()` or `RequiresReplaceIf()` function could be added to an extension of the `attr.Type` interface:

```go
type TypeWithModifyPlan interface {
  Type
  
  ModifyPlan(context.Context, ModifyAttributeTypePlanRequest, *ModifyAttributeTypePlanResponse) bool
```

This would allow bundling reusable `ModifyPlan` behaviour up with a custom type's validation and other behaviours. This could be useful, for example, in a custom timestamp type to squash semantically meaningless diffs, so provider developers do not have to specify the attribute plan modifier wherever the attribute appears in a schema.

Without knowing how custom types will be used by provider developers, this option seems premature, and makes less sense than bundling validation functions with custom types.


## Recommendations

We recommend implementing `schema.Attribute.PlanModifiers`, and the `ResourceWithModifyPlan` interface. Composition, `attr.TypeWithModifyPlan`, and other helpers can be implemented as required.
