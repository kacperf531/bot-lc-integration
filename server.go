package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const sendEventURL = "https://api.labs.livechatinc.com/v3.5/agent/action/send_event"
const TokenURL = "https://accounts.labs.livechat.com/v2/token"

type BotServer struct {
	HttpClient             http.Client
	RequestHeader          http.Header
	RichMessageTemplate    json.RawMessage
	OAuthClientID          string
	OAuthClientSecret      string
	OAuthClientRedirectURI string
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

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
		tokenDetails, err := bs.GetAuthToken(&bs.HttpClient, code)
		if err != nil {
			log.Fatalf("There was an error when exchanging the authorization code for token %v", err)
		}
		// TODO: check if `Bearer ` prefix is not needed here
		os.Setenv("TOKEN", tokenDetails.AccessToken)
		// TODO: Store refresh token
		// TODO: Create new bot & set it to accept chats
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

func (bs *BotServer) GetAuthToken(c *http.Client, code string) (*TokenDetails, error) {
	r, err := http.NewRequest("POST", TokenURL, nil)
	if err != nil {
		return nil, err
	}
	q := r.URL.Query()
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	q.Set("client_id", bs.OAuthClientID)
	q.Set("client_secret", bs.OAuthClientSecret)
	q.Set("redirect_uri", bs.OAuthClientRedirectURI)
	r.URL.RawQuery = q.Encode()
	var tokenDetails TokenDetails
	response, err := bs.HttpClient.Do(r)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		rejectionDetails, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("Livechat SSO rejected token request: %s", rejectionDetails)
	}
	json.NewDecoder(response.Body).Decode(&tokenDetails)
	return &tokenDetails, nil
}

func (bs *BotServer) sendRequest(url string, payload []byte) (*http.Response, error) {
	r, _ := http.NewRequest("POST", sendEventURL, bytes.NewBuffer(payload))
	r.Header = bs.RequestHeader
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
	replyEvent := SendEventRequest{ChatID: chatId,
		Event: bs.RichMessageTemplate}
	requestBody, _ := json.Marshal(replyEvent)
	bs.sendRequest(sendEventURL, requestBody)
}
