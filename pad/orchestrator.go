package pad

import (
	"bufio"
	"fmt"
	"io"
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
			if line[len(line)-1] == '1' {
				continue
			}
			fmt.Printf("Got: %s\n", line)
			a := o.actions[line[0:len(line)-1]]
			if a == nil {
				continue
			}
			a.Execute()
			break
		case msg = <-o.Com:
			fmt.Printf("Msg: %s", msg.ActionName)
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
