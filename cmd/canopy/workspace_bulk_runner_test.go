package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRunBulkWorkspaceOperationsSuccess(t *testing.T) {
	t.Parallel()

	report := runBulkWorkspaceOperations(context.Background(), []string{"ws-1", "ws-2"}, bulkWorkspaceRunOptions[string]{
		Parallelism: 2,
	}, func(_ context.Context, id string) (string, error) {
		return id + "-done", nil
	})

	if report.Cancelled {
		t.Fatal("expected non-cancelled report")
	}

	if report.FirstErr != nil {
		t.Fatalf("expected nil first error, got %v", report.FirstErr)
	}

	if len(report.SuccessIDs) != 2 {
		t.Fatalf("expected 2 success IDs, got %v", report.SuccessIDs)
	}

	successes := map[string]bool{}
	for _, id := range report.SuccessIDs {
		successes[id] = true
	}

	if !successes["ws-1"] || !successes["ws-2"] {
		t.Fatalf("unexpected success IDs: got %v", report.SuccessIDs)
	}

	if len(report.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(report.Results))
	}

	if report.Results[0].Value != "ws-1-done" || report.Results[1].Value != "ws-2-done" {
		t.Fatalf("unexpected result values: %+v", report.Results)
	}
}

func TestRunBulkWorkspaceOperationsFailure(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	report := runBulkWorkspaceOperations(context.Background(), []string{"ws-1", "ws-2"}, bulkWorkspaceRunOptions[struct{}]{
		Parallelism: 2,
	}, func(_ context.Context, id string) (struct{}, error) {
		if id == "ws-2" {
			return struct{}{}, expectedErr
		}

		return struct{}{}, nil
	})

	if report.FirstErr == nil || !errors.Is(report.FirstErr, expectedErr) {
		t.Fatalf("expected first error %v, got %v", expectedErr, report.FirstErr)
	}

	if got, want := report.FailedIDs, []string{"ws-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected failed IDs: got %v want %v", got, want)
	}
}

func TestRunBulkWorkspaceOperationsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	report := runBulkWorkspaceOperations(ctx, []string{"ws-1", "ws-2"}, bulkWorkspaceRunOptions[struct{}]{
		Parallelism: 2,
	}, func(context.Context, string) (struct{}, error) {
		t.Fatal("execute should not be called when the context is already cancelled")
		return struct{}{}, nil
	})

	if !report.Cancelled {
		t.Fatal("expected cancelled report")
	}

	if len(report.FailedIDs) != 2 {
		t.Fatalf("expected cancelled operations to be recorded as failures, got %d", len(report.FailedIDs))
	}

	for _, res := range report.Results {
		if !errors.Is(res.Err, context.Canceled) {
			t.Fatalf("expected context cancellation error, got %+v", res)
		}
	}
}
