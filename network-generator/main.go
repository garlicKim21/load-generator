package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	mutex     sync.Mutex
	rdb       *redis.Client
	ctx       = context.Background()
	cancel    context.CancelFunc
	running   bool
	client    *http.Client
)

func main() {
	client = &http.Client{
		Timeout: 5 * time.Second,
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: "backend:6379",
	})

	subscriber := rdb.Subscribe(ctx, "load:network:channel")
	defer subscriber.Close()

	// 메시지 수신 대기
	for msg := range subscriber.Channel() {
		action := msg.Payload
		mutex.Lock()
		if action == "start" && !running {
			startNetworkLoad()
		} else if action == "stop" && running {
			stopNetworkLoad()
		}
		mutex.Unlock()
	}
}

func getIntensity() int {
	val, err := rdb.Get(ctx, "load:network:intensity").Result()
	if err != nil {
		return 1
	}
	intensity, err := strconv.Atoi(val)
	if err != nil || intensity < 1 {
		return 1
	}
	if intensity > 10 {
		return 10
	}
	return intensity
}

func startNetworkLoad() {
	if running {
		return
	}

	intensity := getIntensity()
	goroutines := intensity * 10

	var childCtx context.Context
	childCtx, cancel = context.WithCancel(ctx)
	running = true

	fmt.Printf("Starting network load with %d goroutines (intensity %d)\n", goroutines, intensity)

	for i := 0; i < goroutines; i++ {
		go floodHTTP(childCtx)
	}
}

func stopNetworkLoad() {
	if !running || cancel == nil {
		return
	}

	fmt.Println("Stopping network load")
	cancel()
	cancel = nil
	running = false
}

func floodHTTP(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			resp, err := client.Get("http://backend:8080/api/v1/status")
			if err == nil {
				resp.Body.Close()
			}
		}
	}
}
