package app

import (
	"reflect"
	"sync"

	skeltype "go.yorun.ai/skel/types"
	vineapp "go.yorun.ai/vine/app"
	"go.yorun.ai/vine/core/di"
	"go.yorun.ai/vine/core/ex"

	skeled "example.com/greeting/skeled/golang"
	"example.com/greeting/src/server/impl"
)

// Test-only adapters inject failures after real side effects. Runtime delivery,
// generated launchers and the production handlers remain on the exercised path.
type greetingTestApp struct{ GreetingApp }

func (a *greetingTestApp) BindCommon(b *di.Binder) {
	a.GreetingApp.BindCommon(b)
	b.BindInstance(asyncDeliveries)
}

func (a *greetingTestApp) EventerInitListeners(add vineapp.ListenerTypeAdder) {
	a.GreetingApp.EventerInitListeners(func(kind reflect.Type, options ...vineapp.ListenerOption) {
		if kind == vineapp.T[*impl.GreetingCreatedListener]() {
			kind = vineapp.T[*retryingGreetingListener]()
		}
		add(kind, options...)
	})
}

func (a *greetingTestApp) TaskerInitRunners(add vineapp.RunnerTypeAdder) {
	a.GreetingApp.TaskerInitRunners(func(kind reflect.Type, options ...vineapp.RunnerOption) {
		if kind == vineapp.T[*impl.GreetingReceiptRunner]() {
			kind = vineapp.T[*retryingReceiptRunner]()
		}
		add(kind, options...)
	})
}

const retryMessage = "exercise event and task retries"

var asyncDeliveries = &deliveryProbe{counts: make(map[string]deliveryCounts)}

type deliveryCounts struct {
	Events int
	Tasks  int
}

type deliveryProbe struct {
	mu     sync.Mutex
	counts map[string]deliveryCounts
}

func (p *deliveryProbe) record(id string, isEvent bool) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	counts := p.counts[id]
	var attempts int
	if isEvent {
		counts.Events++
		attempts = counts.Events
	} else {
		counts.Tasks++
		attempts = counts.Tasks
	}
	p.counts[id] = counts
	return attempts
}

func (p *deliveryProbe) snapshot(id string) deliveryCounts {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.counts[id]
}

type retryingGreetingListener struct {
	skeled.DefaultGreetingCreatedEventListener
	Handler *impl.GreetingCreatedListener `inject:""`
	Probe   *deliveryProbe                `inject:""`
}

func (h *retryingGreetingListener) OnGreetingCreated(event *skeled.GreetingCreatedEvent) {
	h.Handler.OnGreetingCreated(event)
	attempt := h.Probe.record(event.EventId.String(), true)
	if event.Greeting.Message == retryMessage && attempt == 1 {
		ex.PanicNew(ex.ServiceUnavailable, "test failure after task launch")
	}
}

type retryingReceiptRunner struct {
	skeled.DefaultGreetingReceiptTaskRunner
	Handler *impl.GreetingReceiptRunner `inject:""`
	Probe   *deliveryProbe              `inject:""`
}

func (h *retryingReceiptRunner) RunPrepare(jobId skeltype.UUID, greeting skeled.Greeting) {
	h.Handler.RunPrepare(jobId, greeting)
	attempt := h.Probe.record(jobId.String(), false)
	if greeting.Message == retryMessage && attempt == 1 {
		ex.PanicNew(ex.ServiceUnavailable, "test failure after receipt commit")
	}
}
