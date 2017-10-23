package cmd

import (
	"os"
)

func getOpt(name, df string) string {
	value := os.Getenv(name)
	if value == "" {
		value = df
	}
	return value
}
