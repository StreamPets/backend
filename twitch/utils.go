package twitch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrInvalidAccessToken error = errors.New("invalid access token")

func parseResponse(data interface{}, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}
