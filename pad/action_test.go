package pad

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type testActionCommandBuilder struct{}

func (r testActionCommandBuilder) Build(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, command)
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestExecute_Type(t *testing.T) {
	oldBuilder := builder
	defer func() { builder = oldBuilder }()

	builder = testActionCommandBuilder{}

	a := NewAction(ActionType, "t:git", "kp:space", "t:epic", "kp:space", "t:live", "kp:enter")
	r := a.Execute()
	if !r.Success {
		t.Errorf("Should have passed, got: %s instead", r.Notify)
	}
}

func TestExecute_Macro(t *testing.T) {
	oldBuilder := builder
	defer func() { builder = oldBuilder }()

	builder = testActionCommandBuilder{}

	a := NewAction(ActionMacro, "open", "https://track.epic.net")
	r := a.Execute()
	if !r.Success {
		t.Errorf("Should have passed, got: %s instead", r.Notify)
	}

}

func TestExecute_Invalid(t *testing.T) {
	a := NewAction("invalid", nil)
	if a != nil {
		t.Error("Invalid action should not yield a value")
	}
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for TestParameterRun.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(3)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	cmd, args := args[0], args[1:]
	switch cmd {
	case "cliclick":
		expected := []string{"t:git", "kp:space", "t:epic", "kp:space", "t:live", "kp:enter"}
		for i, s := range expected {
			if args[i] != s {
				fmt.Fprintf(os.Stderr, "Invalid parameter %s, expected %s", args[i], s)
				os.Exit(-1)
			}
		}
		os.Exit(0)
		break
	case "open":
		os.Exit(0)
		break
	}
}
