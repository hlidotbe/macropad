package pad

import (
	"log"
	"time"
)

// A Pomodoro implements a ticker and send its progress on a given channel
type Pomodoro struct {
	duration time.Duration
	output   chan<- byte
	cancel   chan bool
	ticker   *time.Ticker
	counter  byte
	running  bool
}

// NewPomodoro creates a pomodoro timer
func NewPomodoro(d time.Duration, out chan<- byte) *Pomodoro {
	p := new(Pomodoro)
	p.duration = d
	p.output = out
	p.running = false
	return p
}

// Start the pomodoro
func (p *Pomodoro) Start() {
	p.counter = 0
	p.cancel = make(chan bool, 1)
	p.ticker = time.NewTicker(p.duration / 100)
	p.running = true
	go p.handleTicker()
}

// Cancel the operation
func (p *Pomodoro) Cancel() {
	p.ticker.Stop()
	p.running = false
	p.cancel <- true
	log.Println("Pomodoro done")
}

// IsRunning returns true if the ticker is running
func (p *Pomodoro) IsRunning() bool {
	return p.running
}

func (p *Pomodoro) handleTicker() {
	for {
		select {
		case <-p.ticker.C:
			p.output <- p.counter
			p.counter++
			if p.counter > 100 {
				p.Cancel()
			}
			break
		case <-p.cancel:
			close(p.cancel)
			close(p.output)
			return
		}
	}
}
