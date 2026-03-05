package controller

import (
	"net/http"
	"sync"

	"kube-api/kube"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TerminalAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

var (
	lastAuth TerminalAuthRequest
	authLock sync.Mutex
)

// ========== 终端认证 ==========
func TerminalAuth(c *gin.Context) {
	var req TerminalAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}
	if req.Port == 0 {
		req.Port = 22
	}

	authLock.Lock()
	lastAuth = req
	authLock.Unlock()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Credentials accepted"})
}

func GetLastAuth() TerminalAuthRequest {
	authLock.Lock()
	defer authLock.Unlock()
	return lastAuth
}

// ========== 自动获取 K8s 控制节点 IP ==========
func GetControlNodeIP(c *gin.Context) {
	clientset, _ := kube.InitClient()

	// 查找带有 control-plane 或 master 标签的节点
	labelSelector := "node-role.kubernetes.io/control-plane"
	nodes, err := clientset.CoreV1().Nodes().List(c, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil || len(nodes.Items) == 0 {
		// 尝试兼容老版本 master 标签
		labelSelector = "node-role.kubernetes.io/master"
		nodes, err = clientset.CoreV1().Nodes().List(c, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil || len(nodes.Items) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到控制节点"})
			return
		}
	}

	// 默认取第一个控制节点的 InternalIP
	node := nodes.Items[0]
	var ip string
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			ip = addr.Address
			break
		}
	}

	if ip == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "控制节点未找到有效 InternalIP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ip": ip})
}
