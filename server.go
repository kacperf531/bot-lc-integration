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

type MessageEvent struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type SendEventRequest struct {
	ChatID string          `json:"chat_id"`
	Event  json.RawMessage `json:"event"`
}

type IncomingEvent struct {
	ChatID   string          `json:"chat_id"`
	ThreadID string          `json:"thread_id"`
	Event    json.RawMessage `json:"event"`
}

type Chat struct {
	ID string `json:"id"`
}

type IncomingChat struct {
	Chat Chat `json:"chat"`
}

type Webhook struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

func unmarshalIncomingEvent(payload json.RawMessage) (IncomingEvent, error) {
	var ie IncomingEvent
	err := json.Unmarshal(payload, &ie)
	if err != nil {
		return ie, err
	}
	return ie, nil
}

func unmarshalIncomingChat(payload json.RawMessage) (IncomingChat, error) {
	var ic IncomingChat
	err := json.Unmarshal(payload, &ic)
	if err != nil {
		return ic, err
	}
	return ic, nil
}

func unmarshalEvent(payload json.RawMessage) (MessageEvent, error) {
	var e MessageEvent
	err := json.Unmarshal(payload, &e)
	if err != nil {
		return e, err
	}
	return e, nil
}

func (bs *BotServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var w Webhook
		json.NewDecoder(r.Body).Decode(&w)
		switch {
		case w.Action == "incoming_event":
			incomingEvent, _ := unmarshalIncomingEvent(w.Payload)
			event, _ := unmarshalEvent(incomingEvent.Event)
			bs.SendEventReply(event, incomingEvent.ChatID)
		case w.Action == "incoming_chat":
			incomingChat, _ := unmarshalIncomingChat(w.Payload)
			bs.SendRichMessage(incomingChat.Chat.ID)
		}
	}

}

func (bs *BotServer) sendRequest(url string, payload []byte) (*http.Response, error) {
	r, _ := http.NewRequest("POST", sendEventURL, bytes.NewBuffer(payload))
	r.Header.Set("Authorization", os.Getenv("TOKEN"))
	r.Header.Set("Content-type", "Application/json")
	r.Header.Set("X-Author-Id", os.Getenv("BOT_ID"))
	response, err := bs.HttpClient.Do(r)
	if response.StatusCode != 200 {
		responseBody, _ := ioutil.ReadAll(response.Body)
		log.Fatalf("Livechat API rejected %s request with message: %s", url, string(responseBody))
	}
	return response, err
}

func (bs *BotServer) SendEventReply(event MessageEvent, chatId string) {
	var text string
	switch {
	case event.Text == "I'm just browsing":
		text = "Sure, let me know if you have any questions."
	case event.Text == "I'd rather talk with the agent":
		text = "Granted, you will be redirected to talk with the agent now."
		// TODO: Perform transfer to agent
	default:
		text = "Ok what would you like to talk about?"
	}
	message, _ := json.Marshal(MessageEvent{
		Text: text,
		Type: "message"})
	replyEvent := SendEventRequest{ChatID: chatId,
		Event: message}
	requestBody, _ := json.Marshal(replyEvent)
	bs.sendRequest(sendEventURL, requestBody)
}

func (bs *BotServer) SendRichMessage(chatId string) {
	richMessage, _ := ioutil.ReadFile("./rich_message.json")
	replyEvent := SendEventRequest{ChatID: chatId,
		Event: richMessage}
	requestBody, _ := json.Marshal(replyEvent)
	bs.sendRequest(sendEventURL, requestBody)
}
