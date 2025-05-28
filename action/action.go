package action

import (
	"context"
)

type SimpleAction interface {
	Metadata(context.Context, MetadataRequest, *MetadataResponse)
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	Plan(context.Context, PlanRequest, *PlanResponse)
	Invoke(context.Context, InvokeRequest, *InvokeResponse)
}

type Action interface {
	Metadata(context.Context, MetadataRequest, *MetadataResponse)
	Schema(context.Context, SchemaRequest, *SchemaResponse)

	Plan(context.Context, PlanRequest, *PlanResponse)
	Invoke(context.Context, InvokeRequest, *InvokeCallbackResponse)
	Cancel(context.Context, CancelRequest, *CancelResponse)
}
