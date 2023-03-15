package livechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const TokenURL = "https://accounts.labs.livechat.com/v2/token"
const SendEventURL = "https://api.labs.livechatinc.com/v3.5/agent/action/send_event"
const SetRoutingStatusURL = "https://api.labs.livechatinc.com/v3.5/agent/action/set_routing_status"
const TransferAgentURL = "https://api.labs.livechatinc.com/v3.5/agent/action/transfer_chat"
const CreateBotURL = "https://api.labs.livechatinc.com/v3.5/configuration/action/create_bot"
const BotName = "Aquarius"

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SendEventRequest struct {
	ChatID string          `json:"chat_id"`
	Event  json.RawMessage `json:"event"`
}

func sendRequest(c *http.Client, url string, payload []byte, header http.Header) (*http.Response, error) {
	// all requests to API are POST so the method is hardcoded here
	r, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	r.Header = header
	response, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		responseBody, _ := ioutil.ReadAll(response.Body)
		log.Fatalf("Livechat API rejected %s request with message: %s", url, string(responseBody))
	}
	return response, nil
}

func GetAuthToken(c *http.Client, code, clientID, clientSecret, redirectURI string) (*TokenDetails, error) {
	URL, _ := url.Parse(TokenURL)
	q := URL.Query()
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	q.Set("client_id", clientID)
	q.Set("client_secret", clientSecret)
	q.Set("redirect_uri", redirectURI)
	URL.RawQuery = q.Encode()
	response, err := sendRequest(c, URL.String(), []byte{}, http.Header{})
	if err != nil {
		return nil, fmt.Errorf("Error occured while exchanging code for token: %v", err)
	}
	tokenDetails := &TokenDetails{}
	json.NewDecoder(response.Body).Decode(&tokenDetails)
	return tokenDetails, nil
}

func SendEvent(c *http.Client, chatID string, event json.RawMessage, header http.Header) error {
	payload, _ := json.Marshal(SendEventRequest{ChatID: chatID, Event: event})
	_, err := sendRequest(c, SendEventURL, payload, header)
	if err != nil {
		return fmt.Errorf("There was a problem with sending event, details: %v", err)
	}
	return nil
}

func CreateBot(c *http.Client, header http.Header) (string, error) {
	payload, _ := json.Marshal(map[string]string{"name": BotName})
	response, err := sendRequest(c, CreateBotURL, payload, header)
	if err != nil {
		return "", err
	}
	type createBotResponse struct {
		ID string `json:"id"`
	}
	parsedResponse := &createBotResponse{}
	json.NewDecoder(response.Body).Decode(&parsedResponse)
	return parsedResponse.ID, nil
}

func SetRoutingStatus(c *http.Client, status, agentID string, header http.Header) error {
	payload, _ := json.Marshal(map[string]string{"status": status, "agent_id": agentID})
	_, err := sendRequest(c, SetRoutingStatusURL, payload, header)
	if err != nil {
		return err
	}
	return nil
}

func TransferChat(c *http.Client, chatID string, header http.Header) error {
	payload, _ := json.Marshal(map[string]string{"id": chatID})
	_, err := sendRequest(c, TransferAgentURL, payload, header)
	if err != nil {
		return err
	}
	return nil
}
