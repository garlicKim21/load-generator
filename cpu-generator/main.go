package main

import (
	"context"
	"os/exec"
	"sync"
	"syscall"

	"github.com/redis/go-redis/v9"
)

var (
	stressCmd *exec.Cmd
	mutex     sync.Mutex
	rdb       *redis.Client
	ctx       = context.Background()
)

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "backend:6379",
	})

	subscriber := rdb.Subscribe(ctx, "load:cpu:channel")
	defer subscriber.Close()

	// 메시지 수신 대기
	for msg := range subscriber.Channel() {
		action := msg.Payload
		mutex.Lock()
		if action == "start" && (stressCmd == nil || stressCmd.ProcessState != nil) {
			startCPULoad()
		} else if action == "stop" && stressCmd != nil {
			stopCPULoad()
		}
		mutex.Unlock()
	}
}

func startCPULoad() error {
	if stressCmd == nil || stressCmd.ProcessState != nil {
		stressCmd = exec.Command("stress-ng", "--cpu", "1")

		stressCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return stressCmd.Start()
	}
	return nil
}

func stopCPULoad() error {
	if stressCmd == nil || stressCmd.Process == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(stressCmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	stressCmd.Wait()
	stressCmd = nil
	return nil
}
