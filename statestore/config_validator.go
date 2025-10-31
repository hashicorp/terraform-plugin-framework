// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statestore

import "context"

// ConfigValidator describes reusable state store configuration validation functionality.
type ConfigValidator interface {
	// Description describes the validation in plain text formatting.
	//
	// This information may be automatically added to state store plain text
	// descriptions by external tooling.
	Description(context.Context) string

	// MarkdownDescription describes the validation in Markdown formatting.
	//
	// This information may be automatically added to state store Markdown
	// descriptions by external tooling.
	MarkdownDescription(context.Context) string

	// ValidateStateStore performs the validation.
	//
	// This method name is separate from ConfigValidators in resource and other packages in
	// order to allow generic validators.
	ValidateStateStore(context.Context, ValidateConfigRequest, *ValidateConfigResponse)
}
