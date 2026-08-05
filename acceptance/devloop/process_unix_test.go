//go:build !windows

package devloop_test

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func configureProcessTree(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func stopProcessTree(t *testing.T, command *exec.Cmd, done <-chan error) {
	t.Helper()
	if command.Process != nil {
		// Spice owns application shutdown. Deliver the same termination signal a
		// real shell or service manager sends so it can interrupt and reap the
		// application's separate process group before the tool exits.
		if err := syscall.Kill(-command.Process.Pid, syscall.SIGTERM); err != nil &&
			!errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
			t.Errorf("stop spice dev process group: %v", err)
		}
	}
	waitForProcess(t, done)
}

func waitForProcess(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Error("timed out waiting for spice dev process tree")
	}
}
