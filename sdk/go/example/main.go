package main

import (
	"log"
	"net/http"

	"limiter.io/sdk/go"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the Rate Limiter SDK client
	// Point to our hosted rate limiter gateway and pass the developer's API key
	limiterClient := sdk.NewClient("http://localhost:8080", "replace_with_developer_api_key")

	r := gin.Default()

	// Use the SDK middleware to automatically rate limit this route group!
	protectedAPI := r.Group("/api")
	protectedAPI.Use(sdk.GinRateLimit(limiterClient))
	{
		protectedAPI.GET("/resource", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Welcome! You successfully passed rate limiting check.",
			})
		})
	}

	log.Println("Developer application running on :9000...")
	_ = r.Run(":9000")
}
