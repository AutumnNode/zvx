package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"kube-api/kube"
	"kube-api/pkg/logger"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// PodBrief represents brief information about a Pod.
type PodBrief struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Node       string    `json:"node"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	Containers int       `json:"containers"`
}

// GetPodStatus determines the detailed status of a Pod.
func GetPodStatus(pod v1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil {
			return cs.State.Waiting.Reason
		}
		if cs.State.Terminated != nil {
			return cs.State.Terminated.Reason
		}
	}

	return string(pod.Status.Phase)
}

// ListPods lists pods in a specific namespace.
func ListPods(namespace string) ([]PodBrief, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("ListPods failed: %v", err)
		return nil, err
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.LogError("Failed to list pods in namespace '%s': %v", namespace, err)
		return nil, err
	}

	var result []PodBrief
	for _, pod := range podList.Items {
		brief := PodBrief{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Node:       pod.Spec.NodeName,
			Status:     GetPodStatus(pod),
			CreatedAt:  pod.CreationTimestamp.Time,
			Containers: len(pod.Spec.Containers),
		}
		result = append(result, brief)
	}

	return result, nil
}

// ListAllPods lists pods in all namespaces.
func ListAllPods() ([]PodBrief, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("ListAllPods failed: %v", err)
		return nil, err
	}

	podList, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.LogError("Failed to list pods in all namespaces: %v", err)
		return nil, err
	}

	var result []PodBrief
	for _, pod := range podList.Items {
		brief := PodBrief{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Node:       pod.Spec.NodeName,
			Status:     GetPodStatus(pod),
			CreatedAt:  pod.CreationTimestamp.Time,
			Containers: len(pod.Spec.Containers),
		}
		result = append(result, brief)
	}

	return result, nil
}

// DeletePod deletes a pod.
func DeletePod(namespace, name string) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("DeletePod failed: %v", err)
		return err
	}
	err := clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logger.LogError("Failed to delete pod '%s' in namespace '%s': %v", name, namespace, err)
	}
	return err
}

// ForceDeletePod finds the deployment managing the pod and deletes it, along with associated resources.
func ForceDeletePod(namespace, name string, deleteVolumes bool) error {
	logger.LogInfo("ForceDeletePod: Starting force delete for pod=%s in namespace=%s, deleteVolumes=%v", name, namespace, deleteVolumes)
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("ForceDeletePod failed: %v", err)
		return err
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		errMsg := fmt.Sprintf("failed to get pod %s: %v", name, err)
		logger.LogError("ForceDeletePod error: %s", errMsg)
		return fmt.Errorf(errMsg)
	}
	logger.LogInfo("ForceDeletePod: Successfully retrieved pod %s", name)

	// --- Delete Associated PVCs and PVs ---
	if deleteVolumes {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				pvcName := volume.PersistentVolumeClaim.ClaimName
				logger.LogInfo("Found PVC %s associated with pod %s", pvcName, name)

				pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), pvcName, metav1.GetOptions{})
				if err != nil {
					logger.LogError("Warning: failed to get PVC %s: %v", pvcName, err)
					continue
				}

				err = clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), pvcName, metav1.DeleteOptions{})
				if err != nil {
					logger.LogError("Warning: failed to delete PVC %s: %v", pvcName, err)
				} else {
					logger.LogInfo("Successfully deleted PVC %s", pvcName)
				}

				if pvc.Spec.VolumeName != "" {
					pvName := pvc.Spec.VolumeName
					logger.LogInfo("Found PV %s associated with PVC %s", pvName, pvcName)
					err = clientset.CoreV1().PersistentVolumes().Delete(context.TODO(), pvName, metav1.DeleteOptions{})
					if err != nil {
						logger.LogError("Warning: failed to delete PV %s: %v", pvName, err)
					} else {
						logger.LogInfo("Successfully deleted PV %s", pvName)
					}
				}
			}
		}
	}

	// --- Delete Deployment, ReplicaSet, and Service ---
	var replicaSetOwner *metav1.OwnerReference
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "ReplicaSet" {
			replicaSetOwner = &owner
			break
		}
	}

	if replicaSetOwner == nil {
		gracePeriodSeconds := int64(0)
		deleteOptions := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}
		return clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, deleteOptions)
	}

	rs, err := clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), replicaSetOwner.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get owner ReplicaSet %s: %w", replicaSetOwner.Name, err)
	}

	var deploymentOwner *metav1.OwnerReference
	for _, owner := range rs.OwnerReferences {
		if owner.Kind == "Deployment" {
			deploymentOwner = &owner
			break
		}
	}

	if deploymentOwner == nil {
		return clientset.AppsV1().ReplicaSets(namespace).Delete(context.TODO(), rs.Name, metav1.DeleteOptions{})
	}

	err = clientset.AppsV1().Deployments(namespace).Delete(context.TODO(), deploymentOwner.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete controlling Deployment %s: %w", deploymentOwner.Name, err)
	}

	// --- Delete Associated Service ---
	services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		logger.LogError("Warning: failed to list services in namespace %s: %v", namespace, err)
	} else {
		for _, service := range services.Items {
			selector := labels.Set(service.Spec.Selector).AsSelector()
			if !selector.Empty() && selector.Matches(labels.Set(pod.Labels)) {
				logger.LogInfo("Found associated service %s for pod %s", service.Name, name)
				err := clientset.CoreV1().Services(namespace).Delete(context.TODO(), service.Name, metav1.DeleteOptions{})
				if err != nil {
					logger.LogError("Warning: failed to delete associated service %s: %v", service.Name, err)
				} else {
					logger.LogInfo("Successfully deleted service %s", service.Name)
				}
			}
		}
	}

	logger.LogInfo("ForceDeletePod: Successfully completed force delete for pod=%s in namespace=%s", name, namespace)
	return nil
}

// RestartPod restarts a pod by deleting it.
func RestartPod(namespace, name string) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("RestartPod failed: %v", err)
		return err
	}

	_, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logger.LogError("Failed to get pod '%s' for restart: %v", name, err)
		return err
	}

	err = clientset.CoreV1().Pods(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		logger.LogError("Failed to delete pod '%s' for restart: %v", name, err)
		return err
	}

	logger.LogInfo("Pod %s in namespace %s has been deleted (restart simulated)", name, namespace)
	return nil
}

// GetPod retrieves a pod object.
func GetPod(namespace, name string) (*v1.Pod, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("GetPod failed: %v", err)
		return nil, err
	}
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		logger.LogError("Failed to get pod '%s' in namespace '%s': %v", name, namespace, err)
	}
	return pod, err
}

// PodToYAML converts a Pod object to its YAML representation.
var yamlSerializer = json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)

func PodToYAML(pod *v1.Pod) (string, error) {
	var buf strings.Builder
	err := yamlSerializer.Encode(pod, &buf)
	if err != nil {
		logger.LogError("Failed to serialize pod to YAML: %v", err)
		return "", err
	}
	return buf.String(), nil
}

// PodLogRequest defines the parameters for fetching pod logs.
type PodLogRequest struct {
	Namespace string
	Pod       string
	Container string
	TailLines int64
}

// GetPodLogs fetches logs for a specific pod container.
func GetPodLogs(req PodLogRequest) (string, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("GetPodLogs failed: %v", err)
		return "", err
	}

	podLogOptions := &v1.PodLogOptions{}
	if req.Container != "" {
		podLogOptions.Container = req.Container
	}
	if req.TailLines > 0 {
		podLogOptions.TailLines = &req.TailLines
	}

	podLogRequest := clientset.CoreV1().Pods(req.Namespace).GetLogs(req.Pod, podLogOptions)
	stream, err := podLogRequest.Stream(context.Background())
	if err != nil {
		logger.LogError("Failed to open log stream for pod '%s': %v", req.Pod, err)
		return "", fmt.Errorf("failed to open log stream: %v", err)
	}
	defer stream.Close()

	data, err := ioutil.ReadAll(stream)
	if err != nil {
		logger.LogError("Failed to read logs from stream for pod '%s': %v", req.Pod, err)
		return "", fmt.Errorf("failed to read logs: %v", err)
	}

	logStr := string(data)
	if logStr == "" {
		logStr = "(no logs available)"
	}

	return logStr, nil
}
