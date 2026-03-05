package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	//"k8s.io/client-go/kubernetes"

	"kube-api/kube"
)

type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"image"`
	Image     string `json:"image"`
}

func GetServices(ctx context.Context) ([]Service, error) {
	clientset := kube.GetClient()
	deployments, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list deployments")
	}

	var services []Service
	for _, d := range deployments.Items {
		services = append(services, Service{
			Name:      d.Name,
			Namespace: d.Namespace,
			Image:     d.Spec.Template.Spec.Containers[0].Image,
		})
	}

	return services, nil
}

// GetK8sServices 获取所有Kubernetes服务列表，包括IP、端口等信息
func GetK8sServices(ctx context.Context) ([]K8sService, error) {
	clientset := kube.GetClient()
	
	// 获取所有命名空间的服务
	services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list services")
	}

	var k8sServices []K8sService
	for _, svc := range services.Items {
		k8sService := convertServiceToK8sService(&svc)
		k8sServices = append(k8sServices, k8sService)
	}

	return k8sServices, nil
}

// convertServiceToK8sService 将Kubernetes Service转换为K8sService
func convertServiceToK8sService(svc *corev1.Service) K8sService {
	// 转换端口信息
	var ports []ServicePort
	for _, p := range svc.Spec.Ports {
		// 获取TargetPort，如果是string类型则尝试使用Port作为TargetPort
		targetPort := p.TargetPort.IntVal
		if targetPort == 0 && p.TargetPort.StrVal != "" {
			// 如果TargetPort是字符串（如端口名），则使用Port作为fallback
			targetPort = p.Port
		}
		if targetPort == 0 {
			targetPort = p.Port
		}

		portInfo := ServicePort{
			Name:       p.Name,
			Protocol:   string(p.Protocol),
			Port:       p.Port,          // 外部端口（Service Port）
			TargetPort: targetPort,      // 内部端口（Pod Port）
			NodePort:   p.NodePort,      // 节点端口（仅NodePort和LoadBalancer类型）
		}
		ports = append(ports, portInfo)
	}

	// Get LoadBalancer IP if available
	var loadBalancerIP string
	if svc.Status.LoadBalancer.Ingress != nil && len(svc.Status.LoadBalancer.Ingress) > 0 {
		if svc.Status.LoadBalancer.Ingress[0].IP != "" {
			loadBalancerIP = svc.Status.LoadBalancer.Ingress[0].IP
		} else if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			loadBalancerIP = svc.Status.LoadBalancer.Ingress[0].Hostname
		}
	}

	return K8sService{
		Name:           svc.Name,
		Namespace:      svc.Namespace,
		Type:           string(svc.Spec.Type),
		ClusterIP:      svc.Spec.ClusterIP,
		Ports:          ports,
		ExternalIPs:    svc.Spec.ExternalIPs,
		LoadBalancerIP: loadBalancerIP,
	}
}

func RestartService(ctx context.Context, namespace, name string) error {
	clientset := kube.GetClient()
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get deployment")
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update deployment")
	}

	return nil
}

func DeleteService(ctx context.Context, namespace, name string) error {
	clientset := kube.GetClient()
	err := clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete deployment")
	}

	serviceName := fmt.Sprintf("%s-service", name)
	err = clientset.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to delete service")
	}

	return nil
}
