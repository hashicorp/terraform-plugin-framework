// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
)

type NoOpListResource struct{}

func (*NoOpListResource) ListResourceConfigSchema(_ context.Context, _ list.ListResourceSchemaRequest, _ *list.ListResourceSchemaResponse) {
}

func (*NoOpListResource) List(_ context.Context, _ list.ListRequest, _ *list.ListResultsStream) {
}
