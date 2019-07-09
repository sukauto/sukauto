package tg

import "sukauto/controler"

var eventEmoji = map[controler.Event]string{
	controler.EventCreated:   "\u2795",
	controler.EventRemoved:   "🗑️",
	controler.EventStarted:   "👌",
	controler.EventRestarted: "♻️",
	controler.EventStopped:   "✋",
	controler.EventUpdated:   "✔️",
	controler.EventEnabled:   "☑️",
	controler.EventDisabled:  "⛔",
	controler.EventJoined:    "\u26D3",
	controler.EventLeaved:    "❗",
}

var statusEmoji = map[string]string{
	"running": "\u2699",
	"dead":    "⚰️",
}
