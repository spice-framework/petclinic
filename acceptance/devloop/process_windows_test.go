//go:build windows

package devloop_test

import (
	"os/exec"
	"strconv"
	"testing"
	"time"
)

func configureProcessTree(_ *exec.Cmd) {}

func stopProcessTree(t *testing.T, command *exec.Cmd, done <-chan error) {
	t.Helper()
	if command.Process != nil {
		process := exec.Command( // #nosec G204 -- PID comes from the test-owned process.
			"taskkill",
			"/PID",
			strconv.Itoa(command.Process.Pid),
			"/T",
			"/F",
		)
		if output, err := process.CombinedOutput(); err != nil {
			select {
			case <-done:
				return
			default:
				t.Errorf("kill spice dev process tree: %v\n%s", err, output)
			}
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
