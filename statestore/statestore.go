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

	// Initialize is called one time, prior to executing any state store RPCs (excluding offline validation) but after
	// the provider is configured, and is when Terraform sends the values the user specified in the state_store configuration
	// block to the provider. These are supplied in the InitializeRequest argument.
	//
	// As this method is always executed after provider configuration, data can be passed from [provider.ConfigureResponse.StateStoreData]
	// to [InitializeRequest.ProviderData]. This provider-level data along with the values from state store configuration are often used
	// to initialize an API client, which can be set to [InitializeResponse.StateStoreData], then eventually stored on the struct implementing
	// the StateStore interface in the [StateStoreWithConfigure.Configure] method.
	Initialize(context.Context, InitializeRequest, *InitializeResponse)

	// TODO: package docs, mention optional nature of locking, just return with no lock ID
	Lock(context.Context, LockRequest, *LockResponse)

	// TODO: package docs, mention it is only called if a lock ID is returned from Lock
	Unlock(context.Context, UnlockRequest, *UnlockResponse)
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
