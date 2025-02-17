package twitch

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// A struct used to communicate with the Twitch Api.
type TwitchApi struct {
	client  *http.Client
	baseUrl string
}

// Creates a new TwitchApi client.
func New(
	client *http.Client,
	baseUrl string,
) *TwitchApi {
	return &TwitchApi{
		client:  client,
		baseUrl: baseUrl,
	}
}

func (t *TwitchApi) AuthorizationMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		token, err := ctx.Cookie(Authorization)
		if err != nil {
			slog.Debug("no 'Authorization' cookie present")
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		}

		channelId, err := t.validateToken(ctx, token)
		if err == ErrInvalidUserToken {
			slog.Debug("invalid access token in header")
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		} else if err != nil {
			slog.Error("error when validating access token", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.Set(ChannelId, channelId)
		ctx.Next()
	}
}

// Validates a Twitch user access token.
// Returns ErrInvalidAccessToken if the access token is not valid.
// Otherwise it returns the Twitch user id associated with the token.
func (t *TwitchApi) validateToken(ctx context.Context, accessToken string) (UserId, error) {
	type validateResponse struct {
		UserId UserId `json:"user_id"`
	}

	url := t.baseUrl + "/oauth/validate"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
