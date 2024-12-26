package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	cpuGeneratorURL    = "http://cpu-generator:8081"
	memoryGeneratorURL = "http://memory-generator:8081"
	allowedOrigins     = "*"
)

func main() {
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

	resp, err := http.Post(
		cpuGeneratorURL+"/load/"+action,
		"application/json",
		nil,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to communicate with CPU generator",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"status":  "error",
			"message": "CPU generator returned an error",
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

	resp, err := http.Post(
		memoryGeneratorURL+"/load/"+action,
		"application/json",
		nil,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to communicate with Memory generator",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"status":  "error",
			"message": "Memory generator return an error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"action": action,
	})
}
