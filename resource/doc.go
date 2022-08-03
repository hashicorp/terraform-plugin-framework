// Package resource contains all interfaces, request types, and response types
// for a managed resource implementation.
//
// In Terraform, a managed resource is a concept which enables provider
// developers to offer practitioners full lifecycle management (create, read,
// update, and delete) of a infrastructure component. Managed resources can
// also stand in for one-time infrastructure operations that require tracking,
// by implementing create logic, while omitting update and delete logic.
//
// Resources are saved into the Terraform state and can be referenced by other
// parts of a configuration. Resources are defined by a resource type/name,
// such as "example_thing", a schema representing the structure and data types
// of configuration, plan, and state, and lifecycle logic.
//
// The main starting point for implementations in this package is the
// Resource type which represents an instance of a resource type that has
// its own configuration, plan, state, and lifecycle logic. A Resource is
// instantiated from a provider.ResourceType type NewResource method, which
// also defines the resource schema. The provider.ResourceType types are
// referenced by a provider.Provider type GetResources method, which enables
// the resource for practitioner and testing usage.
package resource
