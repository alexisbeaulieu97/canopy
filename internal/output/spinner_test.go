package output

import (
	"bytes"
	"testing"
	"time"
)

func TestSpinnerStopWaitsForAnimateExit(t *testing.T) {
	t.Parallel()

	spinner := &Spinner{
		writer:  &bytes.Buffer{},
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
		running: true,
		isTTY:   true,
	}

	stopped := make(chan struct{})

	go func() {
		spinner.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		t.Fatal("Stop returned before animate signaled exit")
	case <-time.After(25 * time.Millisecond):
	}

	close(spinner.stopped)

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return after animate signaled exit")
	}
}

func TestSpinnerStopWithMessageWaitsForAnimateExit(t *testing.T) {
	t.Parallel()

	spinner := &Spinner{
		writer:  &bytes.Buffer{},
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
		running: true,
		isTTY:   true,
	}

	stopped := make(chan struct{})

	go func() {
		spinner.StopWithMessage("ok", "done")
		close(stopped)
	}()

	select {
	case <-stopped:
		t.Fatal("StopWithMessage returned before animate signaled exit")
	case <-time.After(25 * time.Millisecond):
	}

	close(spinner.stopped)

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("StopWithMessage did not return after animate signaled exit")
	}
}
