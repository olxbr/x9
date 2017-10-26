package pkg

import (
	"os"
)

var SLACK_BOT_URL string   // slack bot tokeninzed url
var REDIS_SERVER string    // redis server and port
var TOLERANCE string       // minimum time in seconds an instance must run
var ALERT_TIMEFRAME string // checks and alerts sleep time in seconds
var REGIONS string         // Regions to be chceckd
var SERVICE_PORT string    // web http server listen port

func Init() {

	SLACK_BOT_URL = getOptEnv("SLACK_BOT_URL", "error")
	REDIS_SERVER = getOptEnv("REDIS_SERVER", "localhost:6379")
	TOLERANCE = getOptEnv("TOLERANCE", "3000")
	ALERT_TIMEFRAME = getOptEnv("ALERT_TIMEFRAME", "1200")
	REGIONS = getOptEnv("REGIONS", "sa-east-1,us-east-1")
	SERVICE_PORT = getOptEnv("SERVICE_PORT", "6969")

	Web()
}

func getOptEnv(name, df string) string {
	value := os.Getenv(name)
	if value == "" {
		value = df
	}
	return value
}
