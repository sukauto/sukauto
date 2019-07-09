package integration

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sukauto/controler"
	"text/template"
)

type ExtraTelegram struct {
	Enable   bool   `long:"enable" env:"ENABLE" description:"Enable telegram events"`
	Token    string `long:"token" env:"TOKEN" description:"Telegram BOT token"`
	ChatID   string `long:"chat-id" env:"CHAT_ID" description:"Telegram chat or channel id"`
	Template string `long:"template" env:"TEMPLATE" description:"Template to send" default:"sukauto: {{.Name}} {{.Type}}"`
}

func (et ExtraTelegram) Run(events <-chan controler.SystemEvent) error {
	if !et.Enable {
		return nil
	}
	tpl, err := template.New("").Parse(et.Template)
	if err != nil {
		return err
	}
	for event := range events {
		if err := et.send(event, tpl); err != nil {
			log.Println("[ERROR]", "send to telegram:", err)
		}
	}
	return nil
}

func (et *ExtraTelegram) send(event controler.SystemEvent, tpl *template.Template) error {
	var URL = "https://api.telegram.org/bot" + et.Token + "/sendMessage"
	var msg struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}
	buff := &bytes.Buffer{}
	err := tpl.Execute(buff, event)
	if err != nil {
		return err
	}
	msg.Text = buff.String()
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
