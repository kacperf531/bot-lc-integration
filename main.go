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
	header := http.Header{
		"Content-type": {"Application/json"}}
	livechatAPIClient := &livechat.LivechatAPIClient{HTTPClient: *httpClient, Header: header}
	richMessageTemplate, err := os.ReadFile("./rich_message.json")
	if err != nil {
		log.Fatal("Could not find the rich message template")
	}
	server := &BotServer{
		LivechatAPI:             *livechatAPIClient,
		RichMessageTemplate:     richMessageTemplate,
		OAuthClientID:           os.Getenv("CLIENT_ID"),
		OAuthClientSecret:       os.Getenv("CLIENT_SECRET"),
		OAuthClientRedirectURI:  os.Getenv("CLIENT_REDIRECT_URI"),
		IntegrationDataRequired: true}
	installationDataBytes, err := os.ReadFile("./installation_data.json")
	if err != nil {
		log.Print("Warning: installation data for the server not found. Ignore this if the integration has not been installed yet.")
	} else {
		err := server.SetClientHeadersFromInstallationData(installationDataBytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	http.HandleFunc("/install", server.InstallHandler)
	http.HandleFunc("/reply", server.ReplyHandler)
	log.Fatal(http.ListenAndServe(":5000", nil))
}
