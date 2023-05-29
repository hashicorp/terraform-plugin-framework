// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func testNewRawState(t *testing.T, jsonMap map[string]interface{}) *tfprotov6.RawState {
	t.Helper()

	rawStateJSON, err := json.Marshal(jsonMap)

	if err != nil {
		t.Fatalf("unable to create RawState JSON: %s", err)
	}

	return &tfprotov6.RawState{
		JSON: rawStateJSON,
	}
}
