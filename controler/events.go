package controler

import (
	"encoding/json"
	"strings"
)

//go:generate go-enum -f=$GOFILE --marshal --lower
/*
ENUM(
Created, Removed, Started, Restarted, Stopped, Updated, Enabled, Disabled, Joined, Leaved
)
*/
type Event int

func (x Event) MarshalJSON() ([]byte, error) {
	return []byte("\"" + strings.ToLower(x.String()) + "\""), nil
}

func (x *Event) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	return x.UnmarshalText([]byte(str))
}

type SystemEvent struct {
	Type Event  `json:"type"`
	Name string `json:"name"`
}

func WithStateFilter(events <-chan SystemEvent) <-chan SystemEvent {
	state := make(map[string]bool)
	res := make(chan SystemEvent)
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
