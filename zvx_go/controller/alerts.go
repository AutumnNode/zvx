package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
)

type Alert struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
}

var (
	alerts      []Alert
	alertsMutex sync.Mutex
	alertID     int = 1
	clientset   *kubernetes.Clientset
)

func InitAlerts(cs *kubernetes.Clientset) {
	clientset = cs
}

func GetAlerts(c *gin.Context) {
	if clientset == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kubernetes client not initialized"})
		return
	}
	// ... 获取警报逻辑 ...
	c.JSON(http.StatusOK, alerts) // 返回 JSON
}

func ClearAllAlerts(c *gin.Context) {
	alertsMutex.Lock()
	defer alertsMutex.Unlock()
	alerts = []Alert{}
	c.JSON(http.StatusOK, gin.H{
		"message": "All alerts cleared",
	})
}
