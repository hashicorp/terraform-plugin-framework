package action

import (
	"context"
)

type Action interface {
	Metadata(context.Context, MetadataRequest, *MetadataResponse)
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	Plan(context.Context, PlanRequest, *PlanResponse)
	Invoke(context.Context, InvokeRequest, *InvokeResponse)
	Cancel(context.Context, CancelRequest, *CancelResponse)
}
