package controller

import (
    "net/http"
    "kube-api/service"

    "github.com/gin-gonic/gin"
)


func ListPortIpRules(c *gin.Context) {
    rules := service.GetMockPortIpRules()
    c.JSON(http.StatusOK, gin.H{"rules": rules})
}

func CreatePortIpRule(c *gin.Context) {
    var rule service.PortIpRule
    if err := c.BindJSON(&rule); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }
    createdRule := service.CreateMockPortIpRule(rule)
    c.JSON(http.StatusOK, createdRule)
}

func UpdatePortIpRule(c *gin.Context) {
    id := c.Param("id")
    var rule service.PortIpRule
    if err := c.BindJSON(&rule); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }
    updatedRule := service.UpdateMockPortIpRule(id, rule)
    c.JSON(http.StatusOK, updatedRule)
}

func DeletePortIpRule(c *gin.Context) {
    id := c.Param("id")
    service.DeleteMockPortIpRule(id)
    c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
