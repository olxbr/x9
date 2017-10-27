package pkg

import (
	"bytes"
	"fmt"
	"github.com/go-redis/redis"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func sendAlert(messageJson string) {
	URL := SLACK_BOT_URL
	if strings.HasPrefix(URL, "http") {
		client := http.Client{}
		req, _ := http.NewRequest("POST", URL, bytes.NewBufferString(messageJson))

		req.Header.Set("Content-Type", "application/json")
		client.Do(req)
		req.Body.Close()
		fmt.Printf("%v - [Alert sent]\n", time.Now())
	} else {
		fmt.Printf("%v [WARNING] - [Alert IGNORED, invalid or empty slack url]\n", time.Now())
	}
}

func alertSlack() {

	fmt.Printf("%v - [Check alerts]\n", time.Now())
	// connect redis
	rc := redis.NewClient(&redis.Options{
		Addr:     REDIS_SERVER,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	keys, _ := rc.ZRangeWithScores("alertas", 0, -1).Result()
	count := len(keys)

	if count > 0 {

		text := ""
		var qtd float64

		for _, k := range keys {
			text += k.Member.(string) + ": " + strconv.FormatFloat(k.Score, 'f', 0, 64) + "\n"
			qtd += k.Score

		}

		sendAlert("Warning: " + strconv.FormatFloat(qtd, 'f', 0, 64) + " instances terminated before " + TOLERANCE + " seconds\n" + text)
	}
}
