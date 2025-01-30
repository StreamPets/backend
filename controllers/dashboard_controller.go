package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streampets/backend/models"
)

type validateResponse struct {
	Login  string          `json:"login"`
	UserId models.TwitchId `json:"user_id"`
}

var ErrUnauthorized error = errors.New("this user is not authorized")

func HandleLogin(ctx *gin.Context) {
	token, err := ctx.Cookie("Authorization")
	if err == http.ErrNoCookie {
		slog.Debug("no 'Authorization' cooker header present")
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	if err != nil {
		slog.Error("error", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	response, err := getValidate(token)
	if err == ErrUnauthorized {
		slog.Debug("invalid access token in header")
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	if err != nil {
		slog.Error("error", "err", err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func getValidate(accessToken string) (*validateResponse, error) {
	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == 401 {
		return nil, ErrUnauthorized
	}

	var data validateResponse
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
