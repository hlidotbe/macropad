package pad

import (
	"bufio"
	"io"
	"strings"
)

// Orchestrator processes input from serial connexion and execute corresponding actions
type Orchestrator struct {
	actions   map[string]Action
	serialIn  *bufio.Reader
	serialOut io.Writer
	com       chan ActionMessage
	input     chan string
}

// NewOchestrator returns a configured orchestrator ready to be Run
func NewOchestrator(serial io.ReadWriter) *Orchestrator {
	o := &Orchestrator{
		serialIn:  bufio.NewReader(serial),
		serialOut: serial,
		com:       make(chan ActionMessage, 10),
		input:     make(chan string, 10),
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
			//a := o.actions[line]
			//a.Execute()
			o.serialOut.Write([]byte("Got: "))
			o.serialOut.Write([]byte(line))
			o.serialOut.Write([]byte("\n"))
			break
		case msg = <-o.com:
			o.serialOut.Write([]byte(msg.ActionName))
			o.serialOut.Write([]byte("FF0000\n"))
			break
		}
	}
}

func (o *Orchestrator) readLines() {
	for {
		line, _, err := o.serialIn.ReadLine()
		if err == nil {
			o.input <- strings.Trim(string(line), "\n")
		}
	}

}
