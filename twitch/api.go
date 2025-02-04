package twitch

import (
	"fmt"
	"net/http"

	"github.com/streampets/backend/models"
)

type TwitchApi struct {
	client *http.Client
}

func NewTwitchApi(client *http.Client) *TwitchApi {
	return &TwitchApi{client: client}
}

func (t *TwitchApi) ValidateToken(accessToken string) (models.TwitchId, error) {
	type validateResponse struct {
		UserId models.TwitchId `json:"user_id"`
	}

	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	response, err := t.client.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode == 401 {
		return "", ErrInvalidAccessToken
	}

	var data validateResponse
	if err = parseResponse(&data, response); err != nil {
		return "", err
	}

	return data.UserId, nil
}
