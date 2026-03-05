package service

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// K8sService represents a Kubernetes Service with its essential details.
type K8sService struct {
	Name           string        `json:"name"`
	Namespace      string        `json:"namespace"`
	Type           string        `json:"type"`
	ClusterIP      string        `json:"clusterIP"`
	Ports          []ServicePort `json:"ports"`
	ExternalIPs    []string      `json:"externalIPs"`
	LoadBalancerIP string        `json:"loadBalancerIP"` // Added for LoadBalancer type services
}

// ServicePort represents a service port.
type ServicePort struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`         // External port (Service Port)
	TargetPort int32  `json:"targetPort"`   // Internal port (Pod Port)
	NodePort   int32  `json:"nodePort"`     // Node Port (for NodePort and LoadBalancer types)
}

// ListK8sServices retrieves a list of services from a specific namespace or all namespaces.
func ListK8sServices(clientset *kubernetes.Clientset, namespace string) ([]K8sService, error) {
	var services []K8sService
	serviceList, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	for _, item := range serviceList.Items {
		var ports []ServicePort
		for _, p := range item.Spec.Ports {
			// Handle TargetPort - could be int or string
			targetPort := p.TargetPort.IntVal
			if targetPort == 0 && p.TargetPort.StrVal != "" {
				// If TargetPort is a string (like port name), use Port as fallback
				targetPort = p.Port
			}
			if targetPort == 0 {
				targetPort = p.Port
			}

			ports = append(ports, ServicePort{
				Name:       p.Name,
				Protocol:   string(p.Protocol),
				Port:       p.Port,
				TargetPort: targetPort,
				NodePort:   p.NodePort,
			})
		}

		// Get LoadBalancer IP if available
		var loadBalancerIP string
		if item.Status.LoadBalancer.Ingress != nil && len(item.Status.LoadBalancer.Ingress) > 0 {
			if item.Status.LoadBalancer.Ingress[0].IP != "" {
				loadBalancerIP = item.Status.LoadBalancer.Ingress[0].IP
			} else if item.Status.LoadBalancer.Ingress[0].Hostname != "" {
				loadBalancerIP = item.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		services = append(services, K8sService{
			Name:           item.Name,
			Namespace:      item.Namespace,
			Type:           string(item.Spec.Type),
			ClusterIP:      item.Spec.ClusterIP,
			Ports:          ports,
			ExternalIPs:    item.Spec.ExternalIPs,
			LoadBalancerIP: loadBalancerIP,
		})
	}
	return services, nil
}

// CreateK8sService creates a new Kubernetes Service.
func CreateK8sService(clientset *kubernetes.Clientset, namespace string, svc *K8sService) (*K8sService, error) {
	// Placeholder implementation - service creation is not yet fully implemented
	return nil, fmt.Errorf("service creation is not yet fully implemented")
}

// UpdateK8sService updates an existing Kubernetes Service.
func UpdateK8sService(clientset *kubernetes.Clientset, namespace, name string, svc *K8sService) (*K8sService, error) {
	ctx := context.TODO()
	coreV1 := clientset.CoreV1()
	
	// Get the existing service
	existingService, err := coreV1.Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	
	// Update service type if provided
	if svc.Type != "" {
		switch svc.Type {
		case "ClusterIP", "NodePort", "LoadBalancer", "ExternalName":
			existingService.Spec.Type = corev1.ServiceType(svc.Type)
		}
	}
	
	// Update external IPs if provided
	if len(svc.ExternalIPs) > 0 {
		existingService.Spec.ExternalIPs = svc.ExternalIPs
	}
	
	// Update ports if provided
	if len(svc.Ports) > 0 {
		var updatedPorts []corev1.ServicePort
		for _, p := range svc.Ports {
			// Handle NodePort - only set if it's not zero and service type supports it
			nodePort := p.NodePort
			if nodePort == 0 {
				nodePort = existingService.Spec.Ports[0].NodePort // Keep existing NodePort if not specified
			}

			port := corev1.ServicePort{
				Name:       p.Name,
				Protocol:   corev1.Protocol(p.Protocol),
				Port:       p.Port,
				TargetPort: intstr.FromInt(int(p.TargetPort)),
				NodePort:   nodePort,
			}
			updatedPorts = append(updatedPorts, port)
		}
		existingService.Spec.Ports = updatedPorts
	}
	
	// Update the service
	updatedService, err := coreV1.Services(namespace).Update(ctx, existingService, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}
	
	// Convert back to K8sService format
	var ports []ServicePort
	for _, p := range updatedService.Spec.Ports {
		ports = append(ports, ServicePort{
			Name:       p.Name,
			Protocol:   string(p.Protocol),
			Port:       p.Port,
			TargetPort: p.TargetPort.IntVal,
			NodePort:   p.NodePort,
		})
	}

	// Get LoadBalancer IP if available
	var loadBalancerIP string
	if updatedService.Status.LoadBalancer.Ingress != nil && len(updatedService.Status.LoadBalancer.Ingress) > 0 {
		if updatedService.Status.LoadBalancer.Ingress[0].IP != "" {
			loadBalancerIP = updatedService.Status.LoadBalancer.Ingress[0].IP
		} else if updatedService.Status.LoadBalancer.Ingress[0].Hostname != "" {
			loadBalancerIP = updatedService.Status.LoadBalancer.Ingress[0].Hostname
		}
	}
	
	result := &K8sService{
		Name:           updatedService.Name,
		Namespace:      updatedService.Namespace,
		Type:           string(updatedService.Spec.Type),
		ClusterIP:      updatedService.Spec.ClusterIP,
		Ports:          ports,
		ExternalIPs:    updatedService.Spec.ExternalIPs,
		LoadBalancerIP: loadBalancerIP,
	}
	
	return result, nil
}

// DeleteK8sService deletes a Kubernetes Service.
func DeleteK8sService(clientset *kubernetes.Clientset, namespace, name string) error {
	ctx := context.TODO()
	propagationPolicy := metav1.DeletePropagationBackground
	
	err := clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	
	return nil
}
