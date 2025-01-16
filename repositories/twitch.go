package repositories

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/streampets/backend/models"
)

type userResponse struct {
	Data []struct {
		UserId   models.UserId `json:"id"`
		Username string        `json:"login"`
	} `json:"data"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type TwitchRepository struct {
	clientId     string
	clientSecret string
	accessToken  string
}

func NewTwitchRepository(id, secret string) (*TwitchRepository, error) {
	repo := &TwitchRepository{clientId: id, clientSecret: secret}
	if err := repo.refreshAccessToken(); err != nil {
		return repo, err
	}
	return repo, nil
}

func (repo *TwitchRepository) GetUsername(userId models.UserId) (string, error) {
	params := fmt.Sprintf("id=%s", userId)

	user, err := repo.getUserWithRefresh(params)
	if err != nil {
		return "", err
	}

	return user.Data[0].Username, nil
}

func (repo *TwitchRepository) GetUserId(username string) (models.UserId, error) {
	params := fmt.Sprintf("login=%s", username)

	user, err := repo.getUserWithRefresh(params)
	if err != nil {
		return "", err
	}

	return user.Data[0].UserId, nil
}

func (repo *TwitchRepository) getUserWithRefresh(params string) (userResponse, error) {
	resp, err := getUser(params, repo.accessToken, repo.clientId)
	if err != nil {
		return userResponse{}, err
	}

	if resp.StatusCode == 401 {
		if err := repo.refreshAccessToken(); err != nil {
			return userResponse{}, err
		}
		resp, err = getUser(params, repo.accessToken, repo.clientId)
		if err != nil {
			return userResponse{}, err
		}
	}

	if resp.StatusCode != 200 {
		return userResponse{}, errors.New("received incorrect status code from twitch")
	}

	var data userResponse
	if err := parseResponse(&data, resp); err != nil {
		return userResponse{}, err
	}

	return data, nil
}

func (repo *TwitchRepository) refreshAccessToken() error {
	resp, err := getAccessToken(repo.clientId, repo.clientSecret)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data accessTokenResponse
	if err := parseResponse(&data, resp); err != nil {
		return err
	}

	repo.accessToken = data.AccessToken
	return nil
}

func parseResponse(data interface{}, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}

// Could be in separate TwitchApi file
func getAccessToken(clientId, clientSecret string) (*http.Response, error) {
	return http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"grant_type":    {"client_credentials"},
	})
}

// Could be in separate TwitchApi file
func getUser(params, accessToken, clientId string) (*http.Response, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/users?%s", params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Client-Id", clientId)

	return http.DefaultClient.Do(req)
}
