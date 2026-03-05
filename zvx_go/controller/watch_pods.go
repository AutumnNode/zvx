package controller

import (
	"encoding/json"
	"kube-api/kube"
	"kube-api/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WebSocket 用的 upgrader
var watchUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket 实现
func WatchPodsWS(c *gin.Context) {
	ns := c.Query("namespace")
	if ns == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	conn, err := watchUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	watcher, err := kube.Clientset.CoreV1().Pods(ns).Watch(c, metav1.ListOptions{})
	if err != nil {
		conn.WriteJSON(gin.H{"error": err.Error()})
		return
	}
	defer watcher.Stop()

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			podObj, ok := event.Object.(*v1.Pod)
			if !ok {
				continue // Or handle error
			}

			brief := service.PodBrief{
				Name:      podObj.Name,
				Namespace: podObj.Namespace,
				Node:      podObj.Spec.NodeName,
				Status:    service.GetPodStatus(*podObj),
				CreatedAt: podObj.CreationTimestamp.Time,
			}

			msg := gin.H{
				"type":   event.Type,
				"object": brief,
			}
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("WebSocket send error:", err)
				return
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}

// 原始 HTTP 流式接口
func WatchPods(c *gin.Context) {
	ns := c.Query("namespace")
	if ns == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	watcher, err := kube.Clientset.CoreV1().Pods(ns).Watch(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer watcher.Stop()

	// 设置为流式响应
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.WriteHeader(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(c.Writer)
	timeout := time.After(30 * time.Second) // 最多等待 30 秒关闭

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return
			}
			podObj := event.Object
			eventType := event.Type
			enc.Encode(gin.H{
				"type":   eventType,
				"object": podObj,
			})
			flusher.Flush()
		case <-timeout:
			return
		case <-c.Request.Context().Done():
			return
		}
	}
}
