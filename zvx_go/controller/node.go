// package controller
//
// import (
//
//	"kube-api/kube"
//	"net/http"
//
//	"github.com/gin-gonic/gin"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
// )
//
//	type NodeIPInfo struct {
//		Name string `json:"name"`
//		IP   string `json:"ip"`
//	}
//
//	func ListNodeIPs(c *gin.Context) {
//		client := kube.Clientset
//
//		nodes, err := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
//		if err != nil {
//			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//			return
//		}
//
//		var result []NodeIPInfo
//		for _, node := range nodes.Items {
//			var ip string
//			for _, addr := range node.Status.Addresses {
//				if addr.Type == "InternalIP" {
//					ip = addr.Address
//					break
//				}
//			}
//			result = append(result, NodeIPInfo{
//				Name: node.Name,
//				IP:   ip,
//			})
//		}
//
//		c.JSON(http.StatusOK, result)
//	}
package controller

import (
	"context"
	"fmt"
	"log" // 导入 log 包
	"kube-api/kube"
	"net/http"

	"github.com/gin-gonic/gin"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type NodeIPInfo struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

func ListNodeIPs(c *gin.Context) {
	client := kube.Clientset

	nodes, err := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []NodeIPInfo
	for _, node := range nodes.Items {
		var ip string
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				ip = addr.Address
				break
			}
		}
		result = append(result, NodeIPInfo{
			Name: node.Name,
			IP:   ip,
		})
	}

	c.JSON(http.StatusOK, result)
}

// ListNodes 获取节点列表（仅返回节点名称）
func ListNodes(c *gin.Context) {
	client := kube.Clientset

	nodes, err := client.CoreV1().Nodes().List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nodeNames []string
	for _, node := range nodes.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodeNames})
}

// =============== 以下为新增代码 ===============

type NodeUsage struct {
	Name        string  `json:"name"`
	CPU         float64 `json:"cpu"`         // 当前 CPU 使用量（单位：核）
	CPUCapacity float64 `json:"cpuCapacity"` // CPU 总量（单位：核）
	Memory      float64 `json:"memory"`      // 当前内存使用量（单位：GB）
	MemoryTotal float64 `json:"memoryTotal"` // 总内存容量（单位：GB）
}

// 获取节点使用率信息
func GetNodeUsageHandler(c *gin.Context) {
	clientset := kube.Clientset
	metricsClient := kube.MetricsClient

	if metricsClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Metrics Server 客户端未初始化。请检查 Metrics Server 是否已安装并正确配置。", "details": "MetricsClient is nil"})
		return
	}

	usage, err := getNodeUsage(clientset, metricsClient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "无法获取节点指标数据。",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, usage)
}

func getNodeUsage(clientset *kubernetes.Clientset, metricsClient *metricsclient.Clientset) ([]NodeUsage, error) {
	var result []NodeUsage

	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点列表失败: %v", err)
	}
	if len(nodeList.Items) == 0 {
		return nil, fmt.Errorf("节点列表为空，请检查 Kubernetes 集群状态。")
	}

	metricsList, err := metricsClient.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取节点指标失败，Metrics Server 可能未安装或未运行: %v", err)
	}
	if len(metricsList.Items) == 0 {
		return nil, fmt.Errorf("节点指标列表为空，请检查 Metrics Server 是否正常工作。")
	}

	metricsMap := make(map[string]v1core.ResourceList)
	for _, m := range metricsList.Items {
		metricsMap[m.Name] = m.Usage
	}

	for _, node := range nodeList.Items {
		metrics, ok := metricsMap[node.Name]
		if !ok {
			log.Printf("警告：节点 %s 未找到对应的指标数据。", node.Name)
			// 如果某个节点没有指标数据，我们仍然返回该节点，但使用量为0
			cpuCap := node.Status.Capacity[v1core.ResourceCPU]
			memCap := node.Status.Capacity[v1core.ResourceMemory]
			result = append(result, NodeUsage{
				Name:        node.Name,
				CPU:         0.0, // 使用量为0
				CPUCapacity: float64(cpuCap.MilliValue()) / 1000.0,
				Memory:      0.0, // 使用量为0
				MemoryTotal: float64(memCap.Value()) / (1024 * 1024 * 1024),
			})
			continue
		}

		cpuQuantity := metrics[v1core.ResourceCPU]
		memQuantity := metrics[v1core.ResourceMemory]

		cpuCap := node.Status.Capacity[v1core.ResourceCPU]
		memCap := node.Status.Capacity[v1core.ResourceMemory]

		result = append(result, NodeUsage{
			Name:        node.Name,
			CPU:         float64(cpuQuantity.MilliValue()) / 1000.0,
			CPUCapacity: float64(cpuCap.MilliValue()) / 1000.0,
			Memory:      float64(memQuantity.Value()) / (1024 * 1024 * 1024),
			MemoryTotal: float64(memCap.Value()) / (1024 * 1024 * 1024),
		})
	}

	return result, nil
}
