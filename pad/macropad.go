package pad

import "io"

// Macropad is the main orchestrator between all components
type Macropad struct {
	serial  io.ReadWriter
	actions []Action
}
