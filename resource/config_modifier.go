package resource

import (
	"context"
)

type ConfigModifiers interface {
	Metadata(context.Context, MetadataRequest, *MetadataResponse)

	ModifyConfig(context.Context, ModifyConfigRequest, *ModifyConfigResponse)
}
