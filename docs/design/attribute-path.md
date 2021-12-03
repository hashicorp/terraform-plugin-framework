# Attribute Paths

There are several situations in which a provider developer wants to identify a specific subset of the state, config, or plan. They may wish to identify the source of a problem in a diagnostic, or reference another part of the config during validation. We call this functionality "attribute paths", and it is meant to enable arbitrarily-precise targeting of a Terraform value.

## Prior Art

### protocol

The protocol defines paths using protobufs:

```protobuf
message AttributePath {
    message Step {
        oneof selector {
            // Set "attribute_name" to represent looking up an attribute
            // in the current object value.
            string attribute_name = 1;
            // Set "element_key_*" to represent looking up an element in
            // an indexable collection type.
            string element_key_string = 2;
            int64 element_key_int = 3;
        }
    }
    repeated Step steps = 1;
}
```

### tftypes

There exists a [`tftypes.AttributePath`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-go/tftypes#AttributePath) implementation in terraform-plugin-go. This is a somewhat-verbose implementation that mirrors the protocol rather directly, with some conveniences added.

Paths are made up of "steps", each adding more specificity regarding the value that is being targeted. A step can be of four types:

1. An AttributeName, the name of an attribute within an object
2. An ElementKeyString, a string key identifying an element in a map
3. An ElementKeyInt, an integer key identifying an element in a list or tuple
4. An ElementKeyValue, a `tftypes.Value` key identifying an element in a set

It is worth noting that `ElementKeyValue` is an identity key--the key is also the value being identified--and cannot be sent over the protocol, as the protocol has no way to encode that logic.

Paths are created by calling `tftypes.NewAttributePath()`, then using `With*` methods to add steps until the path is complete:

```go
tftypes.NewAttributePath().WithAttributeName("foo").WithElementKeyString("bar").WithElementKeyInt(1)
```

This is a flexible interface that matches the protocol rather directly, and therefore will be able to evolve with the protocol in the future.

### cty

There exists a [`cty.Path`](https://pkg.go.dev/github.com/zclconf/go-cty/cty#Path) implementation, which is what terraform-plugin-sdk and terraform itself use. This implementation is a little more general purpose than specific to Terraform.

The `Path` is a slice of `PathStep`s, with the following `PathStep` implementations available:

1. GetAttr, retrieving an attribute from an object
2. IndexString, retrieving an element from a map by its key
3. IndexInt, retrieving en element from a list or tuple by its position
4. Index, retrieving an element from a set by its identity

## Use Cases

There are two distinct use cases for wanting to specify part of a value: specifying a concrete value, and specifying a pattern of values to match.

### Concrete Values

Concrete values are the simplest case; the paths point to a specific, concrete values. Specific elements in lists, specific attributes in objects. These are seen most often in diagnostics. This is the only type of value that the `tftypes`, `cty`, and `protocol` implementations can indicate.

Something that no implementation does yet, however, is encode _relative_ paths to concrete values. For example, there's no way to say "attribute 'foo' of the object I am an attribute of". This would be useful for validation logic and helpers.

### Patterns of Values

There exist situations in which specifying a pattern of values is useful. For example, when validating a configuration, it's useful to say that if an attribute is set on one element in a list, it cannot be set on any _other_ elements in the list. Or to say that every object element in a map must have a specific attribute set.

These patterns are naturally a superset of the concrete values, as the pattern can be restrictive enough to identify only a single attribute.

As validation helpers naturally deal with specifying pieces of the configuration, it is probably best that we keep these use cases in mind as we develop abstractions, so we can enable them.

## Options

### A Single Abstraction

We could attempt to build a single abstraction for both these use cases, allowing them to be mixed and matched anywhere. This has the benefit of lower cognitive load on provider developers, but has the cost of not making it clear when concrete values are required or not. For example, diagnostics would require concrete values, as would our `GetAttribute` helpers.

### Separate Abstractions

We could implement an abstraction for each of these use cases separately. This has the benefit of making it clear when concrete values are needed. But it does limit us in the future, if we ever want to be able to accept concrete values or a pattern of values; we'd need to come up with an interface, or provide an implementation that worked with each, or otherwise find a way around the type system in that case.

## Proposal

We're proposing to build an abstraction for each of these use cases, as the pattern use case is mostly needed for validation helpers, which can opt into the most expansive possibilities there.
