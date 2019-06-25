package controler

type EventType int

const (
	EventCreated   EventType = 1
	EventRemoved   EventType = 2
	EventStarted   EventType = 3
	EventRestarted EventType = 4
	EventStopped   EventType = 5
	EventUpdated   EventType = 6
	EventEnabled   EventType = 7
	EventDisabled  EventType = 8
)

func (et EventType) String() string {
	switch et {
	case EventCreated:
		return "created"
	case EventRemoved:
		return "removed"
	case EventStarted:
		return "started"
	case EventRestarted:
		return "restarted"
	case EventStopped:
		return "stopped"
	case EventUpdated:
		return "updated"
	case EventEnabled:
		return "enabled"
	case EventDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

type Event struct {
	Type EventType
	Name string
}

func WithStateFilter(events <-chan Event) <-chan Event {
	state := make(map[string]bool)
	res := make(chan Event)
	go func() {
		defer close(res)
		for event := range events {
			ok := false
			switch event.Type {
			case EventStarted:
				ok = !state[event.Name]
				state[event.Name] = true
			case EventStopped:
				running, exists := state[event.Name]
				ok = !exists || running
				state[event.Name] = false
			default:
				ok = true
			}
			if ok {
				res <- event
			}
		}
	}()
	return res
}
