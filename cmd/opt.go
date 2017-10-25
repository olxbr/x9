package cmd

import (
	"os"
)

func getOptEnv(name, df string) string {
	value := os.Getenv(name)
	if value == "" {
		value = df
	}
	return value
}
