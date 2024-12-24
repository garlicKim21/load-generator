package main

import (
	"sync"

	"math"
	"runtime"

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
	numCPU := runtime.NumCPU()
	wg := sync.WaitGroup{}

	// 각 CPU 코어당 goroutine 생성
	for i := 0; i < numCPU; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := make([]float64, 1000000)

			for {
				select {
				case <-stop:
					return
				default:
					// 메모리 접근과 복잡한 연산 조합
					for j := 0; j < len(data); j++ {
						data[j] = math.Sin(float64(j)) * math.Cos(float64(j))
						_ = math.Pow(data[j], 2.0)
					}
				}
			}
		}()
	}
	wg.Wait()
}
