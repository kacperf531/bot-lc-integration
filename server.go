package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kacperf531/bot-lc-integration/livechat"
)

type BotServer struct {
	LivechatAPI            livechat.LivechatAPIClient
	RichMessageTemplate    json.RawMessage
	OAuthClientID          string
	OAuthClientSecret      string
	OAuthClientRedirectURI string
}
type MessageEvent struct {
	Text string `json:"text"`
	Type string `json:"type"`
}
type IncomingEvent struct {
	ChatID   string          `json:"chat_id"`
	ThreadID string          `json:"thread_id"`
	Event    json.RawMessage `json:"event"`
}

type IncomingChat struct {
	Chat struct {
		ID string `json:"id"`
	} `json:"chat"`
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
	switch {
	case r.URL.Path == "/install":
		code := r.URL.Query().Get("code")
		fmt.Fprint(w, "Thanks for using my app - kacperf531")
		tokenDetails, err := bs.LivechatAPI.GetAuthToken(code, bs.OAuthClientID, bs.OAuthClientSecret, bs.OAuthClientRedirectURI)
		if err != nil {
			log.Fatalf("There was an error during installation when exchanging the authorization code for token %v", err)
		}
		bs.LivechatAPI.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenDetails.AccessToken))
		// TODO: Store refresh token
		botID, err := bs.LivechatAPI.CreateBot("Aquarius")
		bs.LivechatAPI.Header.Set("X-Author-ID", botID)
		if err != nil {
			log.Fatalf("Could not create the bot due to an error: %v", err)
		}
		err = bs.LivechatAPI.SetRoutingStatus("accepting_chats", botID)
		if err != nil {
			log.Fatalf("Could not update bot's status due to an error: %v", err)
		}
	case r.Method == "POST":
		// TODO: Check if bs.token is set. If not, get it by exchanging refresh token
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
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Whoops I'm afraid it's 404")
	}

}

func (bs *BotServer) SendEventReply(event MessageEvent, chatID string) {
	var text string
	switch {
	case event.Text == "I'm just browsing":
		text = "Sure, let me know if you have any questions."
	case event.Text == "I'd rather talk with the agent":
		text = "Granted, you will be redirected to talk with the agent."
		defer bs.TransferChatToAgent(chatID)
	default:
		text = "Ok what would you like to talk about?"
	}
	bs.SendMessage(chatID, text)
}

func (bs *BotServer) TransferChatToAgent(chatID string) {
	err := bs.LivechatAPI.TransferChat(chatID)
	if err != nil {
		log.Printf("Could not transfer chat due to error %s", err)
		bs.SendMessage(chatID, "Sorry, currently no agents available")
	}
}

func (bs *BotServer) SendMessage(chatID, text string) {
	event, _ := json.Marshal(map[string]string{"type": "message", "text": text})
	err := bs.LivechatAPI.SendEvent(chatID, event)
	if err != nil {
		log.Fatal(err)
	}
}

func (bs *BotServer) SendRichMessage(chatID string) {
	err := bs.LivechatAPI.SendEvent(chatID, bs.RichMessageTemplate)
	if err != nil {
		log.Fatal(err)
	}
}
