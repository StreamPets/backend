package config

import (
	"fmt"
	"os"
)

func mustGetEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		panic(fmt.Errorf("%s not set", name))
	}
	return value
}
