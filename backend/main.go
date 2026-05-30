package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	namespace = "load-tester"
)

var (
	rdb       *redis.Client
	ctx       = context.Background()
	k8sClient *kubernetes.Clientset
	mcClient  *metricsv.Clientset
)

// loadTypes enumerates all supported load types.
var loadTypes = []string{"cpu", "memory", "network"}

// ---------- auth helpers ----------

func getAdminPassword() string {
	if pw := os.Getenv("ADMIN_PASSWORD"); pw != "" {
		return pw
	}
	return "changeme" // 운영 시 ADMIN_PASSWORD 환경변수로 반드시 덮어쓸 것
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	if req.Password != getAdminPassword() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Store token in Redis with 24-hour TTL.
	redisKey := fmt.Sprintf("auth:token:%s", token)
	if err := rdb.Set(ctx, redisKey, "admin", 24*time.Hour).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token, Role: "admin"})
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		redisKey := fmt.Sprintf("auth:token:%s", token)
		role, err := rdb.Get(ctx, redisKey).Result()
		if err != nil || role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("role", role)
		c.Next()
	}
}

// ---------- request / response types ----------

type IntensityRequest struct {
	Level int `json:"level" binding:"required,min=1,max=10"`
}

type LoadState struct {
	State     string `json:"state"`
	Intensity int    `json:"intensity"`
}

type StatusResponse struct {
	CPU     LoadState `json:"cpu"`
	Memory  LoadState `json:"memory"`
	Network LoadState `json:"network"`
}

type DeploymentMetrics struct {
	Name      string `json:"name"`
	PodCount  int    `json:"podCount"`
	CPUUsage  int64  `json:"cpuUsageMillicores"`
	MemUsage  int64  `json:"memoryUsageBytes"`
}

type HPAStatus struct {
	Name            string `json:"name"`
	CurrentReplicas int32  `json:"currentReplicas"`
	DesiredReplicas int32  `json:"desiredReplicas"`
	CurrentCPUPct   *int32 `json:"currentCpuUtilization"`
}

type MetricsResponse struct {
	Deployments []DeploymentMetrics `json:"deployments"`
	HPAs        []HPAStatus         `json:"hpas"`
}

// ---------- main ----------

func main() {
	// Redis client (sidecar on localhost).
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Kubernetes in-cluster client.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("WARNING: could not create in-cluster k8s config: %v (metrics endpoints will be unavailable)", err)
	} else {
		k8sClient, err = kubernetes.NewForConfig(cfg)
		if err != nil {
			log.Printf("WARNING: could not create k8s clientset: %v", err)
		}
		mcClient, err = metricsv.NewForConfig(cfg)
		if err != nil {
			log.Printf("WARNING: could not create metrics clientset: %v", err)
		}
	}

	r := gin.Default()

	// CORS — allow all origins.
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	v1 := r.Group("/api/v1")
	{
		// Public endpoints (no auth).
		v1.GET("/status", handleStatus)
		v1.GET("/metrics", handleMetrics)
		v1.GET("/stream", handleStream)
		v1.POST("/auth/login", handleLogin)
		v1.GET("/auth/validate", authRequired(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"valid": true, "role": "admin"})
		})

		// Protected endpoints (admin auth required).
		load := v1.Group("/load")
		load.Use(authRequired())
		{
			load.POST("/cpu/:action", handleLoad("cpu"))
			load.POST("/memory/:action", handleLoad("memory"))
			load.POST("/network/:action", handleLoad("network"))
			load.POST("/:type/intensity", handleIntensity)
		}
	}

	r.Run(":8080")
}

// ---------- load start/stop handler ----------

func handleLoad(loadType string) gin.HandlerFunc {
	channel := fmt.Sprintf("load:%s:channel", loadType)
	stateKey := fmt.Sprintf("load:%s:state", loadType)

	return func(c *gin.Context) {
		action := c.Param("action")

		if action != "start" && action != "stop" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "action must be 'start' or 'stop'",
			})
			return
		}

		// Publish to the appropriate Redis channel.
		if err := rdb.Publish(ctx, channel, action).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to publish action",
			})
			return
		}

		// Update state key in Redis.
		state := "inactive"
		if action == "start" {
			state = "active"
		}
		if err := rdb.Set(ctx, stateKey, state, 0).Err(); err != nil {
			log.Printf("WARNING: failed to set state key %s: %v", stateKey, err)
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"type":   loadType,
			"action": action,
		})
	}
}

