package output

import (
	"bytes"
	"os"
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

func TestSpinnerWithWriterNilFallsBackToStderr(t *testing.T) {
	t.Parallel()

	spinner := NewSpinner("loading").WithWriter(nil)

	if spinner.writer != os.Stderr {
		t.Fatal("expected nil writer to fall back to stderr")
	}

	if spinner.isTTY != isWriterTTY(os.Stderr) {
		t.Fatal("expected TTY state to match stderr")
	}
}

func TestSpinnerWithWriterIgnoredWhileRunning(t *testing.T) {
	t.Parallel()

	initialWriter := &bytes.Buffer{}
	replacementWriter := &bytes.Buffer{}
	spinner := NewSpinner("loading").WithWriter(initialWriter)
	spinner.running = true

	spinner.WithWriter(replacementWriter)

	if spinner.writer != initialWriter {
		t.Fatal("expected running spinner to keep its current writer")
	}

	if spinner.isTTY != isWriterTTY(initialWriter) {
		t.Fatal("expected TTY state to remain unchanged while running")
	}
}
