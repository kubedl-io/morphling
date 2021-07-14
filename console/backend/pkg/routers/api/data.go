package api

import (
	"fmt"
	"github.com/alibaba/morphling/console/backend/pkg/handlers"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
)

func NewDataAPIsController(dataHandler *handlers.DataHandler) *DataAPIsController {
	return &DataAPIsController{
		dataHandler: dataHandler,
	}
}

type DataAPIsController struct {
	dataHandler *handlers.DataHandler
}

func (ctrl *DataAPIsController) RegisterRoutes(routes *gin.RouterGroup) {
	data := routes.Group("/data")
	data.GET("/total", ctrl.getClusterTotal)
	data.GET("/request/:podPhase", ctrl.getClusterRequest)
	data.GET("/nodeInfos", ctrl.getClusterNodeInfos)
	data.GET("/namespaces", ctrl.getNamespaces)

}

func (ctrl *DataAPIsController) getClusterTotal(c *gin.Context) {
	clusterTotal, err := ctrl.dataHandler.GetClusterTotalResource()
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to getClusterTotal, err=%v", err))
		return
	}
	utils.Succeed(c, clusterTotal)
}

func (ctrl *DataAPIsController) getClusterRequest(c *gin.Context) {
	podPhase := c.Param("podPhase")
	if podPhase == "" {
		podPhase = string(corev1.PodRunning)
	}
	clusterRequest, err := ctrl.dataHandler.GetClusterRequestResource(podPhase)
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to getClusterRequest, err=%v", err))
		return
	}
	utils.Succeed(c, clusterRequest)
}

func (ctrl *DataAPIsController) getClusterNodeInfos(c *gin.Context) {
	nodeInfo, err := ctrl.dataHandler.GetNodesInfo()
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to getClusterNodeInfos, err=%v", err))
		return
	}
	utils.Succeed(c, nodeInfo)
}

func (ctrl *DataAPIsController) getNamespaces(c *gin.Context) {
	namespaces, err := ctrl.dataHandler.GetNamespaces()
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to getNamespaces, err=%v", err))
		return
	}
	utils.Succeed(c, namespaces)
}
