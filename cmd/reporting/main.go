package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()

	db := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	defer db.Close()

	r.GET("/message/list/:sender/:receiver", newMessagesHandler(db))

	log.Fatal(r.Run(":8082"))
}

type httpError struct {
	Error string `json:"error"`
}

type message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func newMessagesHandler(db *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sender := c.Param("sender")
		receiver := c.Param("receiver")

		vals, err := db.LRange(
			c.Request.Context(),
			fmt.Sprintf("%s:%s", sender, receiver),
			0,
			-1,
		).Result()

		if err != nil {
			c.JSON(http.StatusBadRequest, httpError{Error: err.Error()})
			return
		}

		var messages []message

		for _, m := range vals {
			var msg message

			err = json.Unmarshal([]byte(m), &msg)
			if err != nil {
				c.JSON(http.StatusBadRequest, httpError{Error: err.Error()})
				return
			}

			messages = append(messages, msg)
		}

		c.JSON(http.StatusOK, messages)
	}
}
