package tg

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sukauto/controler"
	"text/template"
	"time"
)

type ExtraTelegram struct {
	Enable   bool    `long:"enable" env:"ENABLE" description:"Enable telegram events"`
	Token    string  `long:"token" env:"TOKEN" description:"Telegram BOT token"`
	ChatID   int64   `long:"chat-id" env:"CHAT_ID" description:"Telegram chat or channel id"`
	Template string  `long:"template" env:"TEMPLATE" description:"Template to sendEvent" default:"sukauto: {{.Name}} {{.Type}}"`
	Admins   []int64 `long:"admins" env:"ADMINS" description:"Administrator user ID" env-delim:","`
}

func (et ExtraTelegram) Run(system controler.ServiceController, events <-chan controler.SystemEvent) error {
	if !et.Enable {
		return nil
	}
	tpl, err := template.New("").Parse(et.Template)
	if err != nil {
		return err
	}
	go et.listenCommands(system)
	for event := range events {
		if err := et.sendEvent(event, tpl); err != nil {
			log.Println("[ERROR]", "sendEvent to telegram:", err)
		}
	}
	return nil
}

func (et *ExtraTelegram) sendEvent(event controler.SystemEvent, tpl *template.Template) error {
	buff := &bytes.Buffer{}
	err := tpl.Execute(buff, event)
	if err != nil {
		return err
	}
	icon := eventEmoji[event.Type]
	return et.sendTo(et.ChatID, icon+" "+buff.String())
}

func (et *ExtraTelegram) sendTo(chatID int64, text string) error {
	var URL = "https://api.telegram.org/bot" + et.Token + "/sendMessage"
	var msg struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}
	msg.Text = text
	msg.ChatID = et.ChatID
	data, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	res, err := http.Post(URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resp, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(string(resp))
	}
	return nil
}

func (et *ExtraTelegram) listenCommands(system controler.ServiceController) {
	var offset int64
	for {
		for _, upd := range et.getUpdates(offset) {
			if upd.ID >= offset {
				offset = upd.ID + 1
			}

			if upd.Message == nil {
				continue
			}
			if upd.Message.From == nil {
				continue
			}
			if upd.Message.From.Bot {
				continue
			}
			var isAdmin bool
			for _, usr := range et.Admins {
				if usr == upd.Message.From.ID {
					isAdmin = true
					break
				}
			}
			if !isAdmin {
				continue
			}

			parts := strings.SplitN(upd.Message.Text, " ", 2)
			cmd := parts[0]
			text := ""
			if len(parts) == 2 {
				text = strings.TrimSpace(parts[1])
			}

			h, ok := commands[cmd]
			if !ok {
				continue
			}
			reply, err := h(system, text)
			if err != nil {
				err = et.sendTo(upd.Chat.ID, "⚠️ "+err.Error())
			} else if reply != "" {
				err = et.sendTo(upd.Chat.ID, reply)
			}
			if err != nil {
				log.Println("[ERROR]", "failed to sendEvent reply to telegram:", err)
			}
		}
	}
}

type tgUpdate struct {
	ID   int64 `json:"update_id"`
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	Message *struct {
		Text string `json:"text"`
		From *struct {
			ID  int64 `json:"id"`
			Bot bool  `json:"is_bot"`
		}
	} `json:"message"`
}

func (et *ExtraTelegram) getUpdates(offset int64) []*tgUpdate {
	const penalty = 10 * time.Second
	var URL = "https://api.telegram.org/bot" + et.Token + "/getUpdates?timeout=10"
	if offset != 0 {
		URL += "&offset=" + strconv.FormatInt(offset, 10)
	}
	res, err := http.Get(URL)
	if err != nil {
		log.Println("[ERROR]", "failed to get updates from telegram:", err)
		time.Sleep(penalty)
		return nil
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println("[ERROR]", "failed to read updates from telegram:", err)
		time.Sleep(penalty)
		return nil
	}
	if res.StatusCode != http.StatusOK {
		log.Println("[ERROR]", "failed to fetch updates from telegram:", string(data))
		time.Sleep(penalty)
		return nil
	}
	var updates struct {
		Updates []*tgUpdate `json:"result"`
	}
	err = json.Unmarshal(data, &updates)
	if err != nil {
		log.Println("[ERROR]", "failed to parse updates from telegram:", err)
		time.Sleep(penalty)
		return nil
	}
	return updates.Updates
}
