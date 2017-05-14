package pad

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"time"

	"github.com/hlidotbe/macropad/auxilium"
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
	Stop()
}

// Concrete actions
type actionType struct {
	name string
	out  chan<- ActionMessage
	args []string
}

// NewActionType configure and returns an action that passes args to cliclick
func NewActionType(name string, out chan<- ActionMessage, args ...string) Action {
	a := new(actionType)
	a.name = name
	a.out = out
	a.args = args
	return a
}

func (a *actionType) Execute() error {
	a.out <- ActionMessage{ActionName: a.name, Notify: "", Progress: 127}
	cmd := builder.Build("cliclick", a.args...)
	out, err := cmd.CombinedOutput()
	a.out <- ActionMessage{ActionName: a.name, Notify: "", Progress: 0}
	if err != nil {
		return fmt.Errorf("%s (%v)", string(out), err)
	}
	return nil
}

func (a *actionType) Stop() {

}

type actionMacro struct {
	name          string
	out           chan<- ActionMessage
	displayOutput bool
	args          []string
}

// NewActionMacro configure and return an action executing given command, optionnaly notifying the user with its output
func NewActionMacro(name string, out chan<- ActionMessage, display bool, args ...string) Action {
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
		a.out <- ActionMessage{ActionName: a.name, Notify: string(out), State: 0}
	}
	return nil
}

func (a *actionMacro) Stop() {

}

type actionPomodoro struct {
	name     string
	out      chan<- ActionMessage
	duration time.Duration
	pomodoro *Pomodoro
	progress chan byte
}

// NewActionPomodoro configure and returns a Pomodoro action for the given duration
func NewActionPomodoro(name string, out chan<- ActionMessage, duration time.Duration) Action {
	a := new(actionPomodoro)
	a.name = name
	a.out = out
	a.duration = duration
	return a
}

// IsAProgressAction checks if an action should update its progress state
func IsAProgressAction(a Action) bool {
	_, ok := a.(*actionPomodoro)
	if ok {
		return ok
	}
	_, ok = a.(*actionType)
	return ok
}

func (a *actionPomodoro) setupPomodoro() {
	a.progress = make(chan byte, 100)
	go a.readProgress()
	a.pomodoro = NewPomodoro(a.duration, a.progress)
}

func (a *actionPomodoro) Execute() error {
	if a.pomodoro != nil && a.pomodoro.IsRunning() {
		a.pomodoro.Cancel()
	} else {
		a.setupPomodoro()
		a.pomodoro.Start()
		a.out <- ActionMessage{ActionName: a.name, Progress: 1, State: 0}
	}
	return nil
}

func (a *actionPomodoro) Stop() {
	if a.pomodoro != nil && a.pomodoro.IsRunning() {
		a.pomodoro.Cancel()
	}
}

func (a *actionPomodoro) readProgress() {
	for {
		p, more := <-a.progress
		if more {
			m := math.Floor((2.55 * float64(p)) + 0.5)
			if p == 0 {
				m = 1.0
			}
			a.out <- ActionMessage{ActionName: a.name, Progress: byte(m), State: 0}
		} else {
			log.Println("Action Pomodoro done")
			a.out <- ActionMessage{ActionName: a.name, Progress: 0, State: 0}
			a.pomodoro = nil
			// channel closed, let's go
			return
		}
	}
}

type actionTrack struct {
	name         string
	out          chan<- ActionMessage
	projectLabel string
	projectID    int
	profile      string
	currentTime  *auxilium.TimeTrack
	client       *auxilium.Client
}

// NewActionTrack configure and returns a time track action to auxilium
func NewActionTrack(name string, out chan<- ActionMessage, client *auxilium.Client, projectLabel string, projectID int, profile string) Action {
	a := new(actionTrack)
	a.name = name
	a.out = out
	a.client = client
	a.projectLabel = projectLabel
	a.projectID = projectID
	a.profile = profile
	return a
}

func (a *actionTrack) Execute() error {
	var err error
	if a.currentTime == nil {
		a.currentTime = new(auxilium.TimeTrack)
		a.currentTime.Profile = a.profile
		a.currentTime.ProjectId = a.projectID
		a.currentTime.Status = "running"
		a.currentTime.Billable = true
		a.currentTime.Duration = 0
		a.currentTime.Direction = false
		a.currentTime.Started = time.Now().Format("2006-01-02")
		_, _, err = a.client.TimeTrack.Create(a.currentTime)
		a.out <- ActionMessage{ActionName: a.name, Notify: fmt.Sprintf("Began tracking on %s", a.projectLabel), State: 1}
	} else {
		a.currentTime.Status = "pending"
		_, err = a.client.TimeTrack.Update(a.currentTime)
		a.currentTime = nil
		a.out <- ActionMessage{ActionName: a.name, Notify: fmt.Sprintf("Stopped tracking on %s", a.projectLabel), State: -1}
	}
	return err
}

func (a *actionTrack) Stop() {
	if a.currentTime != nil {
		a.currentTime.Status = "pending"
		a.client.TimeTrack.Update(a.currentTime)
		a.currentTime = nil
		a.out <- ActionMessage{ActionName: a.name, Notify: fmt.Sprintf("Stopped tracking on %s", a.projectLabel), State: -1}
	}
}

// ActionMessage is sent by the Action on the side channel when an asynchronous operation happens
type ActionMessage struct {
	// ActionName can be used to lookup the sending Action
	ActionName string
	// Notify the user of something
	Notify string
	// State of the Action
	State int8
	// Progress of the Action when relevant (e.g. ActionPomodoro)
	Progress byte
}
