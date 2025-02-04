package twitch

import (
	"fmt"
	"net/http"

	"github.com/streampets/backend/models"
)

// A struct used to communicate with the Twitch Api.
type TwitchApi struct {
	client *http.Client
}

// Creates a new TwitchApi client.
func New(client *http.Client) *TwitchApi {
	return &TwitchApi{client: client}
}

// Validates a Twitch user access token.
// Returns ErrInvalidAccessToken if the access token is not valid.
// Otherwise it returns the user id associated with the token.
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
		return "", ErrInvalidUserToken
	}

	var data validateResponse
	if err = parseResponse(&data, response); err != nil {
		return "", err
	}

	return data.UserId, nil
}
