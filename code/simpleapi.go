package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/name", func(c *gin.Context) {

		result := os.Getenv("name")

		if result == "" {
			result = "not found env var name"
		}

		c.JSON(http.StatusOK, gin.H{
			"message": result,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
