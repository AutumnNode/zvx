package controller

import (
	"net/http"
	"strconv"

	"kube-api/service"

	"github.com/gin-gonic/gin"
)

// 获取 Pod 历史日志，兼容 /logs/:namespace/:pod 和 /pods/:pod/logs?namespace=xxx
func GetPodLogsHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		namespace = c.Query("namespace")
	}
	pod := c.Param("pod")
	container := c.Query("container")

	tail := int64(200)
	if v := c.Query("tail"); v != "" {
		if t, err := strconv.ParseInt(v, 10, 64); err == nil {
			tail = t
		}
	}

	req := service.PodLogRequest{
		Namespace: namespace,
		Pod:       pod,
		Container: container,
		TailLines: tail,
	}

	logs, err := service.GetPodLogs(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespace": namespace,
		"pod":       pod,
		"container": container,
		"logs":      logs,
	})
}
