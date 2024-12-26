package main

import (
	"log"
	"net/http"
	"os/exec"
	"sync"
	"syscall"

	"github.com/gin-gonic/gin"
)

var (
	stressCmd *exec.Cmd
	mutex     sync.Mutex
)

type Response struct {
	Status string `json:"status"`
}

func main() {
	r := gin.Default()
	r.POST("/load/:action", handleLoad)

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}

func handleLoad(c *gin.Context) {
	action := c.Param("action")
	mutex.Lock()
	defer mutex.Unlock()

	switch action {
	case "start":
		if err := startMemoryLoad(); err != nil {
			c.JSON(http.StatusInternalServerError, Response{Status: "error"})
			return
		}
		c.JSON(http.StatusOK, Response{Status: "started"})

	case "stop":
		if err := stopMemoryLoad(); err != nil {
			c.JSON(http.StatusInternalServerError, Response{Status: "error"})
			return
		}
		c.JSON(http.StatusOK, Response{Status: "stopped"})

	default:
		c.JSON(http.StatusBadRequest, Response{Status: "invalid action"})
	}
}

func startMemoryLoad() error {
	if stressCmd == nil || stressCmd.ProcessState != nil {
		stressCmd = exec.Command("stress-ng", "--vm", "1", "--vm-bytes", "1G", "--vm-hang", "0")

		stressCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		return stressCmd.Start()
	}
	return nil
}

func stopMemoryLoad() error {
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
