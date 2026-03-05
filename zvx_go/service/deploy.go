package service

import (
	"context"
	"fmt"
	"strings"

	"kube-api/kube"
	"kube-api/pkg/logger"

	"github.com/pkg/errors"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NodeUsage represents the current usage and capacity of a Kubernetes node.
type NodeUsage struct {
	Name        string  `json:"name"`
	CPU         float64 `json:"cpu"`         // Used CPU in cores
	CPUCapacity float64 `json:"cpuCapacity"` // Total CPU capacity in cores
	Memory      float64 `json:"memory"`      // Used Memory in GB
	MemoryTotal float64 `json:"memoryTotal"` // Total Memory capacity in GB
}

// Port represents a port mapping.
type Port struct {
	ContainerPort int32 `json:"containerPort"`
	NodePort      int32 `json:"nodePort,omitempty"`
}

// VolumeMount represents a volume mount configuration.
type VolumeMount struct {
	PvcName   string `json:"pvcName"`
	MountPath string `json:"mountPath"`
}

// HorizontalScaling represents horizontal scaling configuration.
type HorizontalScaling struct {
	MinReplicas int `json:"minReplicas"`
	MaxReplicas int `json:"maxReplicas"`
	TargetCPU   int `json:"targetCPU"`
}

// VerticalScaling represents vertical scaling configuration.
type VerticalScaling struct {
	MinCPU   int `json:"minCPU"`
	MaxCPU   int `json:"maxCPU"`
	MinMemory int `json:"minMemory"`
	MaxMemory int `json:"maxMemory"`
}

// DeploymentRequest defines the structure for an image deployment request.
type DeploymentRequest struct {
	ProjectName            string             `json:"projectName"`
	ImageName              string             `json:"imageName"`
	EnvironmentVariables   []string           `json:"environmentVariables"` // KEY=VALUE format
	CPU                    float64            `json:"cpu"`                  // CPU in cores
	Memory                 int                `json:"memory"`               // Memory in MB
	Ports                  []Port             `json:"ports"`                // Port mappings
	NodeName               string             `json:"nodeName,omitempty"`   // Optional: specific node to deploy to
	Namespace              string             `json:"namespace,omitempty"`  // Optional: namespace
	ScheduleOnControlNode  bool               `json:"scheduleOnControlNode,omitempty"`
	VolumeMounts           []VolumeMount      `json:"volumeMounts,omitempty"`
	HorizontalScaling     *HorizontalScaling `json:"horizontalScaling,omitempty"`
	VerticalScaling        *VerticalScaling  `json:"verticalScaling,omitempty"`
}

// DeployImage creates a Kubernetes Deployment and Service for the given image.
func DeployImage(ctx context.Context, req DeploymentRequest) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("DeployImage failed: %v", err)
		return err
	}

	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
	}

	if err := CreateNamespaceIfNotExist(clientset, namespace); err != nil {
		logger.LogError("Failed to ensure namespace '%s' exists: %v", namespace, err)
		return errors.Wrap(err, "failed to create namespace")
	}

	if req.ProjectName == "" || req.ImageName == "" || req.CPU <= 0 || req.Memory <= 0 || len(req.Ports) == 0 {
		return errors.New("projectName, imageName, cpu, memory, and at least one port are required")
	}
	for _, port := range req.Ports {
		if port.ContainerPort <= 0 {
			return errors.New("containerPort is required for all ports")
		}
	}

	cpuRequest := resource.NewMilliQuantity(int64(req.CPU*1000), resource.DecimalSI)
	memoryRequest := resource.NewQuantity(int64(req.Memory*1024*1024), resource.BinarySI)

	var envVars []apiv1.EnvVar
	for _, ev := range req.EnvironmentVariables {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 2 {
			envVars = append(envVars, apiv1.EnvVar{Name: parts[0], Value: parts[1]})
		} else {
			logger.LogInfo("Warning: Invalid environment variable format: %s. Skipping.", ev)
		}
	}

	deploymentName := strings.ToLower(fmt.Sprintf("%s-deployment", req.ProjectName))
	labels := map[string]string{
		"app":                            strings.ToLower(req.ProjectName),
		"app.kubernetes.io/managed-by": "zvx-dashboard",
	}

	// Set replica count based on horizontal scaling
	replicas := int32(1)
	if req.HorizontalScaling != nil {
		replicas = int32(req.HorizontalScaling.MinReplicas)
	}

	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace, Labels: labels},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  strings.ToLower(req.ProjectName),
							Image: req.ImageName,
							Ports: buildContainerPorts(req.Ports),
							Env:   envVars,
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{apiv1.ResourceCPU: *cpuRequest, apiv1.ResourceMemory: *memoryRequest},
								Limits:   apiv1.ResourceList{apiv1.ResourceCPU: *cpuRequest, apiv1.ResourceMemory: *memoryRequest},
							},
						},
					},
				},
			},
		},
	}

	if req.ScheduleOnControlNode {
		deployment.Spec.Template.Spec.Tolerations = []apiv1.Toleration{
			{Key: "node-role.kubernetes.io/control-plane", Operator: apiv1.TolerationOpExists, Effect: apiv1.TaintEffectNoSchedule},
			{Key: "node-role.kubernetes.io/master", Operator: apiv1.TolerationOpExists, Effect: apiv1.TaintEffectNoSchedule},
		}
	}

	if req.NodeName != "" {
		deployment.Spec.Template.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": req.NodeName}
	}

	if len(req.VolumeMounts) > 0 {
		for i, vm := range req.VolumeMounts {
			volumeName := fmt.Sprintf("storage-%d", i)
			deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, apiv1.Volume{
				Name: volumeName,
				VolumeSource: apiv1.VolumeSource{
					PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{ClaimName: vm.PvcName},
				},
			})
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, apiv1.VolumeMount{
				Name: volumeName, MountPath: vm.MountPath,
			})
		}
	}

	logger.LogInfo("Creating Deployment %s in namespace %s", deploymentName, namespace)
	_, err := clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		logger.LogError("Failed to create deployment '%s': %v", deploymentName, err)
		return errors.Wrap(err, "failed to create deployment")
	}

	// Create Horizontal Pod Autoscaler if horizontal scaling is enabled
	if req.HorizontalScaling != nil {
		hpaName := strings.ToLower(fmt.Sprintf("%s-hpa", req.ProjectName))
		hpa := &autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{Name: hpaName, Namespace: namespace},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       deploymentName,
				},
				MinReplicas: int32Ptr(int32(req.HorizontalScaling.MinReplicas)),
				MaxReplicas: int32(req.HorizontalScaling.MaxReplicas),
				Metrics: []autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: apiv1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: int32Ptr(int32(req.HorizontalScaling.TargetCPU)),
							},
						},
					},
				},
			},
		}

		logger.LogInfo("Creating HorizontalPodAutoscaler %s in namespace %s", hpaName, namespace)
		_, err = clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, hpa, metav1.CreateOptions{})
		if err != nil {
			logger.LogError("Failed to create HorizontalPodAutoscaler '%s': %v", hpaName, err)
			return errors.Wrap(err, "failed to create HorizontalPodAutoscaler")
		}
	}

	// Create Vertical Pod Autoscaler if vertical scaling is enabled
	if req.VerticalScaling != nil {
		// Note: Vertical Pod Autoscaler is not a standard Kubernetes resource yet.
		// For now, we'll log the configuration but won't create actual VPA resources.
		logger.LogInfo("Vertical scaling configured - MinCPU: %d, MaxCPU: %d, MinMemory: %d, MaxMemory: %d",
			req.VerticalScaling.MinCPU, req.VerticalScaling.MaxCPU, 
			req.VerticalScaling.MinMemory, req.VerticalScaling.MaxMemory)
	}

	serviceName := strings.ToLower(fmt.Sprintf("%s-service", req.ProjectName))
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: namespace, Labels: labels},
		Spec: apiv1.ServiceSpec{
			Selector: labels,
			Ports:    buildServicePorts(req.Ports),
			Type:     apiv1.ServiceTypeNodePort,
		},
	}

	logger.LogInfo("Creating Service %s in namespace %s", serviceName, namespace)
	_, err = clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		logger.LogError("Failed to create service '%s': %v", serviceName, err)
		return errors.Wrap(err, "failed to create service")
	}

	return nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func buildContainerPorts(ports []Port) []apiv1.ContainerPort {
	var containerPorts []apiv1.ContainerPort
	for _, p := range ports {
		containerPorts = append(containerPorts, apiv1.ContainerPort{ContainerPort: p.ContainerPort})
	}
	return containerPorts
}

func buildServicePorts(ports []Port) []apiv1.ServicePort {
	var servicePorts []apiv1.ServicePort
	for _, p := range ports {
		sp := apiv1.ServicePort{
			Protocol:   apiv1.ProtocolTCP,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromInt(int(p.ContainerPort)),
		}
		if p.NodePort > 0 {
			sp.NodePort = p.NodePort
		}
		servicePorts = append(servicePorts, sp)
	}
	return servicePorts
}
