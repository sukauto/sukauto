package controler

import (
	"log"
	"os"
	"os/exec"
	"time"
)

func WithBackgroundCheck(events <-chan SystemEvent, interval time.Duration, controller ServiceController) <-chan SystemEvent {
	ans := make(chan SystemEvent)
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
						ans <- SystemEvent{Type: EventStarted, Name: status.Name}
					} else {
						ans <- SystemEvent{Type: EventStopped, Name: status.Name}
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

func WithScriptRunner(events <-chan SystemEvent, command string) <-chan SystemEvent {
	ans := make(chan SystemEvent)
	go func() {
		defer close(ans)
		for event := range events {
			cmd := exec.Command(SHELL, "-c", command)
			cmd.Env = os.Environ()
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Env = append(cmd.Env, EnvService+"="+event.Name)
			cmd.Env = append(cmd.Env, EnvEvent+"="+event.Type.String())
			if err := cmd.Run(); err != nil {
				log.Println("failed run script", command, ":", err)
			}
			ans <- event
		}
	}()
	return ans
}

func Tee(events <-chan SystemEvent) (<-chan SystemEvent, <-chan SystemEvent) {
	a, b := make(chan SystemEvent), make(chan SystemEvent)
	go func() {
		defer close(a)
		defer close(b)
		for event := range events {
			a <- event
			b <- event
		}
	}()
	return a, b
}
