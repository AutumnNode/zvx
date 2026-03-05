package service

import (
	"context"
	"fmt"
	"kube-api/kube"
	"kube-api/pkg/logger"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NamespaceInfo represents basic information about a namespace.
type NamespaceInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

// ListNamespaces returns a list of all namespaces with their status and age.
func ListNamespaces() ([]NamespaceInfo, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("ListNamespaces failed: %v", err)
		return nil, err
	}

	namespaceList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.LogError("Failed to list namespaces from Kubernetes API: %v", err)
		return nil, err
	}

	var namespaces []NamespaceInfo
	for _, ns := range namespaceList.Items {
		age := time.Since(ns.CreationTimestamp.Time).String()
		namespaces = append(namespaces, NamespaceInfo{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Age:    age,
		})
	}
	return namespaces, nil
}

// CreateNamespace creates a new namespace.
func CreateNamespace(name string) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("CreateNamespace failed: %v", err)
		return err
	}

	// Check if namespace already exists
	_, err := clientset.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err == nil {
		// Namespace already exists
		return fmt.Errorf("namespace '%s' already exists", name)
	}

	if !errors.IsNotFound(err) {
		// Some other error occurred
		logger.LogError("Failed to check if namespace '%s' exists: %v", name, err)
		return fmt.Errorf("failed to check if namespace exists: %v", err)
	}

	// Namespace doesn't exist, create it
	ns := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	if err != nil {
		logger.LogError("Failed to create namespace '%s': %v", name, err)
		return fmt.Errorf("failed to create namespace: %v", err)
	}
	
	logger.LogInfo("Successfully created namespace '%s'", name)
	return nil
}

// CreateNamespaceIfNotExist ensures a namespace exists, creating it if it does not.
func CreateNamespaceIfNotExist(clientset *kubernetes.Clientset, namespace string) error {
	_, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Namespace does not exist, create it
			ns := &apiv1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, createErr := clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
			if createErr != nil {
				logger.LogError("Failed to create namespace '%s' which did not exist: %v", namespace, createErr)
				return createErr
			}
			logger.LogInfo("Successfully created namespace '%s'", namespace)
			return nil
		}
		logger.LogError("Failed to get namespace '%s': %v", namespace, err)
		return err
	}
	// Namespace already exists, this is not an error
	logger.LogInfo("Namespace '%s' already exists, skipping creation", namespace)
	return nil
}

// DeleteNamespace deletes a namespace.
func DeleteNamespace(name string) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("DeleteNamespace failed: %v", err)
		return err
	}
	// Forcefully delete the namespace by setting GracePeriodSeconds to 0
	gracePeriodSeconds := int64(0)
	propagationPolicy := metav1.DeletePropagationBackground

	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
		PropagationPolicy:  &propagationPolicy,
	}

	err := clientset.CoreV1().Namespaces().Delete(context.Background(), name, deleteOptions)
	if err != nil && !errors.IsNotFound(err) {
		err = fmt.Errorf("failed to delete namespace %s: %v", name, err)
		logger.LogError(err.Error())
		return err
	}
	return nil // Success if deleted or not found
}
