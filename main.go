package main

import (
	"io/ioutil"
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
	richMessageTemplate, _ := ioutil.ReadFile("./rich_message.json")
	server := &BotServer{HttpClient: *httpClient,
		RichMessageTemplate: richMessageTemplate,
		RequestHeader: http.Header{
			"X-Author-Id":   {os.Getenv("BOT_ID")},
			"Authorization": {os.Getenv("TOKEN")},
			"Content-type":  {"Application/json"}},
		OAuthClientID:          os.Getenv("CLIENT_ID"),
		OAuthClientSecret:      os.Getenv("CLIENT_SECRET"),
		OAuthClientRedirectURI: os.Getenv("CLIENT_REDIRECT_URI")}
	log.Fatal(http.ListenAndServe(":5000", server))
}
