package pad

import (
	"bytes"
	"io"
	"testing"
	"time"
)

type rw struct {
	in  io.Reader
	out io.Writer
}

func (s *rw) Read(p []byte) (n int, err error) {
	n, err = s.in.Read(p)
	return n, err
}

func (s *rw) Write(p []byte) (n int, err error) {
	n, err = s.out.Write(p)
	return n, err
}

func newRw(in io.Reader, out io.Writer) *rw {
	s := new(rw)
	s.in = in
	s.out = out
	return s
}

type dummyAction struct {
	out chan<- ActionMessage
}

func (a *dummyAction) Execute() error {

	a.out <- ActionMessage{ActionName: "K1", Notify: "", State: false, Progress: 0}

	return nil
}

func TestRun(t *testing.T) {
	in := new(bytes.Buffer)
	out := new(bytes.Buffer)
	s := newRw(in, out)

	orch := NewOchestrator(s)

	go orch.Run()

	in.WriteString("K1")
	time.Sleep(10 * time.Millisecond)

	line, err := out.ReadString('\n')
	if err != nil || line != "Got: K1\n" {
		t.Errorf("Expected 'Got: K1', got: '%s' (%v)", line, err)
	}

	orch.Shutdown()

}
