package tg

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sukauto/controler"
)

type tgFunc func(system controler.ServiceController, text string) (string, error)

func tgWithArg(handler func(system controler.ServiceController, arg, text string) (string, error)) tgFunc {
	return func(system controler.ServiceController, text string) (s string, e error) {
		parts := strings.SplitN(text, " ", 2)
		if len(parts) != 2 {
			return "", errors.New("argument not provided")
		}
		return handler(system, parts[0], strings.TrimSpace(parts[1]))
	}
}

func tgNonEmpty(fn tgFunc) tgFunc {
	return func(system controler.ServiceController, text string) (s string, e error) {
		if len(text) == 0 {
			return "", errors.New("argument required")
		}
		return fn(system, text)
	}
}

var commands = map[string]tgFunc{
	"status": func(system controler.ServiceController, name string) (s string, e error) {
		if name != "" {
			status := system.Status(name)
			return statusEmoji[status.Status] + " " + name + " is " + status.Status, nil
		}
		all := system.RefreshStatus()
		var parts []string
		for _, srv := range all.Services {
			parts = append(parts, fmt.Sprintf("%v %v is %v", statusEmoji[srv.Status], srv.Name, srv.Status))
		}
		return strings.Join(parts, "\n"), nil
	},
	"start": tgNonEmpty(func(system controler.ServiceController, text string) (s string, e error) {
		return "OK", system.Run(text)
	}),
	"stop": tgNonEmpty(func(system controler.ServiceController, text string) (s string, e error) {
		return "OK", system.Stop(text)
	}),
	"update": tgNonEmpty(func(system controler.ServiceController, text string) (s string, e error) {
		return "OK", system.Update(text)
	}),
	"restart": tgNonEmpty(func(system controler.ServiceController, text string) (s string, e error) {
		return "OK", system.Restart(text)
	}),
}

func init() {
	commands["help"] = func(system controler.ServiceController, text string) (s string, e error) {
		var names []string
		for k := range commands {
			names = append(names, k)
		}
		sort.Strings(names)
		return "commands:\n\n" + strings.Join(names, "\n"), nil
	}
}
