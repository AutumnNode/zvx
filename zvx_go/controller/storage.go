package controller

import (
	"net/http"
	"time"

	"kube-api/pkg/logger"
	"kube-api/service"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GET /storage/pv
func GetPersistentVolumes(c *gin.Context) {
	pvs, err := service.ListPersistentVolumes()
	if err != nil {
		logger.LogError("Failed to get persistent volumes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := []gin.H{}
	for _, pv := range pvs {
		capacity := ""
		if q, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
			capacity = q.String()
		}

		status := string(pv.Status.Phase)
		if pv.ObjectMeta.DeletionTimestamp != nil {
			status = "Terminating"
		}

		items = append(items, gin.H{
			"name":          pv.Name,
			"capacity":      capacity,
			"accessModes":   pv.Spec.AccessModes,
			"reclaimPolicy": string(pv.Spec.PersistentVolumeReclaimPolicy),
			"status":        status,
			"storageClass":  pv.Spec.StorageClassName,
			"createdAt":     pv.CreationTimestamp.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GET /storage/pv/available
func GetAvailablePersistentVolumes(c *gin.Context) {
	pvs, err := service.ListPersistentVolumes()
	if err != nil {
		logger.LogError("Failed to get available persistent volumes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	availablePvs := []gin.H{}
	for _, pv := range pvs {
		if pv.Status.Phase == v1.VolumeAvailable {
			capacity := ""
			if q, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
				capacity = q.String()
			}
			availablePvs = append(availablePvs, gin.H{
				"name":     pv.Name,
				"capacity": capacity,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": availablePvs})
}

// GET /storage/pv/:name
func GetPersistentVolume(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is empty"})
		return
	}

	pv, err := service.GetPersistentVolume(name)
	if err != nil {
		logger.LogError("Failed to get persistent volume '%s': %v", name, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pv)
}

// PvCreateRequest defines the structure for a PV creation request from the frontend
type PvCreateRequest struct {
	Name          string                          `json:"name"`
	Capacity      string                          `json:"capacity"`
	AccessModes   []v1.PersistentVolumeAccessMode `json:"accessModes"`
	ReclaimPolicy v1.PersistentVolumeReclaimPolicy `json:"reclaimPolicy"`
	StorageClass  string                          `json:"storageClass"`
	Type          string                          `json:"type"` // 'hostPath' or 'nfs'
	HostPath      struct {
		Path string `json:"path"`
	} `json:"hostPath"`
	NFS struct {
		Server string `json:"server"`
		Path   string `json:"path"`
	} `json:"nfs"`
}

// POST /storage/pv
func CreatePersistentVolume(c *gin.Context) {
	var req PvCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if req.Name == "" || req.Capacity == "" || len(req.AccessModes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, capacity, and accessModes are required"})
		return
	}

	pvSpec := v1.PersistentVolumeSpec{
		Capacity: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(req.Capacity),
		},
		AccessModes:                   req.AccessModes,
		PersistentVolumeReclaimPolicy: req.ReclaimPolicy,
		StorageClassName:              req.StorageClass,
	}

	switch req.Type {
	case "hostPath":
		if req.HostPath.Path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hostPath path is required"})
			return
		}
		pvSpec.PersistentVolumeSource = v1.PersistentVolumeSource{
			HostPath: &v1.HostPathVolumeSource{Path: req.HostPath.Path},
		}
	case "nfs":
		if req.NFS.Server == "" || req.NFS.Path == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "NFS server and path are required"})
			return
		}
		pvSpec.PersistentVolumeSource = v1.PersistentVolumeSource{
			NFS: &v1.NFSVolumeSource{
				Server: req.NFS.Server,
				Path:   req.NFS.Path,
			},
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid storage type specified"})
		return
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: req.Name},
		Spec:       pvSpec,
	}

	if pv.Spec.PersistentVolumeReclaimPolicy == "" {
		pv.Spec.PersistentVolumeReclaimPolicy = v1.PersistentVolumeReclaimRetain
	}

	createdPv, err := service.CreatePersistentVolume(pv)
	if err != nil {
		logger.LogError("Controller failed to create PersistentVolume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PersistentVolume: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdPv)
}

// PUT /storage/pv/:name
func UpdatePersistentVolume(c *gin.Context) {
	var pv v1.PersistentVolume
	if err := c.ShouldBindJSON(&pv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	updatedPv, err := service.UpdatePersistentVolume(&pv)
	if err != nil {
		logger.LogError("Controller failed to update PersistentVolume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update PersistentVolume: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedPv)
}

// DELETE /storage/pv/:name
func DeletePersistentVolume(c *gin.Context) {
	name := c.Param("name")

	err := service.DeletePersistentVolume(name)
	if err != nil {
		logger.LogError("Controller failed to delete PersistentVolume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PersistentVolume deleted successfully"})
}

// ForceDeletePersistentVolume handles the force deletion of a PersistentVolume.
func ForceDeletePersistentVolume(c *gin.Context) {
	name := c.Param("name")

	err := service.ForceDeletePersistentVolume(name)
	if err != nil {
		logger.LogError("Controller failed to force delete PersistentVolume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to force delete PersistentVolume: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PersistentVolume force deleted successfully"})
}
