// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerCancelInFlightContexts(t *testing.T) {
	t.Parallel()

	// let's test and make sure the code we use to Stop will actually
	// cancel in flight contexts how we expect and not, y'know, crash or
	// something

	// first, let's create a bunch of goroutines
	wg := new(sync.WaitGroup)
	s := &Server{}
	testCtx := context.Background()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			ctx = s.registerContext(ctx)
			select {
			case <-time.After(time.Second * 10):
				t.Error("timed out waiting to be canceled")
				return
			case <-ctx.Done():
				return
			}
		}()
	}
	// avoid any race conditions around canceling the contexts before
	// they're all set up
	//
	// we don't need this in prod as, presumably, Terraform would not keep
	// sending us requests after it told us to stop
	time.Sleep(200 * time.Millisecond)

	s.cancelRegisteredContexts(testCtx)

	wg.Wait()
	// if we got here, that means that either all our contexts have been
	// canceled, or we have an error reported
}

func testNewDynamicValue(t *testing.T, schemaType tftypes.Type, schemaValue map[string]tftypes.Value) *tfprotov5.DynamicValue {
	t.Helper()

	dynamicValue, err := tfprotov5.NewDynamicValue(schemaType, tftypes.NewValue(schemaType, schemaValue))

	if err != nil {
		t.Fatalf("unable to create DynamicValue: %s", err)
	}

	return &dynamicValue
}

func testNewTfprotov5RawState(t *testing.T, jsonMap map[string]interface{}) *tfprotov5.RawState {
	t.Helper()

	rawStateJSON, err := json.Marshal(jsonMap)

	if err != nil {
		t.Fatalf("unable to create RawState JSON: %s", err)
	}

	return &tfprotov5.RawState{
		JSON: rawStateJSON,
	}
}

func testNewTfprotov6RawState(t *testing.T, jsonMap map[string]interface{}) *tfprotov6.RawState {
	t.Helper()

	rawStateJSON, err := json.Marshal(jsonMap)

	if err != nil {
		t.Fatalf("unable to create RawState JSON: %s", err)
	}

	return &tfprotov6.RawState{
		JSON: rawStateJSON,
	}
}
