package router

import (
	"kube-api/controller"
	"kube-api/router/network"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// InitRouter 初始化路由
func InitRouter() *gin.Engine {
	r := gin.Default()
	r.Use(CORSMiddleware())

	api := r.Group("")
	{
		// =========================
		// Pod & Namespace
		// =========================
		api.GET("/pods/all", controller.ListAllPods)  // 获取所有命名空间的Pod
		api.GET("/pods", controller.ListPods)
		api.GET("/api/namespaces", controller.ListNamespaces)
		api.POST("/api/namespaces", controller.CreateNamespace)  // 添加创建命名空间的路由
		api.DELETE("/api/namespaces/:name", controller.DeleteNamespace) // 删除命名空间
		api.GET("/pods/:name", controller.GetPod)
		api.DELETE("/pods/:name", controller.DeletePodByQuery)
		api.DELETE("/pods/:name/force", controller.ForceDeletePodByQuery) // 添加强制删除路由
		api.POST("/pods/:name/restart", controller.RestartPodByQuery)
		api.DELETE("/pod/:namespace/:name", controller.DeletePod)
		api.POST("/pod/:namespace/:name/restart", controller.RestartPod)

		// =========================
		// Pod Logs
		// =========================
		api.GET("/logs/:namespace/:pod", controller.GetPodLogsHandler)
		api.GET("/pods/:name/logs", controller.GetPodLogsByQuery) // 兼容 query

		// =========================
		// Pod Shell & Watch
		// =========================
		api.GET("/pods/shell", controller.PodShellWS)
		api.GET("/pods/watch", controller.WatchPods)
		api.GET("/pods/watch/ws", controller.WatchPodsWS)

		// =========================
		// Node
		// =========================
		api.GET("/nodes/ips", controller.ListNodeIPs)
		api.GET("/api/usage", controller.GetNodeUsageHandler) // Updated to new handler
		api.POST("/api/deploy", controller.DeployImageHandler) // New deployment handler

		// =========================
		// Storage: PV / PVC
		// =========================
		storageGroup := api.Group("/storage")
		{
			storageGroup.GET("/pv", controller.GetPersistentVolumes)
			storageGroup.POST("/pv", controller.CreatePersistentVolume)
			storageGroup.GET("/pv/available", controller.GetAvailablePersistentVolumes)
			storageGroup.GET("/pv/:name", controller.GetPersistentVolume)
			storageGroup.PUT("/pv/:name", controller.UpdatePersistentVolume)
			storageGroup.DELETE("/pv/:name", controller.DeletePersistentVolume)
			storageGroup.DELETE("/pv/:name/force", controller.ForceDeletePersistentVolume) // Add force delete route
		}

		// =========================
		// Network
		// =========================
		api.GET("/network/watch", controller.WatchNetwork)
		network.SetupNetworkRoutes(api)
		api.GET("/network/portip", controller.ListPortIpRules)
		api.POST("/network/portip", controller.CreatePortIpRule)
		api.PUT("/network/portip/:id", controller.UpdatePortIpRule)
		api.DELETE("/network/portip/:id", controller.DeletePortIpRule)
		api.GET("/api/control-node-ip", controller.GetControlNodeIP)
		api.POST("/api/terminal-auth", controller.TerminalAuth)
		api.GET("/control-node-ip", controller.GetControlNodeIP)
		api.POST("/terminal-auth", controller.TerminalAuth)
		r.GET("/terminal", controller.TerminalWS)

		// =========================
		// Version Control
		// =========================
		// version_control.SetupVersionControlRoutes(api)

		// =========================
		// Alerts
		// =========================
		api.GET("/api/alerts", controller.GetAlerts)         // 获取警报
		api.DELETE("/api/alerts", controller.ClearAllAlerts) // 清空警报
	}

	return r
}
