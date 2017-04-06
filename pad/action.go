package pad

import "os/exec"

type actionCommandBuilder interface {
	Build(string, ...string) *exec.Cmd
}

var builder actionCommandBuilder

type realBuilder struct{}

func (r realBuilder) Build(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

const (
	// ActionPomodoro action toggles a pomorodo timer of the given duration
	ActionPomodoro string = "pomodoro"
	// ActionType action send the configured string to the foremost active window
	ActionType = "type"
	// ActionTrack toggles time tracking on the given project
	ActionTrack = "track"
	// ActionMacro silently execute a command
	ActionMacro = "macro"
)

// Action describe what a keypress should do
type Action interface {
	Execute() ActionResult
}

// NewAction creates an action with the given k(ind) and d(ata)
func NewAction(k string, args ...interface{}) Action {
	switch k {
	case ActionType:
		a := new(actionType)
		strArgs := make([]string, len(args))
		for i, v := range args {
			strArgs[i] = v.(string)
		}
		a.args = strArgs
		return a
	case ActionMacro:
		a := new(actionMacro)
		strArgs := make([]string, len(args))
		for i, v := range args {
			strArgs[i] = v.(string)
		}
		a.args = strArgs
		return a
	}
	return nil
}

// Concrete actions

type actionType struct {
	args []string
}

func (a *actionType) Execute() ActionResult {
	cmd := builder.Build("cliclick", a.args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ActionResult{Success: false, Notify: string(out)}
	}

	return ActionResult{Success: true}
}

type actionMacro struct {
	args []string
}

func (a *actionMacro) Execute() ActionResult {
	cmd := builder.Build(a.args[0], a.args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ActionResult{Success: false, Notify: string(out)}
	}

	return ActionResult{Success: true}
}
