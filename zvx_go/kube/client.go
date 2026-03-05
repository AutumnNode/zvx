package kube

import (
	"os"
	"path/filepath"
	"strconv"

	"kube-api/pkg/logger"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	Clientset     *kubernetes.Clientset
	MetricsClient *metricsclient.Clientset
)

// GetKubeconfigPath 优先从 KUBECONFIG 环境变量读取路径，
// 其次检查当前目录下是否存在 config 文件，
// 最后回退到默认的 kubeconfig 路径。
func GetKubeconfigPath() string {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		if _, err := os.Stat(kubeconfig); err == nil {
			logger.LogInfo("从 KUBECONFIG 环境变量加载配置: %s", kubeconfig)
			return kubeconfig
		}
	}

	if _, err := os.Stat("config"); err == nil {
		logger.LogInfo("从当前目录加载配置文件: config")
		return "config"
	}

	defaultPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	logger.LogInfo("回退到默认 kubeconfig 路径: %s", defaultPath)
	return defaultPath
}

// 自动初始化全局 Clientset 和 MetricsClient
func init() {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		logger.LogInfo("无法加载 in-cluster 配置: %v。正在尝试使用 kubeconfig 文件...", err)
		kubeconfigPath := GetKubeconfigPath()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	if err != nil {
		logger.LogError("无法加载 Kubernetes 配置 (in-cluster 和 kubeconfig 都失败了): %v", err)
		return
	}
	if config == nil {
		logger.LogError("无法加载 Kubernetes 配置: config 为 nil，但没有返回错误")
		return
	}

	// [新增] 检查环境变量以跳过 TLS 验证
	// 警告：这在生产中是不安全的，仅用于解决 CA 证书问题
	insecureSkip, _ := strconv.ParseBool(os.Getenv("KUBE_INSECURE_SKIP_TLS_VERIFY"))
	if insecureSkip {
		logger.LogInfo("警告: KUBE_INSECURE_SKIP_TLS_VERIFY 设置为 true，将跳过 Kubernetes API Server 的 TLS 证书验证。")
		config.Insecure = true
		// 当 Insecure 为 true 时，必须清除 CAData
		config.CAData = nil
		config.CAFile = ""
	}

	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		logger.LogError("无法创建 Kubernetes 客户端: %v", err)
		return
	}

	MetricsClient, err = metricsclient.NewForConfig(config)
	if err != nil {
		logger.LogError("警告：无法创建 Metrics 客户端，Metrics Server 可能未安装或配置不正确: %v", err)
	}
}

// InitClient 显式返回 clientset 和 config（用于 SPDYExecutor 场景）
func InitClient() (*kubernetes.Clientset, *rest.Config) {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		logger.LogInfo("无法加载 in-cluster 配置: %v。正在尝试使用 kubeconfig 文件...", err)
		kubeconfigPath := GetKubeconfigPath()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}

	if err != nil {
		logger.LogError("无法加载 Kubernetes 配置 (in-cluster 和 kubeconfig 都失败了): %v", err)
		return nil, nil
	}
	if config == nil {
		logger.LogError("无法加载 Kubernetes 配置: config 为 nil，但没有返回错误")
		return nil, nil
	}

	// [新增] 同样在此处应用 TLS 跳过逻辑
	insecureSkip, _ := strconv.ParseBool(os.Getenv("KUBE_INSECURE_SKIP_TLS_VERIFY"))
	if insecureSkip {
		config.Insecure = true
		config.CAData = nil
		config.CAFile = ""
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.LogError("无法创建 Kubernetes 客户端: %v", err)
		return nil, nil
	}

	return clientset, config
}

// GetClient 返回全局 Clientset（方便 service 调用）
func GetClient() *kubernetes.Clientset {
	if Clientset == nil {
		logger.LogError("Kubernetes Clientset 未初始化")
		return nil
	}
	return Clientset
}

// GetMetricsClient 返回全局 MetricsClient
func GetMetricsClient() *metricsclient.Clientset {
	if MetricsClient == nil {
		logger.LogError("Metrics Clientset 未初始化")
		return nil
	}
	return MetricsClient
}
