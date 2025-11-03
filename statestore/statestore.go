// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import "context"

type StateStore interface {
	// Metadata should return the full name of the state store, such
	// as examplecloud_store.
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// Schema should return the schema for this state store.
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	// Configure enables provider-level data or clients to be set in the
	// provider-defined StateStore type. Configure can also be used to
	// perform "online" validation, such as verifying permissions for storing
	// state.
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)

	// Read returns the entire state from the state store.
	Read(context.Context, ReadRequest, *ReadResponse)

	// Write writes the entire state to the state store. When a new state is
	// created, Terraform will call Write with an empty state blob.
	Write(context.Context, WriteRequest, *WriteResponse)

	// Lock attempts to obtain a lock for the given state ID in a state
	// store.
	Lock(context.Context, LockRequest, *LockResponse)

	// Unlock releases a lock for the given state ID in a state store.
	Unlock(context.Context, UnlockRequest, *UnlockResponse)

	// GetStates returns all state IDs for a state store.
	GetStates(context.Context, GetStatesRequest, *GetStatesResponse)

	// DeleteState deletes a given state ID from a state store.
	DeleteState(context.Context, DeleteStatesRequest, *DeleteStatesResponse)
}

// StateStoreWithConfigure is an interface type that extends StateStore to
// include a method which the framework will automatically call so provider
// developers have the opportunity to setup any necessary provider-level data
// or clients in the StateStore type.
type StateStoreWithConfigure interface {
	StateStore

	// Configure enables provider-level data or clients to be set in the
	// provider-defined StateStore type.
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}

// StateStoreWithConfigValidators is an interface type that extends StateStore to include declarative validations.
//
// Declaring validation using this methodology simplifies the implementation of
// reusable functionality. These also include descriptions, which can be used
// for automating documentation.
//
// Validation will include ConfigValidators and ValidateConfig if both are
// implemented, in addition to any Attribute or Type validation.
type StateStoreWithConfigValidators interface {
	StateStore

	// ConfigValidators returns a list of functions which will all be performed during validation.
	ConfigValidators(context.Context) []ConfigValidator
}

// StateStoreWithValidateConfig is an interface type that extends StateStore to include imperative validation.
//
// Declaring validation using this methodology simplifies one-off
// functionality that typically applies to a single StateStore. Any documentation
// of this functionality must be manually added into schema descriptions.
//
// Validation will include ConfigValidators and ValidateConfig if both are
// implemented, in addition to any Attribute or Type validation.
type StateStoreWithValidateConfig interface {
	StateStore

	// ValidateConfig performs the validation.
	ValidateConfig(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}

type StateStoreWithUpgradeConfigState interface {
	StateStore

	// A mapping of prior state store version to current state store schema
	// version upgrade implementations.
	UpgradeConfigState(context.Context) map[int64]ConfigStateUpgrader
}
