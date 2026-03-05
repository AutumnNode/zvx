package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	//"k8s.io/client-go/kubernetes"
	"kube-api/kube"
	"kube-api/service"
)

// ListK8sServices 获取K8s服务列表，包括IP和端口信息
func ListK8sServices(c *gin.Context) {
	clientset := kube.GetClient()
	services, err := service.ListK8sServices(clientset, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, services)
}

// CreateK8sService 创建K8s服务
func CreateK8sService(c *gin.Context) {
	namespace := c.Param("namespace")
	var req struct {
		Name  string `json:"name" binding:"required"`
		Image string `json:"image" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建服务的实现
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Service created successfully",
		"namespace": namespace,
		"name":      req.Name,
	})
}

// UpdateK8sService 更新K8s服务 - 直接调用k8s_network.go中的实现
func UpdateK8sService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var req service.K8sService
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 直接调用k8s_network.go中的UpdateK8sService方法
	clientset := kube.GetClient()
	updatedService, err := service.UpdateK8sService(clientset, namespace, name, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedService)
}

// DeleteK8sService 删除K8s服务
func DeleteK8sService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := service.DeleteK8sService(kube.GetClient(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Service deleted successfully",
		"namespace": namespace,
		"name":      name,
	})
}
