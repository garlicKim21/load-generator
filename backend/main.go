package main

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	allowedOrigins = "*"
)

var (
	rdb *redis.Client
	ctx = context.Background()
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	r := gin.Default()

	// CORS 설정
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{allowedOrigins},
		AllowMethods: []string{"POST", "GET", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	// API 라우트
	v1 := r.Group("/api/v1")
	{
		load := v1.Group("/load")
		{
			load.POST("/cpu/:action", handleCPULoad)
			load.POST("/memory/:action", handleMemoryLoad)
		}
	}

	r.Run(":8080")
}

func handleCPULoad(c *gin.Context) {
	action := c.Param("action")

	err := rdb.Publish(ctx, "load:cpu:channel", action).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update state",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"action": action,
	})
}

func handleMemoryLoad(c *gin.Context) {
	action := c.Param("action")

	err := rdb.Publish(ctx, "load:memory:channel", action).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update state",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"action": action,
	})
}