// ---------- intensity handler ----------

func handleIntensity(c *gin.Context) {
	loadType := c.Param("type")

	// Validate load type.
	valid := false
	for _, t := range loadTypes {
		if t == loadType {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid load type; must be cpu, memory, or network",
		})
		return
	}

	var req IntensityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "level is required and must be between 1 and 10",
		})
		return
	}

	intensityKey := fmt.Sprintf("load:%s:intensity", loadType)
	if err := rdb.Set(ctx, intensityKey, req.Level, 0).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to set intensity",
		})
		return
	}

	// Also publish the intensity change so generators can react in real time.
	channel := fmt.Sprintf("load:%s:channel", loadType)
	msg := fmt.Sprintf("intensity:%d", req.Level)
	_ = rdb.Publish(ctx, channel, msg).Err()

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"type":   loadType,
		"level":  req.Level,
	})
}

// ---------- status handler ----------

func handleStatus(c *gin.Context) {
	resp := StatusResponse{}

	for _, t := range loadTypes {
		stateVal, _ := rdb.Get(ctx, fmt.Sprintf("load:%s:state", t)).Result()
		if stateVal == "" {
			stateVal = "inactive"
		}
		intensityVal, _ := rdb.Get(ctx, fmt.Sprintf("load:%s:intensity", t)).Result()
		intensity, _ := strconv.Atoi(intensityVal)
		if intensity == 0 {
			intensity = 5 // default
		}

		ls := LoadState{State: stateVal, Intensity: intensity}
		switch t {
		case "cpu":
			resp.CPU = ls
		case "memory":
			resp.Memory = ls
		case "network":
			resp.Network = ls
		}
	}

	c.JSON(http.StatusOK, resp)
}

// ---------- metrics handler ----------

func fetchMetrics() (*MetricsResponse, error) {
	if k8sClient == nil || mcClient == nil {
		return nil, fmt.Errorf("kubernetes client not initialised")
	}

	deploymentNames := []string{"cpu-generator", "memory-generator", "network-generator"}
	result := &MetricsResponse{}

	// --- Pod metrics per deployment ---
	podMetricsList, err := mcClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}

	for _, dName := range deploymentNames {
		dm := DeploymentMetrics{Name: dName}
		for _, pm := range podMetricsList.Items {
			// Match pods by label prefix (pod names start with deployment name).
			if len(pm.Name) >= len(dName) && pm.Name[:len(dName)] == dName {
				dm.PodCount++
				for _, c := range pm.Containers {
					dm.CPUUsage += c.Usage.Cpu().MilliValue()
					dm.MemUsage += c.Usage.Memory().Value()
				}
			}
		}
		result.Deployments = append(result.Deployments, dm)
	}

	// --- HPA status ---
	hpaList, err := k8sClient.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list HPAs: %w", err)
	}

	for _, hpa := range hpaList.Items {
		hs := HPAStatus{
			Name:            hpa.Name,
			CurrentReplicas: hpa.Status.CurrentReplicas,
			DesiredReplicas: hpa.Status.DesiredReplicas,
		}
		// Find the CPU metric among current metrics.
		for _, m := range hpa.Status.CurrentMetrics {
			if m.Resource != nil && m.Resource.Name == "cpu" && m.Resource.Current.AverageUtilization != nil {
				val := *m.Resource.Current.AverageUtilization
				hs.CurrentCPUPct = &val
			}
		}
		result.HPAs = append(result.HPAs, hs)
	}

	return result, nil
}

