package controller

import (
	"log"
	"net/http" // Keep http for status codes
	"strings"

	"github.com/gin-gonic/gin" // Import gin
	"kube-api/service"
)

// DeployImageHandler handles requests to deploy a Docker image.
func DeployImageHandler(c *gin.Context) {
	log.Println("Received request for image deployment")
	var req service.DeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error decoding deployment request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.DeployImage(c.Request.Context(), req); err != nil {
		log.Printf("Error deploying image: %v", err)
		// Differentiate between user input errors and internal server errors
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deployment initiated successfully"})
}
