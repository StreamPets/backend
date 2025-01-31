package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/streampets/backend/models"
)

var ErrInvalidAccessToken error = errors.New("invalid access token")

type ValidateTokenResponse struct {
	Login  string          `json:"login"`
	UserId models.TwitchId `json:"user_id"`
}

type TwitchApi struct {
	client *http.Client
}

func NewTwitchApi(client *http.Client) *TwitchApi {
	return &TwitchApi{client: client}
}

func (t *TwitchApi) ValidateToken(accessToken string) (*ValidateTokenResponse, error) {
	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	response, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == 401 {
		return nil, ErrInvalidAccessToken
	}

	var data ValidateTokenResponse
	if err = parseResponse(&data, response); err != nil {
		return nil, err
	}

	return &data, nil
}

func parseResponse(data interface{}, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}
