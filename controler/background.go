package controler

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func WithBackgroundCheck(events <-chan Event, interval time.Duration, controller ServiceController) <-chan Event {
	ans := make(chan Event)
	go func() {
		defer close(ans)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
	LOOP:
		for {
			select {
			case <-ticker.C:
				statuses := controller.RefreshStatus()
				for _, status := range statuses.Services {
					if status.Status == "running" {
						ans <- Event{Type: EventStarted, Name: status.Name}
					} else {
						ans <- Event{Type: EventStopped, Name: status.Name}
					}
				}
			case event, ok := <-events:
				if !ok {
					break LOOP
				}
				ans <- event
			}
		}
	}()
	return ans
}

func WithScriptRunner(events <-chan Event, command string) <-chan Event {
	ans := make(chan Event)
	go func() {
		defer close(ans)
		for event := range events {
			cmd := exec.Command(SHELL, "-c", command)
			cmd.Env = os.Environ()
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Env = append(cmd.Env, "SERVICE="+event.Name)
			cmd.Env = append(cmd.Env, "EVENT="+event.Type.String())
			if err := cmd.Run(); err != nil {
				log.Println("failed run script", command, ":", err)
			}
			ans <- event
		}
	}()
	return ans
}
