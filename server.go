package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kacperf531/bot-lc-integration/livechat"
)

type BotServer struct {
	LivechatAPI             livechat.LivechatAPIClient
	RichMessageTemplate     json.RawMessage
	OAuthClientID           string
	OAuthClientSecret       string
	OAuthClientRedirectURI  string
	IntegrationDataRequired bool
}
type RichMessage struct {
	Text     string `json:"text"`
	Type     string `json:"type"`
	Postback struct {
		Id string `json:"id"`
	} `json:"postback"`
}
type IncomingEvent struct {
	ChatID   string      `json:"chat_id"`
	ThreadID string      `json:"thread_id"`
	Event    RichMessage `json:"event"`
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

type InstallationData struct {
	RefreshToken string `json:"refresh_token"`
	BotID        string `json:"bot_id"`
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

func (bs *BotServer) InstallHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	fmt.Fprint(w, "Thanks for using my app - kacperf531")
	tokenDetails, err := bs.LivechatAPI.GetAuthToken(code, bs.OAuthClientID, bs.OAuthClientSecret, bs.OAuthClientRedirectURI)
	if err != nil {
		log.Fatalf("There was an error during installation when exchanging the authorization code for token %v", err)
	}
	bs.LivechatAPI.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenDetails.AccessToken))
	botID, err := bs.LivechatAPI.CreateBot("Aquarius")
	if err != nil {
		log.Fatalf("Error when creating the bot %v", err)
	}
	installationData := InstallationData{RefreshToken: tokenDetails.RefreshToken, BotID: botID}
	installationDataBytes, err := json.Marshal(installationData)
	if err != nil {
		log.Fatalf("Could not save installation details due to an error %v", err)
	}
	os.WriteFile("installation_data.json", installationDataBytes, 0644)
	bs.LivechatAPI.Header.Set("X-Author-ID", botID)
	if err != nil {
		log.Fatalf("Could not create the bot due to an error: %v", err)
	}
	err = bs.LivechatAPI.SetRoutingStatus("accepting_chats", botID)
	if err != nil {
		log.Fatalf("Could not update bot's status due to an error: %v", err)
	}
	bs.IntegrationDataRequired = false
}

func (bs *BotServer) ReplyHandler(w http.ResponseWriter, r *http.Request) {
	if bs.IntegrationDataRequired {
		log.Fatal("Received request from Livechat, but can't respond due to missing installation data.")
	}

	var wh Webhook
	json.NewDecoder(r.Body).Decode(&wh)
	switch {
	case wh.Action == "incoming_event":
		incomingEvent, err := unmarshalIncomingEvent(wh.Payload)
		if err != nil {
			log.Fatalf("Received `incoming_event` webhook but failed to unmarshal it due to error %v", err)
		}
		bs.SendRichMessageReply(incomingEvent.Event, incomingEvent.ChatID)
	case wh.Action == "incoming_chat":
		incomingChat, err := unmarshalIncomingChat(wh.Payload)
		if err != nil {
			log.Fatalf("Received `incoming_chat` webhook but failed to unmarshal it due to error %v", err)
		}
		bs.SendRichMessage(incomingChat.Chat.ID)
	}
}

func (bs *BotServer) SendRichMessageReply(event RichMessage, chatID string) {
	switch event.Postback.Id {
	case "just_browsing":
		bs.SendMessage(chatID, "Sure, let me know if you have any questions.")
	case "transfer_to_agent":
		bs.SendMessage(chatID, "Granted, you will be redirected to talk with the agent.")
		bs.TransferChatToAgent(chatID)
	case "continue_chat":
		bs.SendMessage(chatID, "Ok what would you like to talk about?")
	default:
		bs.SendMessage(chatID, "Well, sorry I'm not smart enough to help you (yet)")
	}
}

func (bs *BotServer) TransferChatToAgent(chatID string) {
	err := bs.LivechatAPI.TransferChat(chatID)
	if err != nil {
		log.Printf("Could not transfer chat due to error %s", err)
		bs.SendMessage(chatID, "Sorry, currently no agents available")
	}
}

func (bs *BotServer) SendMessage(chatID, text string) {
	event, err := json.Marshal(map[string]string{"type": "message", "text": text})
	if err != nil {
		log.Fatal(err)
	}
	err = bs.LivechatAPI.SendEvent(chatID, event)
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

func (bs *BotServer) SetClientHeadersFromInstallationData(installationDataBytes []byte) error {
	var installationData InstallationData
	err := json.Unmarshal(installationDataBytes, &installationData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal installation data")
	}
	authToken, err := bs.LivechatAPI.GetAuthTokenFromRefresh(installationData.RefreshToken, bs.OAuthClientID, bs.OAuthClientSecret)
	if err != nil {
		return err
	}
	bs.LivechatAPI.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	bs.LivechatAPI.Header.Set("X-Author-Id", installationData.BotID)
	bs.IntegrationDataRequired = false
	return nil
}
