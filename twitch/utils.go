package twitch

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Indicates an invalid Twitch user access token.
var ErrInvalidUserToken error = errors.New("invalid access token")

// Unmarshals a response body into a specified struct.
//
//	var data dataStruct
//	if err = parseResponse(&data, response); err != nil {
//		return err
//	}
func parseResponse(data interface{}, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}
