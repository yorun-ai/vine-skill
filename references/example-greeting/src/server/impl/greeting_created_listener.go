package impl

import skeled "example.com/greeting/skeled/golang"

type GreetingCreatedListener struct {
	skeled.DefaultGreetingCreatedEventListener
	Tasks skeled.GreetingReceiptTaskLauncher `inject:""`
}

func (h *GreetingCreatedListener) OnGreetingCreated(event *skeled.GreetingCreatedEvent) {
	// Reuse the event ID so every delivery requests the same logical job.
	h.Tasks.LaunchPrepare(event.EventId, event.Greeting)
}
