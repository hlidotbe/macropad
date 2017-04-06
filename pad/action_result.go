package pad

// ActionResult describe what the orchestrator should do after completion of the action
type ActionResult struct {
	// Success will be true if Action was executed successfully
	Success bool
	// Notify tells the orchestrator to display something using terminal notifier
	Notify string
	// Set LED state for key
	Set string
}
