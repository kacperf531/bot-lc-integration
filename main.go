package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kacperf531/bot-lc-integration/livechat"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	httpClient := &http.Client{}
	livechatAPIClient := &livechat.LivechatAPIClient{HTTPClient: *httpClient, Header: http.Header{
		"Content-type": {"Application/json"}}}
	richMessageTemplate, _ := os.ReadFile("./rich_message.json")
	server := &BotServer{
		LivechatAPI:            *livechatAPIClient,
		RichMessageTemplate:    richMessageTemplate,
		OAuthClientID:          os.Getenv("CLIENT_ID"),
		OAuthClientSecret:      os.Getenv("CLIENT_SECRET"),
		OAuthClientRedirectURI: os.Getenv("CLIENT_REDIRECT_URI")}
	log.Fatal(http.ListenAndServe(":5000", server))
}
