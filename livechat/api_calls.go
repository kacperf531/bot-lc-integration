package livechat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	TokenURL            = "https://accounts.livechat.com/v2/token"
	APIURL              = "https://api.livechatinc.com/v3.5"
	SendEventURL        = APIURL + "/agent/action/send_event"
	SetRoutingStatusURL = APIURL + "/agent/action/set_routing_status"
	TransferAgentURL    = APIURL + "/agent/action/transfer_chat"
	CreateBotURL        = APIURL + "/configuration/action/create_bot"
)

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

func (c *LivechatAPIClient) sendRequest(url string, payloadBytes []byte) (*http.Response, error) {
	// all requests to API are POST so the method is hardcoded here
	r, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("Error building %s request: %v", url, err)
	}
	r.Header = c.Header
	response, err := c.HTTPClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("Error sending %s request: %v", url, err)
	}
	if response.StatusCode != 200 {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("There was an issue with reading response from API %v", err)
		}
		return nil, fmt.Errorf("Livechat API rejected %s request with message: %s", url, string(responseBody))
	}
	return response, nil
}

func (c *LivechatAPIClient) sendJSONRequest(url string, payload interface{}) (*http.Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Could not marshal provided payload due to an error: %v", err)
	}
	response, err := c.sendRequest(url, payloadBytes)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *LivechatAPIClient) GetAuthToken(code, clientID, clientSecret, redirectURI string) (*TokenDetails, error) {
	URL, err := url.Parse(TokenURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the %s URL: %v", TokenURL, err)
	}
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
	_, err := c.sendJSONRequest(SendEventURL, SendEventRequest{ChatID: chatID, Event: event})
	if err != nil {
		return fmt.Errorf("There was a problem with sending event, details: %v", err)
	}
	return nil
}

func (c *LivechatAPIClient) CreateBot(name string) (string, error) {
	response, err := c.sendJSONRequest(CreateBotURL, map[string]string{"name": name})
	if err != nil {
		return "", err
	}
	parsedResponse := struct {
		ID string `json:"id"`
	}{}
	json.NewDecoder(response.Body).Decode(&parsedResponse)
	return parsedResponse.ID, nil
}

func (c *LivechatAPIClient) SetRoutingStatus(status, agentID string) error {
	_, err := c.sendJSONRequest(SetRoutingStatusURL, map[string]string{"status": status, "agent_id": agentID})
	if err != nil {
		return err
	}
	return nil
}

func (c *LivechatAPIClient) TransferChat(chatID string) error {
	_, err := c.sendJSONRequest(TransferAgentURL, map[string]string{"id": chatID})
	if err != nil {
		return err
	}
	return nil
}
