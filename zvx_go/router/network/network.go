package network

import (
	"github.com/gin-gonic/gin"
	"kube-api/controller"
)

func SetupNetworkRoutes(api *gin.RouterGroup) {
	// Service routes
	services := api.Group("/services")
	{
		services.GET("", controller.ListK8sServices)
		services.POST("/:namespace", controller.CreateK8sService)
		services.PUT("/:namespace/:name", controller.UpdateK8sService)
		services.DELETE("/:namespace/:name", controller.DeleteK8sService)
	}
}
