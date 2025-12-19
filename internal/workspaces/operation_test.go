package workspaces

import (
	"errors"
	"reflect"
	"testing"
)

func TestOperationExecute_Success(t *testing.T) {
	t.Parallel()

	var calls []string

	op := NewOperation(nil)
	op.AddStep(func() error {
		calls = append(calls, "action-1")
		return nil
	}, func() error {
		calls = append(calls, "rollback-1")
		return nil
	})
	op.AddStep(func() error {
		calls = append(calls, "action-2")
		return nil
	}, func() error {
		calls = append(calls, "rollback-2")
		return nil
	})

	if err := op.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(calls, []string{"action-1", "action-2"}) {
		t.Fatalf("unexpected calls: %v", calls)
	}
}

func TestOperationExecute_Rollback(t *testing.T) {
	t.Parallel()

	var calls []string

	op := NewOperation(nil)
	op.AddStep(func() error {
		calls = append(calls, "action-1")
		return nil
	}, func() error {
		calls = append(calls, "rollback-1")
		return nil
	})
	op.AddStep(func() error {
		calls = append(calls, "action-2")
		return errors.New("boom")
	}, func() error {
		calls = append(calls, "rollback-2")
		return nil
	})

	err := op.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}

	expected := []string{"action-1", "action-2", "rollback-2", "rollback-1"}
	if !reflect.DeepEqual(calls, expected) {
		t.Fatalf("unexpected calls: %v", calls)
	}
}
