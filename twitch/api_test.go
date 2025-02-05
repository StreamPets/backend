package twitch

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/streampets/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateToken(t *testing.T) {
	t.Run("user id retrieved when authorization token is valid", func(t *testing.T) {
		mockResponse := `{"user_id":"12345"}`
		expected := models.TwitchId("12345")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "OAuth valid token" {
				http.Error(w, "no authorization token", http.StatusUnauthorized)
				return
			}
			fmt.Fprintln(w, mockResponse)
		}))
		defer server.Close()

		client := &http.Client{}
		api := New(client, server.URL)

		ctx := context.Background()
		userId, err := api.ValidateToken(ctx, "valid token")

		assert.NoError(t, err)
		assert.Equal(t, expected, userId)
	})

	t.Run("invalid user token error when unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "no authorization token", http.StatusUnauthorized)
		}))
		defer server.Close()

		client := &http.Client{}
		api := New(client, server.URL)

		ctx := context.Background()
		_, err := api.ValidateToken(ctx, "valid token")

		if assert.Error(t, err) {
			assert.Equal(t, ErrInvalidUserToken, err)
		}
	})
}
