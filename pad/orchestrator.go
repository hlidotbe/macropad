package pad

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

// Orchestrator processes input from serial connexion and execute corresponding actions
type Orchestrator struct {
	// Com channel for ActionMessages
	Com       chan ActionMessage
	actions   map[string]Action
	serialIn  *bufio.Reader
	serialOut io.Writer
	input     chan string
	done      chan bool
}

// NewOchestrator returns a configured orchestrator ready to be Run
func NewOchestrator(serial io.ReadWriter) *Orchestrator {
	o := &Orchestrator{
		Com:       make(chan ActionMessage, 10),
		serialIn:  bufio.NewReader(serial),
		serialOut: serial,
		input:     make(chan string, 10),
		done:      make(chan bool),
		actions:   make(map[string]Action),
	}
	go o.readLines()
	return o
}

// Run the orchestrator
func (o *Orchestrator) Run() {
	var msg ActionMessage
	var line string
	for {
		select {
		case line = <-o.input:
			go o.executeAction(line)
			break
		case msg = <-o.Com:
			go o.notifyIfNeeded(msg)
			o.updateState(msg)
			if IsAProgressAction(o.actions[msg.ActionName]) {
				o.updateProgress(msg)
			}
			break
		case <-o.done:
			close(o.Com)
			close(o.input)
			return
		}
	}
}

// Shutdown the orchestrator and cleanup everything
func (o *Orchestrator) Shutdown() {
	o.done <- true
}

func (o *Orchestrator) readLines() {
	for {
		line, _, err := o.serialIn.ReadLine()
		if err == nil {
			o.input <- strings.Trim(string(line), "\n")
		}
	}

}

// RegisterAction for given key
func (o *Orchestrator) RegisterAction(key string, a Action) {
	o.actions[key] = a
}

func (o *Orchestrator) notifyIfNeeded(msg ActionMessage) {
	if len(msg.Notify) == 0 {
		return
	}
	cmd := exec.Command("terminal-notifier", "-message", msg.Notify, "-timeout", "2")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%s (%v)\n", string(out), err)
	}
}

func (o *Orchestrator) executeAction(line string) {
	if line[len(line)-1] == '1' {
		return
	}
	log.Printf("Got: %s\n", line)
	a := o.actions[line[0:len(line)-1]]
	if a == nil {
		return
	}
	err := a.Execute()
	if err != nil {
		log.Println(err)
	}
}

func (o *Orchestrator) updateState(msg ActionMessage) {
	if msg.State == 0 {
		return
	}
	if msg.State == 1 {
		o.serialOut.Write([]byte(fmt.Sprintf("%s1\n", msg.ActionName)))
	} else {
		o.serialOut.Write([]byte(fmt.Sprintf("%s0\n", msg.ActionName)))
	}
	log.Printf("Sent: %s%d\n", msg.ActionName, msg.State)
}

func (o *Orchestrator) updateProgress(msg ActionMessage) {
	o.serialOut.Write([]byte(fmt.Sprintf("%s-%d\n", strings.Replace(msg.ActionName, "K", "P", 1), msg.Progress)))
	log.Printf(fmt.Sprintf("%s-%d\n", strings.Replace(msg.ActionName, "K", "P", 1), msg.Progress))
}
