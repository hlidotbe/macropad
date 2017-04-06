package pad

import "time"

// A Pomodoro implements a ticker and send its progress on a given channel
type Pomodoro struct {
	duration time.Duration
	tick     time.Duration
	output   chan<- byte
	cancel   chan bool
	ticker   *time.Ticker
	counter  byte
}

func NewPomodoro(d time.Duration, t time.Duration, out chan<- byte) *Pomodoro {
	p := new(Pomodoro)
	p.duration = d
	p.tick = t
	p.output = out
	return p
}

// Start the pomodoro
func (p *Pomodoro) Start() {
	p.counter = 0
	p.cancel = make(chan bool, 1)
	p.ticker = time.NewTicker(p.tick)
	go p.handleTicker()
}

// Cancel the operation
func (p *Pomodoro) Cancel() {
	p.ticker.Stop()
	p.cancel <- true
}

func (p *Pomodoro) handleTicker() {
	for {
		select {
		case <-p.ticker.C:
			p.output <- p.counter
			p.counter += 1
			break
		case <-p.cancel:
			return
		}
	}
}
