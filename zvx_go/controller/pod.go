// package controller
//
// import (
//
//	"kube-api/service"
//	"net/http"
//
//	"github.com/gin-gonic/gin"
//
// )
//
//	func ListPods(c *gin.Context) {
//		ns := c.Query("namespace")
//		pods, err := service.ListPods(ns)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"pods": pods})
//	}
//
//	func DeletePod(c *gin.Context) {
//		ns := c.Param("namespace")
//		name := c.Param("name")
//		err := service.DeletePod(ns, name)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"message": "Pod deleted"})
//	}
//
//	func RestartPod(c *gin.Context) {
//		ns := c.Param("namespace")
//		name := c.Param("name")
//		err := service.RestartPod(ns, name)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"message": "Pod restarted"})
//	}
//
// // 新增：支持 query 参数方式的删除
//
//	func DeletePodByQuery(c *gin.Context) {
//		ns := c.Query("namespace")
//		name := c.Param("name")
//		if ns == "" || name == "" {
//			c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
//			return
//		}
//		err := service.DeletePod(ns, name)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"message": "Pod deleted (query)"})
//	}
//
// // 新增：支持 query 参数方式的重启
//
//	func RestartPodByQuery(c *gin.Context) {
//		ns := c.Query("namespace")
//		name := c.Param("name")
//		if ns == "" || name == "" {
//			c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
//			return
//		}
//		err := service.RestartPod(ns, name)
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"message": "Pod restarted (query)"})
//	}
package controller

import (
	"kube-api/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ListPods(c *gin.Context) {
	ns := c.Query("namespace")
	pods, err := service.ListPods(ns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pods": pods})
}

func ListAllPods(c *gin.Context) {
	pods, err := service.ListAllPods()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pods": pods})
}

func DeletePod(c *gin.Context) {
	ns := c.Param("namespace")
	name := c.Param("name")
	err := service.DeletePod(ns, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod deleted"})
}

func RestartPod(c *gin.Context) {
	ns := c.Param("namespace")
	name := c.Param("name")
	err := service.RestartPod(ns, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod restarted"})
}

// DeletePodByQuery handles Pod deletion using query parameters.
// Deletes a Pod by extracting namespace from query parameter and name from URL parameter.
// 删除Pod：支持通过query参数传递命名空间，URL参数传递Pod名称
func DeletePodByQuery(c *gin.Context) {
	ns := c.Query("namespace")
	name := c.Param("name")
	if ns == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
		return
	}
	err := service.DeletePod(ns, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod deleted (query)"})
}

// RestartPodByQuery handles Pod restart using query parameters.
// Restarts a Pod by extracting namespace from query parameter and name from URL parameter.
// 重启Pod：支持通过query参数传递命名空间，URL参数传递Pod名称
func RestartPodByQuery(c *gin.Context) {
	ns := c.Query("namespace")
	name := c.Param("name")
	if ns == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
		return
	}
	err := service.RestartPod(ns, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod restarted (query)"})
}

// ForceDeletePodByQuery handles force deletion of a Pod via query parameters.
// Forcefully deletes a Pod with optional volume deletion and immediate removal (grace period = 0).
// 强制删除Pod：支持通过query参数传递命名空间和删除卷选项
func ForceDeletePodByQuery(c *gin.Context) {
	ns := c.Query("namespace")
	name := c.Param("name")
	deleteVolumes := c.Query("deleteVolumes") == "true"

	if ns == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
		return
	}
	
	log.Printf("ForceDeletePodByQuery: namespace=%s, name=%s, deleteVolumes=%v", ns, name, deleteVolumes)
	
	err := service.ForceDeletePod(ns, name, deleteVolumes)
	if err != nil {
		log.Printf("ForceDeletePodByQuery error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pod force deleted (query)"})
}

// GetPod retrieves a single Pod's YAML format information.
// Returns Pod details in YAML format by converting Kubernetes resource object.
// 获取单个Pod的YAML格式信息：将Kubernetes资源对象转换为YAML返回
func GetPod(c *gin.Context) {
	ns := c.Query("namespace")
	name := c.Param("name")
	if ns == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace and name are required"})
		return
	}

	pod, err := service.GetPod(ns, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	yamlStr, err := service.PodToYAML(pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert to YAML"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"yaml": yamlStr})
}
