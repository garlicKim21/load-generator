package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	cpuGeneratorURL = os.Getenv("CPU_GENERATOR_URL")
	allowedOrigins  = os.Getenv("ALLOWED_ORIGINS")
)

func init() {
	if cpuGeneratorURL == "" {
		cpuGeneratorURL = "http://localhost:8081"
	}
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
}

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
		}
	}

	r.Run(":8080")
}

func handleCPULoad(c *gin.Context) {
	action := c.Param("action")

	// CPU Generator 서비스 호출
	resp, err := http.Post(
		cpuGeneratorURL+"/load/"+action,
		"application/json",
		nil,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "Failed to communicate with CPU generator",
		})
		return
	}
	defer resp.Body.Close()

	// Generator의 응답 상태코드 확인
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{
			"status":  "error",
			"message": "CPU generator returned an error",
		})
		return
	}

	c.JSON(200, gin.H{
		"status": "success",
		"action": action,
	})
}
