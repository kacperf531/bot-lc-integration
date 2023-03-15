package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	httpClient := &http.Client{}
	richMessageTemplate, _ := os.ReadFile("./rich_message.json")
	server := &BotServer{HttpClient: *httpClient,
		RichMessageTemplate: richMessageTemplate,
		RequestHeader: http.Header{
			"Content-type": {"Application/json"}},
		OAuthClientID:          os.Getenv("CLIENT_ID"),
		OAuthClientSecret:      os.Getenv("CLIENT_SECRET"),
		OAuthClientRedirectURI: os.Getenv("CLIENT_REDIRECT_URI")}
	log.Fatal(http.ListenAndServe(":5000", server))
}
