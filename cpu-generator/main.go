package main

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	isLoading bool
	loadMutex sync.Mutex
	stopChan  chan bool
)

func main() {
	r := gin.Default()
	stopChan = make(chan bool)

	r.POST("/load/:action", handleLoad)
	r.Run(":8081")
}

func handleLoad(c *gin.Context) {
	action := c.Param("action")
	loadMutex.Lock()
	defer loadMutex.Unlock()

	switch action {
	case "start":
		if !isLoading {
			isLoading = true
			stopChan = make(chan bool)
			go generateLoad(stopChan)
			c.JSON(200, gin.H{"status": "started"})
		} else {
			c.JSON(400, gin.H{"status": "already running"})
		}
	case "stop":
		if isLoading {
			stopChan <- true
			isLoading = false
			c.JSON(200, gin.H{"status": "stopped"})
		} else {
			c.JSON(400, gin.H{"status": "not running"})
		}
	default:
		c.JSON(400, gin.H{"error": "invalid action"})
	}
}

func generateLoad(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		default:
			// CPU 부하 생성
			for i := 0; i < 1000000; i++ {
				_ = i * i
			}
		}
	}
}