func handleMetrics(c *gin.Context) {
	metrics, err := fetchMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Transform to frontend format
	result := make(map[string]interface{})

	podCounts := make(map[string]map[string]interface{})
	for _, d := range metrics.Deployments {
		maxReplicas := int32(10)
		desiredReplicas := int32(d.PodCount)
		for _, h := range metrics.HPAs {
			if strings.HasPrefix(h.Name, d.Name) {
				desiredReplicas = h.DesiredReplicas
				hpaObj, hErr := k8sClient.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, h.Name, metav1.GetOptions{})
				if hErr == nil {
					maxReplicas = hpaObj.Spec.MaxReplicas
				}
			}
		}
		podCounts[d.Name] = map[string]interface{}{
			"current": d.PodCount,
			"desired": desiredReplicas,
			"max":     maxReplicas,
		}
	}
	result["podCounts"] = podCounts

	var totalCPU, totalMem int64
	for _, d := range metrics.Deployments {
		totalCPU += d.CPUUsage
		totalMem += d.MemUsage
	}
	cpuPct := float64(totalCPU) / 12000.0 * 100.0
	memPct := float64(totalMem) / (48.0 * 1024 * 1024 * 1024) * 100.0
	result["clusterMetrics"] = map[string]interface{}{
		"cpuUsagePercent":    cpuPct,
		"memoryUsagePercent": memPct,
	}

	c.JSON(http.StatusOK, result)
}

// ---------- SSE stream handler ----------

func handleStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	clientGone := c.Request.Context().Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Send an initial event immediately.
	sendSSEEvent(c)

	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			return false
		case <-ticker.C:
			sendSSEEvent(c)
			return true
		}
	})
}

func sendSSEEvent(c *gin.Context) {
	payload := make(map[string]interface{})

	// Always include status.
	status := StatusResponse{}
	for _, t := range loadTypes {
		stateVal, _ := rdb.Get(ctx, fmt.Sprintf("load:%s:state", t)).Result()
		if stateVal == "" {
			stateVal = "inactive"
		}
		intensityVal, _ := rdb.Get(ctx, fmt.Sprintf("load:%s:intensity", t)).Result()
		intensity, _ := strconv.Atoi(intensityVal)
		if intensity == 0 {
			intensity = 5
		}
		ls := LoadState{State: stateVal, Intensity: intensity}
		switch t {
		case "cpu":
			status.CPU = ls
		case "memory":
			status.Memory = ls
		case "network":
			status.Network = ls
		}
	}
	// Transform to frontend format: loadStates
	payload["loadStates"] = map[string]map[string]interface{}{
		"cpu":     {"active": status.CPU.State == "active", "intensity": status.CPU.Intensity},
		"memory":  {"active": status.Memory.State == "active", "intensity": status.Memory.Intensity},
		"network": {"active": status.Network.State == "active", "intensity": status.Network.Intensity},
	}

	// Include k8s metrics in frontend-expected format.
	if k8sClient != nil && mcClient != nil {
		m, err := fetchMetrics()
		if err == nil {
			// Transform to frontend format: podCounts
			podCounts := make(map[string]map[string]interface{})
			for _, d := range m.Deployments {
				maxReplicas := int32(10) // default
				desiredReplicas := int32(d.PodCount)
				for _, h := range m.HPAs {
					if strings.HasPrefix(h.Name, strings.TrimSuffix(d.Name, "-generator")) ||
						strings.HasPrefix(h.Name, d.Name) {
						desiredReplicas = h.DesiredReplicas
						// Get max from HPA spec
						hpaObj, hErr := k8sClient.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, h.Name, metav1.GetOptions{})
						if hErr == nil {
							maxReplicas = hpaObj.Spec.MaxReplicas
						}
					}
				}
				podCounts[d.Name] = map[string]interface{}{
					"current": d.PodCount,
					"desired": desiredReplicas,
					"max":     maxReplicas,
				}
			}
			payload["podCounts"] = podCounts

			// clusterMetrics: aggregate CPU and memory as percentage
			var totalCPU, totalMem int64
			for _, d := range m.Deployments {
				totalCPU += d.CPUUsage
				totalMem += d.MemUsage
			}
			// Rough percentages based on worker capacity (3 workers * 4 cores = 12000m CPU, 48GB mem)
			cpuPct := float64(totalCPU) / 12000.0 * 100.0
			memPct := float64(totalMem) / (48.0 * 1024 * 1024 * 1024) * 100.0
			payload["clusterMetrics"] = map[string]interface{}{
				"cpuUsagePercent":    cpuPct,
				"memoryUsagePercent": memPct,
			}
		}
	}

	data, _ := json.Marshal(payload)
	c.SSEvent("message", string(data))
	c.Writer.Flush()
}
