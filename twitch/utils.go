package twitch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/streampets/backend/models"
)

var ErrInvalidAccessToken error = errors.New("invalid access token")

type ValidateTokenResponse struct {
	Login  string          `json:"login"`
	UserId models.TwitchId `json:"user_id"`
}

func parseResponse(data interface{}, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}
