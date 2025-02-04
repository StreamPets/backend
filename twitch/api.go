package twitch

import (
	"fmt"
	"net/http"
)

type TwitchApi struct {
	client *http.Client
}

func NewTwitchApi(client *http.Client) *TwitchApi {
	return &TwitchApi{client: client}
}

func (t *TwitchApi) ValidateToken(accessToken string) (ValidateTokenResponse, error) {
	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ValidateTokenResponse{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	response, err := t.client.Do(req)
	if err != nil {
		return ValidateTokenResponse{}, err
	}

	if response.StatusCode == 401 {
		return ValidateTokenResponse{}, ErrInvalidAccessToken
	}

	var data ValidateTokenResponse
	if err = parseResponse(&data, response); err != nil {
		return ValidateTokenResponse{}, err
	}

	return data, nil
}
