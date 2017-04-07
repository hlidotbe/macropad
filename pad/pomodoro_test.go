package pad

import (
	"testing"
	"time"
)

func TestPomodoroStart(t *testing.T) {
	const expectedTicks = 100

	ticks := make(chan byte, expectedTicks)
	p := NewPomodoro(time.Millisecond*100, ticks)

	p.Start()
	time.Sleep(time.Millisecond * 200)
	if len(ticks) != expectedTicks {
		t.Errorf("Expected %d ticks, got %d", expectedTicks, len(ticks))
		return
	}
	var i byte
	for i = 0; i < expectedTicks; i++ {
		b := <-ticks
		if b != i {
			t.Errorf("Expected tick %d to be %d but got %d", i, i, b)
			return
		}
	}
}

func TestPomodoroCancel(t *testing.T) {
	const maxTicks = 100

	ticks := make(chan byte, maxTicks)
	p := NewPomodoro(time.Millisecond, ticks)

	p.Start()
	time.Sleep(time.Millisecond / 10)
	p.Cancel()
	time.Sleep(time.Millisecond * 2)
	if len(ticks) > maxTicks {
		t.Errorf("Expected less than %d ticks, got %d", maxTicks, len(ticks))
		return
	}
}
