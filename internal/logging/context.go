package logging

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// Cache the log level so that we can shortcut calling into the tfsdklog package, as on
// very large terraform resources/projects the logging can become the majority of the
// runtime when building a plan.
//
// https://github.com/hashicorp/terraform-plugin-framework/issues/721
var (
	level hclog.Level
)

// InitContext creates SDK logger contexts. The incoming context will
// already have the root SDK logger and root provider logger setup from
// terraform-plugin-go tf6server RPC handlers.
func InitContext(ctx context.Context) context.Context {
	ctx = tfsdklog.NewSubsystem(ctx, SubsystemFramework,
		// All calls are through the Framework* helper functions
		tfsdklog.WithAdditionalLocationOffset(1),
		tfsdklog.WithLevelFromEnv(EnvTfLogSdkFramework),
		// Propagate tf_req_id, tf_rpc, etc. fields
		tfsdklog.WithRootFields(),
	)

	level = hclog.LevelFromString(EnvTfLogSdkFramework)
	if level == hclog.NoLevel {
		level = hclog.DefaultLevel
	}

	return ctx
}
