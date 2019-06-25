package controler

import "sync"

type ServiceEvents interface {
	Created(name string)
	Removed(name string)
	Started(name string)
	Restarted(name string)
	Stopped(name string)
	Updated(name string)
	Enabled(name string)
	Disabled(name string)
}

type EventSubscriber interface {
	Subscribe(listener ServiceEvents)
}

type ServiceEventsEmitter interface {
	ServiceEvents
	EventSubscriber
}

func NewEventEmitter() ServiceEventsEmitter {
	return &eventEmitter{}
}

type eventEmitter struct {
	listeners []ServiceEvents
	lock      sync.RWMutex
}

func (ee *eventEmitter) Created(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Created(name)
	}
}

func (ee *eventEmitter) Removed(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Removed(name)
	}
}

func (ee *eventEmitter) Started(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Started(name)
	}
}

func (ee *eventEmitter) Restarted(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Restarted(name)
	}
}

func (ee *eventEmitter) Stopped(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Stopped(name)
	}
}

func (ee *eventEmitter) Updated(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Updated(name)
	}
}

func (ee *eventEmitter) Enabled(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Enabled(name)
	}
}

func (ee *eventEmitter) Disabled(name string) {
	for _, listener := range ee.copyListeners() {
		listener.Disabled(name)
	}
}

func (ee *eventEmitter) Subscribe(listener ServiceEvents) {
	ee.lock.Lock()
	defer ee.lock.Unlock()
	ee.listeners = append(ee.listeners, listener)
}

func (ee *eventEmitter) copyListeners() []ServiceEvents {
	ee.lock.RLock()
	defer ee.lock.RUnlock()
	var ans = make([]ServiceEvents, len(ee.listeners))
	copy(ans, ee.listeners)
	return ans
}
