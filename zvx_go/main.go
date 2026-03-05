package main

import (
	"kube-api/router"
	"kube-api/kube"
	"kube-api/pkg/logger"
)

func main() {
	// 初始化日志系统
	logger.InitLogger()
	
	logger.LogInfo("启动Kubernetes API服务...")
	
	// 尝试初始化Kubernetes客户端
	logger.LogInfo("初始化Kubernetes客户端...")
	if kube.GetClient() == nil {
		logger.LogError("无法初始化Kubernetes客户端，请检查kubeconfig配置")
		return
	}
	logger.LogInfo("Kubernetes客户端初始化成功")
	
	r := router.InitRouter()
	logger.LogInfo("HTTP服务启动，监听端口: 8081")
	
	// 设置panic恢复
	defer func() {
		if r := recover(); r != nil {
			logger.LogPanic(r)
		}
	}()
	
	if err := r.Run(":8081"); err != nil {
		logger.LogError("HTTP服务启动失败: %v", err)
		return
	}
}
