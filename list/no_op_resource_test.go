// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package list_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type NoOpResource struct{}

func (*NoOpResource) Schema(_ context.Context, _ resource.SchemaRequest, _ *resource.SchemaResponse) {
}

func (*NoOpResource) Create(_ context.Context, _ resource.CreateRequest, _ *resource.CreateResponse) {
}

func (*NoOpResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (*NoOpResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (*NoOpResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}
