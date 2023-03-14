package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const sendEventURL = "https://api.labs.livechatinc.com/v3.5/agent/action/send_event"

type BotServer struct {
	HttpClient http.Client
}

type Event struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type SendEventRequest struct {
	ChatID string `json:"chat_id"`
	Event  Event  `json:"event"`
}

type IncomingEvent struct {
	ChatID   string `json:"chat_id"`
	ThreadID string `json:"thread_id"`
	Event    Event  `json:"event"`
}

type Webhook struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func parseIncomingEvent(payload json.RawMessage) (IncomingEvent, error) {
	var ie IncomingEvent
	err := json.Unmarshal(payload, &ie)
	if err != nil {
		return ie, err
	}
	return ie, nil
}

func (bs *BotServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var w Webhook
		json.NewDecoder(r.Body).Decode(&w)
		if w.Action == "incoming_event" {
			incomingEvent, err := parseIncomingEvent(w.Payload)
			if err != nil {
				log.Fatalf("There was a problem when parsing the webhook %+v", w)
			}
			bs.SendEventReply(incomingEvent.Event.Text, incomingEvent.ChatID)
		}
	}

}

func (bs *BotServer) sendRequest(url string, payload []byte) (*http.Response, error) {
	r, _ := http.NewRequest("POST", sendEventURL, bytes.NewBuffer(payload))
	r.Header.Set("Authorization", os.Getenv("TOKEN"))
	r.Header.Set("Content-type", "Application/json")
	r.Header.Set("X-Author-Id", os.Getenv("BOT_ID"))
	return bs.HttpClient.Do(r)
}

func (bs *BotServer) SendEventReply(prompt, chatId string) {
	text := "foo"
	if prompt == "say something else" {
		text = "bar"
	}

	replyEvent := SendEventRequest{ChatID: chatId,
		Event: Event{
			Text: text,
			Type: "message"}}
	requestBody, _ := json.Marshal(replyEvent)
	response, err := bs.sendRequest(sendEventURL, requestBody)
	if err != nil {
		log.Printf(`There was an error sending the "send_event" request, %v`, err)
	}
	if response.StatusCode != 200 {
		responseBody, _ := ioutil.ReadAll(response.Body)
		log.Fatalf("Livechat API rejected send_event request with message: %s", string(responseBody))
	}

}
