// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package schema contains all available schema functionality for actions.
// Action schemas define the structure and value types for configuration data.
// Schemas are implemented via the action.Action type Schema method.
//
// There is currently one type of action schema, which defines how a practitioner can trigger an action,
// as well as what effect the action can have on the state.
//   - [UnlinkedSchema] actions are actions that cannot cause changes to resource states.
package schema
