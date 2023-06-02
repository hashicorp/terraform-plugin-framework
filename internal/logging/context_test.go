// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklogtest"
)

func TestInitContext(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	ctx := tfsdklogtest.RootLogger(context.Background(), &output)

	// Simulate root logger fields that would have been associated by
	// terraform-plugin-go prior to the InitContext() call.
	ctx = tfsdklog.SetField(ctx, "tf_rpc", "GetProviderSchema")
	ctx = tfsdklog.SetField(ctx, "tf_req_id", "123-testing-123")

	ctx = logging.InitContext(ctx)

	logging.FrameworkTrace(ctx, "test message")

	entries, err := tfsdklogtest.MultilineJSONDecode(&output)

	if err != nil {
		t.Fatalf("unable to read multiple line JSON: %s", err)
	}

	expectedEntries := []map[string]interface{}{
		{
			"@level":    "trace",
			"@message":  "test message",
			"@module":   "sdk.framework",
			"tf_rpc":    "GetProviderSchema",
			"tf_req_id": "123-testing-123",
		},
	}

	if diff := cmp.Diff(entries, expectedEntries); diff != "" {
		t.Errorf("unexpected difference: %s", diff)
	}
}
