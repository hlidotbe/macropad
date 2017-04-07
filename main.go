package main

import (
	"io"
	"os"

	"github.com/hlidotbe/macropad/pad"
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

func main() {
	s := new(rw)
	s.in = os.Stdin
	s.out = os.Stdout
	orch := pad.NewOchestrator(s)

	orch.Run()
}
