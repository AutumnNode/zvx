package service

import (
	"context"
	"fmt"

	"kube-api/kube"
	"kube-api/pkg/logger"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPersistentVolumes lists all PersistentVolumes.
func ListPersistentVolumes() ([]v1.PersistentVolume, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("ListPersistentVolumes failed: %v", err)
		return nil, err
	}

	pvList, err := clientset.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logger.LogError("Failed to list persistent volumes: %v", err)
		return nil, err
	}
	return pvList.Items, nil
}

// GetPersistentVolume retrieves a single PV.
func GetPersistentVolume(name string) (*v1.PersistentVolume, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("GetPersistentVolume failed: %v", err)
		return nil, err
	}
	pv, err := clientset.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		logger.LogError("Failed to get persistent volume '%s': %v", name, err)
	}
	return pv, err
}

// UpdatePersistentVolume updates a PersistentVolume.
func UpdatePersistentVolume(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("UpdatePersistentVolume failed: %v", err)
		return nil, err
	}
	updatedPv, err := clientset.CoreV1().PersistentVolumes().Update(context.Background(), pv, metav1.UpdateOptions{})
	if err != nil {
		logger.LogError("Failed to update persistent volume '%s': %v", pv.Name, err)
	}
	return updatedPv, err
}

// CreatePersistentVolume creates a PersistentVolume.
func CreatePersistentVolume(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("CreatePersistentVolume failed: %v", err)
		return nil, err
	}
	createdPv, err := clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		logger.LogError("Failed to create persistent volume '%s': %v", pv.Name, err)
	}
	return createdPv, err
}

// DeletePersistentVolume deletes a PersistentVolume.
func DeletePersistentVolume(name string) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("DeletePersistentVolume failed: %v", err)
		return err
	}

	logger.LogInfo("Attempting to delete PersistentVolume: %s", name)
	err := clientset.CoreV1().PersistentVolumes().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		logger.LogError("Failed to delete PersistentVolume %s: %v", name, err)
	} else {
		logger.LogInfo("Successfully deleted PersistentVolume: %s", name)
	}
	return err
}

// ForceDeletePersistentVolume forcefully deletes a PersistentVolume by removing its finalizers.
func ForceDeletePersistentVolume(name string) error {
	clientset := kube.GetClient()
	if clientset == nil {
		err := fmt.Errorf("kubernetes clientset not initialized")
		logger.LogError("ForceDeletePersistentVolume failed: %v", err)
		return err
	}

	logger.LogInfo("Attempting to force delete PersistentVolume: %s", name)
	pv, err := GetPersistentVolume(name)
	if err != nil {
		logger.LogError("Failed to get PV %s for force deletion: %v", name, err)
		return err
	}

	// Remove finalizers
	pv.ObjectMeta.Finalizers = nil
	_, err = UpdatePersistentVolume(pv)
	if err != nil {
		logger.LogError("Failed to remove finalizers from PV %s: %v", name, err)
		return err
	}

	logger.LogInfo("Finalizers removed from PV %s. It should now be deleted.", name)
	return nil
}
