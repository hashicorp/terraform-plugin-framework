// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package action

import "context"

type Action interface {
	// Schema should return the schema for this action.
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	// Metadata should return the full name of the action, such as examplecloud_do_thing.
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	// TODO:Actions: Eventual landing place for all required methods to implement for an action
}
