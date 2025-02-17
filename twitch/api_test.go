package twitch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationMiddleware(t *testing.T) {

	userId := "12345"
	validToken := "valid token"
	invalidToken := "invalid token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(Authorization) != fmt.Sprintf("OAuth %s", validToken) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fmt.Fprintf(w, `{"user_id":"%s"}`, userId)
	}))
	defer server.Close()

	twitch := New(&http.Client{}, server.URL)

	router := gin.Default()
	router.Use(twitch.AuthorizationMiddleware())
	router.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	t.Run("status ok when using valid authorization cookie", func(t *testing.T) {
		mock.SetUp(t)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{Name: Authorization, Value: validToken})
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("unauthorized status when no 'Authorization' cookie present", func(t *testing.T) {
		mock.SetUp(t)

		req := httptest.NewRequest("GET", "/protected", nil)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("unauthorized status when access token invalid", func(t *testing.T) {
		mock.SetUp(t)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{Name: Authorization, Value: invalidToken})
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}
