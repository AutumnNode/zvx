package controller

import (
	"kube-api/service"
	"net/http"

	"fmt"

	"github.com/gin-gonic/gin"
)

// 兼容接口：/pods/:name/logs?namespace=xxx
func GetPodLogsByQuery(c *gin.Context) {
	namespace := c.Query("namespace")
	pod := c.Param("name") // 路由参数必须和 router.go 对应

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace query parameter is required"})
		return
	}

	fmt.Println("Fetching logs for Pod:", pod, "Namespace:", namespace)

	// 构造日志请求
	req := service.PodLogRequest{
		Namespace: namespace,
		Pod:       pod,
		TailLines: 100, // 默认获取最近100行
	}

	logs, err := service.GetPodLogs(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
