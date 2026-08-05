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
		if err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL); err != nil &&
			!errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
			t.Errorf("kill spice dev process group: %v", err)
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
