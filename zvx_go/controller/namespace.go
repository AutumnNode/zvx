package controller

import (
	"net/http"

	"kube-api/service"

	"github.com/gin-gonic/gin"
)

type NamespaceInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

// ListNamespaces 获取所有命名空间
func ListNamespaces(c *gin.Context) {
	namespaces, err := service.ListNamespaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取命名空间列表失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, namespaces)
}

// CreateNamespace 创建新的命名空间
func CreateNamespace(c *gin.Context) {
	var reqBody struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}

	err := service.CreateNamespace(reqBody.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建命名空间失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "命名空间创建成功",
		"name":    reqBody.Name,
	})
}

// DeleteNamespace 删除命名空间
func DeleteNamespace(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "命名空间名称是必需的",
		})
		return
	}

	// 防止删除系统命名空间
	systemNamespaces := []string{"default", "kube-system", "kube-public", "kube-node-lease"}
	for _, sysNs := range systemNamespaces {
		if name == sysNs {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "不能删除系统命名空间: " + name,
			})
			return
		}
	}

	err := service.DeleteNamespace(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除命名空间失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "命名空间删除成功",
		"name":    name,
	})
}
