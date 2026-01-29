// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package statestore_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
)

func TestLockInfo_JSON_roundtrip(t *testing.T) {
	t.Parallel()

	lockRequest := statestore.LockRequest{
		StateID:   "test-state-123",
		Operation: "apply",
	}

	lockInfo := statestore.NewLockInfo(lockRequest)

	lockJSON, err := json.Marshal(lockInfo)
	if err != nil {
		t.Fatalf("unexpected error when marshalling lock info to JSON: %s", err.Error())
	}

	var unmarshalledLockInfo statestore.LockInfo
	err = json.Unmarshal(lockJSON, &unmarshalledLockInfo)
	if err != nil {
		t.Fatalf("unexpected error when unmarshalling lock info from JSON: %s", err.Error())
	}

	if diff := cmp.Diff(unmarshalledLockInfo, lockInfo); diff != "" {
		t.Fatalf("unexpected difference between original lock and unmarshalled lock: %s", diff)
	}

	if !lockInfo.Equal(unmarshalledLockInfo) {
		t.Fatalf("locks do not match - got: %v, expected: %v", unmarshalledLockInfo, lockInfo)
	}
}

func TestLockInfo_WorkspaceAlreadyLockedDiagnostic(t *testing.T) {
	t.Parallel()

	req := statestore.LockRequest{
		StateID:   "test-state-123",
		Operation: "apply",
	}

	lockInfo := statestore.LockInfo{
		ID:        "lock-123",
		Operation: req.Operation,
		Who:       "hashicorp@test",
		Created:   time.Date(2026, 1, 29, 0, 0, 0, 0, time.UTC),
	}

	got := statestore.WorkspaceAlreadyLockedDiagnostic(req, lockInfo)
	expectedDiag := diag.NewErrorDiagnostic(
		"Workspace Already Locked",
		`"test-state-123" workspace has a lock held by another client.

Lock Info:
  ID:        lock-123
  Operation: apply
  Who:       hashicorp@test
  Created:   2026-01-29 00:00:00 +0000 UTC
`)

	if diff := cmp.Diff(got, expectedDiag); diff != "" {
		t.Fatalf("unexpected difference: %s", diff)
	}
}
