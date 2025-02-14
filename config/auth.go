package config

import (
	"encoding/base64"

	"github.com/streampets/backend/auth"
	"github.com/streampets/backend/database"
)

func CreateAuthService(db *database.DB) *auth.AuthService {
	extensionSecret, err := base64.StdEncoding.DecodeString(mustGetEnv("EXTENSION_SECRET"))
	if err != nil {
		panic(err)
	}

	return auth.New(string(extensionSecret))
}
