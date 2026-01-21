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
}
