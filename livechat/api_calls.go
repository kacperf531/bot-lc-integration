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
const APIURL = "https://api.labs.livechatinc.com/v3.5"
const SendEventURL = APIURL + "/agent/action/send_event"
const SetRoutingStatusURL = APIURL + "/agent/action/set_routing_status"
const TransferAgentURL = APIURL + "/agent/action/transfer_chat"
const CreateBotURL = APIURL + "/configuration/action/create_bot"

type TokenDetails struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SendEventRequest struct {
	ChatID string          `json:"chat_id"`
	Event  json.RawMessage `json:"event"`
}

type LivechatAPIClient struct {
	HTTPClient http.Client
	Header     http.Header
}

func (c *LivechatAPIClient) sendRequest(url string, payload []byte) (*http.Response, error) {
	// all requests to API are POST so the method is hardcoded here
	r, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	r.Header = c.Header
	response, err := c.HTTPClient.Do(r)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		responseBody, _ := ioutil.ReadAll(response.Body)
		log.Fatalf("Livechat API rejected %s request with message: %s", url, string(responseBody))
	}
	return response, nil
}

func (c *LivechatAPIClient) GetAuthToken(code, clientID, clientSecret, redirectURI string) (*TokenDetails, error) {
	URL, _ := url.Parse(TokenURL)
	q := URL.Query()
	q.Set("grant_type", "authorization_code")
	q.Set("code", code)
	q.Set("client_id", clientID)
	q.Set("client_secret", clientSecret)
	q.Set("redirect_uri", redirectURI)
	URL.RawQuery = q.Encode()
	response, err := c.sendRequest(URL.String(), []byte{})
	if err != nil {
		return nil, fmt.Errorf("Error occured while exchanging code for token: %v", err)
	}
	tokenDetails := &TokenDetails{}
	json.NewDecoder(response.Body).Decode(&tokenDetails)
	return tokenDetails, nil
}

func (c *LivechatAPIClient) SendEvent(chatID string, event json.RawMessage) error {
	payload, _ := json.Marshal(SendEventRequest{ChatID: chatID, Event: event})
	_, err := c.sendRequest(SendEventURL, payload)
	if err != nil {
		return fmt.Errorf("There was a problem with sending event, details: %v", err)
	}
	return nil
}

func (c *LivechatAPIClient) CreateBot(name string) (string, error) {
	payload, _ := json.Marshal(map[string]string{"name": name})
	response, err := c.sendRequest(CreateBotURL, payload)
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

func (c *LivechatAPIClient) SetRoutingStatus(status, agentID string) error {
	payload, _ := json.Marshal(map[string]string{"status": status, "agent_id": agentID})
	_, err := c.sendRequest(SetRoutingStatusURL, payload)
	if err != nil {
		return err
	}
	return nil
}

func (c *LivechatAPIClient) TransferChat(chatID string) error {
	payload, _ := json.Marshal(map[string]string{"id": chatID})
	_, err := c.sendRequest(TransferAgentURL, payload)
	if err != nil {
		return err
	}
	return nil
}
