package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/nicklaw5/helix/v2"
)

var EVENT_PATH = map[string]string{
	helix.EventSubTypeChannelChatMessage:     "/wh/message",
	helix.EventSubTypeChannelFollow:          "/wh/follow",
	helix.EventSubTypeChannelSubscription:    "/wh/sub",
	helix.EventSubTypeChannelSubscriptionEnd: "/wh/sub-end",
}

var channels map[string]*TwitchChannel
var clientId string
var appAccessToken string

func init() {
	URI := os.Getenv("APP_URI")
	if clientId == "" {
		panic(fmt.Errorf("URI Path (APP_URI) is not set."))
	}
	clientId = os.Getenv("CLIENT_ID")
	if clientId == "" {
		panic(fmt.Errorf("Twitch App client ID (CLIENT_ID) is not set."))
	}
	secret := os.Getenv("CLIENT_SECRET")
	if secret == "" {
		panic(fmt.Errorf("Twitch app token (CLIENT_SECRET) is not set."))
	}
	var err error
	appAccessToken, err = getAccessToken(clientId, secret)
	if err != nil {
		panic(fmt.Errorf("Error occurred while obtaining App Access Token %s", err))
	}
	channels = make(map[string]*TwitchChannel)

	for k := range EVENT_PATH {
		EVENT_PATH[k] = URI + EVENT_PATH[k]
	}
}

func getAccessToken(clientID, clientSecret string) (string, error) {
	type accessTokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"grant_type":    {"client_credentials"},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data accessTokenResponse
	json.Unmarshal(body, &data)
	return data.AccessToken, nil
}

func CloseAllStreams() {
	for _, channel := range channels {
		channel.Close()
	}
}
