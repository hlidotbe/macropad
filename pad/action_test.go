package pad

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
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

	com := make(chan ActionMessage, 1)
	a := NewAction(ActionType, "K1", com, "t:git", "kp:space", "t:epic", "kp:space", "t:live", "kp:enter")
	err := a.Execute()
	if err != nil {
		t.Errorf("Should have passed, got: \"%v\" instead", err)
	}
	if len(com) != 0 {
		t.Errorf("Did not expect a message from ActionType, got %v", <-com)
	}
}

func TestExecute_Macro(t *testing.T) {
	oldBuilder := builder
	defer func() { builder = oldBuilder }()

	builder = testActionCommandBuilder{}

	com := make(chan ActionMessage, 1)
	a := NewAction(ActionMacro, "K1", com, false, "open", "https://track.epic.net")
	err := a.Execute()
	if err != nil {
		t.Errorf("Should have passed, got: \"%s\" instead", err)
	}
	if len(com) != 0 {
		t.Errorf("Did not expect a message from ActionType, got %v", <-com)
	}
	a = NewAction(ActionMacro, "K1", com, true, "echo", "test")
	err = a.Execute()
	if err != nil {
		t.Errorf("Should have passed, got: \"%s\" instead", err)
	}
	if len(com) != 1 {
		t.Errorf("Got no notification from action")
		return
	}
	msg := <-com
	if msg.Notify != "[test]\n" {
		t.Errorf("Expected '[test]' got '%v'", msg.Notify)
	}
}

func TestExecute_Pomodoro(t *testing.T) {
	out := make(chan ActionMessage, 100)
	a := NewAction(ActionPomodoro, "K1", out, time.Millisecond)
	err := a.Execute()
	if err != nil {
		t.Errorf("Should have passed, got: '%v' instead", err)
		return
	}
	time.Sleep(time.Millisecond * 10)
	if len(out) != 100 {
		t.Error("Pomodoro probably didn't start")
	}
	for i := 0; i < len(out); i++ {
		<-out
	}
	a.Execute()
	time.Sleep(time.Millisecond / 100)
	err = a.Execute()
	if err != nil {
		t.Errorf("Should have passed, got: '%v' instead", err)
		return
	}
	time.Sleep(time.Millisecond * 10)
	if len(out) >= 100 {
		t.Error("Pomodoro continued")
	}
}

func TestExecute_Invalid(t *testing.T) {
	a := NewAction("invalid", "K1", make(chan ActionMessage))
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
	case "echo":
		fmt.Println(args)
		os.Exit(0)
		break

	}
}
