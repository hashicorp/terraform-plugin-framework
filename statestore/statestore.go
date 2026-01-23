// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package statestore

import (
	"context"
)

type StateStore interface {
	// Metadata should return the full name of the state store, such
	// as examplecloud_store.
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// Schema should return the schema for this state store.
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	// TODO: update docs
	// ConfigureStateStore -> the RPC itself, responsible for combining provider-level data (from ConfigureProvider RPC) and
	// state store configuration into a response value that will be set on the server for this state store type
	ConfigureStateStore(context.Context, ConfigureStateStoreRequest, *ConfigureStateStoreResponse)
}

type StateStoreWithConfigure interface {
	StateStore

	// TODO: update docs
	// Configure -> called before every method to interact with a state store implementation, i.e. dependency injection
	// Similar to all of the other configure methods in framework.
	Configure(context.Context, ConfigureRequest, *ConfigureResponse)
}
