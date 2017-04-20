package pad

import (
	"fmt"
	"os/exec"
	"time"
)

type actionCommandBuilder interface {
	Build(string, ...string) *exec.Cmd
}

type realBuilder struct{}

func (r realBuilder) Build(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

var builder actionCommandBuilder

func init() {
	builder = &realBuilder{}
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
	Execute() error
}

// NewAction creates an action with the given k(ind) and d(ata)
func NewAction(kind string, name string, out chan<- ActionMessage, args ...interface{}) Action {
	switch kind {
	case ActionType:
		strArgs := make([]string, len(args))
		for i, v := range args {
			strArgs[i] = v.(string)
		}
		return newActionType(name, out, strArgs...)
	case ActionMacro:
		displayOutput := args[0].(bool)
		strArgs := make([]string, len(args)-1)
		for i, v := range args[1:] {
			strArgs[i] = v.(string)
		}
		return newActionMacro(name, out, displayOutput, strArgs...)
	case ActionPomodoro:
		d := args[0].(time.Duration)
		return newActionPomodoro(name, out, d)
	}
	return nil
}

// Concrete actions

type actionType struct {
	name string
	out  chan<- ActionMessage
	args []string
}

func newActionType(name string, out chan<- ActionMessage, args ...string) *actionType {
	a := new(actionType)
	a.name = name
	a.out = out
	a.args = args
	return a
}

func (a *actionType) Execute() error {
	cmd := builder.Build("cliclick", a.args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s (%v)", string(out), err)
	}
	return nil
}

type actionMacro struct {
	name          string
	out           chan<- ActionMessage
	displayOutput bool
	args          []string
}

func newActionMacro(name string, out chan<- ActionMessage, display bool, args ...string) *actionMacro {
	a := new(actionMacro)
	a.name = name
	a.out = out
	a.displayOutput = display
	a.args = args
	return a
}

func (a *actionMacro) Execute() error {
	cmd := builder.Build(a.args[0], a.args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s (%v)", string(out), err)
	}
	if a.displayOutput {
		a.out <- ActionMessage{ActionName: a.name, Notify: string(out), State: false}
	}
	return nil
}

type actionPomodoro struct {
	name     string
	out      chan<- ActionMessage
	pomodoro *Pomodoro
	progress chan byte
}

func newActionPomodoro(name string, out chan<- ActionMessage, duration time.Duration) *actionPomodoro {
	a := new(actionPomodoro)
	a.name = name
	a.out = out
	a.progress = make(chan byte, 100)
	go a.readProgress()
	a.pomodoro = NewPomodoro(duration, a.progress)
	return a
}

func (a *actionPomodoro) Execute() error {
	if a.pomodoro.IsRunning() {
		a.pomodoro.Cancel()
	} else {
		a.pomodoro.Start()
	}
	return nil
}

func (a *actionPomodoro) readProgress() {
	for {
		p, more := <-a.progress
		if more {
			a.out <- ActionMessage{ActionName: a.name, Progress: p, State: true}
		} else {
			// channel closed, let's go
			return
		}
	}
}

// ActionMessage is sent by the Action on the side channel when an asynchronous operation happens
type ActionMessage struct {
	// ActionName can be used to lookup the sending Action
	ActionName string
	// Notify the user of something
	Notify string
	// State of the Action
	State bool
	// Progress of the Action when relevant (e.g. ActionPomodoro)
	Progress byte
}
