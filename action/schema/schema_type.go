// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import "github.com/hashicorp/terraform-plugin-framework/internal/fwschema"

// TODO:Actions: Implement lifecycle and linked schemas
//
// SchemaType is the interface that an action schema type must implement. Action
// schema types are statically definined in the protocol, so all implementations
// are defined in this package.
//
// SchemaType implementations define how a practitioner can trigger an action, as well
// as what effect the action can have on the state. There are currently three different
// types of actions:
//   - [UnlinkedSchema] actions are actions that cannot cause changes to resource states.
//   - [LifecycleSchema] actions are actions that can cause changes to exactly one resource state.
//   - [LinkedSchema] actions are actions that can cause changes to one or more resource states.
type SchemaType interface {
	fwschema.Schema

	// Action schema types are statically defined in the protocol, so this
	// interface is not meant to be implemented outside of this package
	isActionSchemaType()
}
