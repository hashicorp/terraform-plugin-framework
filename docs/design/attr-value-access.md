# Accessing `attr.Value` Values

There are numerous places in the framework that require access to config,
state, and plan values. In most places, like CRUD methods and the provider
configuration callback, we provide access to the entire config, state, and
plan. `Get` and `GetAttribute` methods on the config, state, and plan are then
utilized to retrieve either the entire config/state/plan or to retrieve a
specific attribute's value.

There are some places in the framework, however, that are dealing with a
specific attribute within a plan, config, or state. Validation and plan
customization, for example, may have helpers in the schema that are meant to
operate on a single attribute's values. While we could provide the entire
config, plan, or state to these helpers, they would all need to pull out the
attribute's value using a path, which is verbose and seems unnecessary. It
would be better to provide just the attribute's value to these types.

This design document talks about ways to do that.

## Options

### Request helpers

In the plan modifier or validator, we could add a helper on their request type
to retrieve that attribute's value. For example, a `req.ConfigAttribute(ctx,
&target)` method could allow for a provider developer to get easy access to
that specific attribute's value. To accomplish this, we'd need to set the
attribute's path on the request before calling the plan modifier or validator,
and then have the helper use that to call `GetAttribute`.

This seems like it would work, but it still requires that first step from
users, making it less discoverable. It fits nicely in with the rest of the
framework, however, and seems to not have unit testing implications as long as
the path is exported. It would make unit testing slightly more challenging, as
the entire config/plan/state would need to be specified, instead of just the
attribute's value.

### Type assertions

We could just pass the `attr.Value` to the plan modifier or validator and then
type assert on it:

```go
func planModifier(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	configVal := req.Config.(types.String)
}
```

This has problems, though: it limits the modifier or validator to only work
with the subset of resources that use that exact `attr.Value`, meaning
reusability suffers. It's also uncomfortably similar to the
`schema.ResourceData.Get` pattern that developers disliked so much. It's,
overall, just a very brittle interface.

It does, however, offer a slightly more unit testable option than the request
helpers, as only the value needs to be in the request, not an entire resource.

### Reflection helpers

We can also add helpers, either to `tfsdk` or to `attr`, to use the same
reflection used in `Get` on the `attr.Value`:

```go
func planModifier(ctx context.Context, req ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	var configVal string
	err := attr.ValueAs(ctx, req.Config, &configVal)
}
```

This matches how config, state, and plan values are accessed everywhere else in
the framework, retains the flexibility and compatibility properties they
afford, and allows the modifier and validator to be reused in much more varied
cases. It's unit testable, by requiring only the specific attribute's value to
be set.

## Recommendation

We're recommending that in situations where a single attribute's value needs to
be accessed, that value be represented by an `attr.Value` and reflection
helpers be created to allow access to it.
