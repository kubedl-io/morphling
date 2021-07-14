package api

import (
	"encoding/json"
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/console/backend/pkg/handlers"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"strconv"
	"time"
)

func NewExperimentAPIsController(experimentHandler *handlers.ExperimentHandler) *ExperimentAPIsController {
	return &ExperimentAPIsController{
		experimentHandler: experimentHandler,
	}
}

type ExperimentAPIsController struct {
	experimentHandler *handlers.ExperimentHandler
}

func (ctrl *ExperimentAPIsController) RegisterRoutes(routes *gin.RouterGroup) {
	experiment := routes.Group("/experiment")
	experiment.GET("/list", ctrl.getExperimentList)
	experiment.GET("/detail", ctrl.getExperimentDetail)
	experiment.POST("/submitYaml", ctrl.submitYaml)
	experiment.POST("/submitPars", ctrl.submitPars)
	experiment.DELETE("/:namespace/:name", ctrl.deleteExperiment)

}

func (ctrl *ExperimentAPIsController) deleteExperiment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	klog.Infof("post /experiment/delete with parameters: namespace=%s, name=%s", namespace, name)
	err := ctrl.experimentHandler.DeleteJobFromBackend(namespace, name)
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to delete experiment, err: %s", err))
	} else {
		utils.Succeed(c, nil)
	}
}

func (ctrl *ExperimentAPIsController) submitYaml(c *gin.Context) {

	data, err := c.GetRawData()
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to get raw posted data from request"))
		return
	}
	if err = ctrl.experimentHandler.SubmitExperiment(data); err != nil {
		handleErr(c, fmt.Sprintf("failed to submit experiment, err: %s", err))
		return
	}
	utils.Succeed(c, nil)
}

func (ctrl *ExperimentAPIsController) submitPars(c *gin.Context) {

	data, err := c.GetRawData()
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to get raw posted data from request"))
		return
	}
	if err = ctrl.experimentHandler.SubmitExperimentPars(data); err != nil {
		handleErr(c, fmt.Sprintf("failed to submit experiment, err: %s", err))
		return
	}
	utils.Succeed(c, nil)
}

func (ctrl *ExperimentAPIsController) getExperimentList(c *gin.Context) {
	var (
		ns, name, status, curPageNum, curPageSize string
	)
	query := utils.Query{}

	if startTime := c.Query("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			handleErr(c, fmt.Sprintf("failed to parse start time[start_time=%s], err=%s", startTime, err))
			return
		}
		query.StartTime = t
	} else {
		handleErr(c, "start_time should not be empty")
		return
	}

	if endTime := c.Query("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			handleErr(c, fmt.Sprintf("failed to parse end time[end_time=%s], err=%s", endTime, err))
			return
		}
		query.EndTime = t
	} else {
		query.EndTime = time.Now()
	}

	if ns = c.Query("namespace"); ns != "" {
		query.Namespace = ns
	}
	if name = c.Query("name"); name != "" {
		query.Name = name
	}
	if status = c.Query("status"); status != "" {
		query.Status = morphlingv1alpha1.ProfilingConditionType(status)
	}

	if curPageNum = c.Query("current_page"); curPageNum != "" {
		pageNum, err := strconv.Atoi(curPageNum)
		if err != nil {
			handleErr(c, fmt.Sprintf("failed to parse url parameter[current_page=%s], err=%s", curPageNum, err))
			return
		}
		if query.Pagination == nil {
			query.Pagination = &utils.QueryPagination{}
		}
		query.Pagination.PageNum = pageNum
	}
	if curPageSize = c.Query("page_size"); curPageSize != "" {
		pageSize, err := strconv.Atoi(curPageSize)
		if err != nil {
			handleErr(c, fmt.Sprintf("failed to parse url parameter[page_size=%s], err=%s", curPageSize, err))
			return
		}
		if query.Pagination == nil {
			query.Pagination = &utils.QueryPagination{}
		}
		query.Pagination.PageSize = pageSize
	}

	klog.Infof("get /experiment/list with parameters: namespace=%s, name=%s, status=%s, pageNum=%s, pageSize=%s",
		ns, name, status, curPageNum, curPageSize)

	peInfos, err := ctrl.experimentHandler.GetExperimentList(&query) // will change the content, e.g., query.Pagination.Count = len(dmoJobs)

	if err != nil {
		handleErr(c, fmt.Sprintf("failed to list jobs from backend, err=%v", err))
		return
	}
	utils.Succeed(c, map[string]interface{}{
		"peInfos": peInfos,
		"total":   query.Pagination.Count,
	})
}


func (ctrl *ExperimentAPIsController) getExperimentDetail(c *gin.Context) {
	var (
		ns, name string
	)
	query := utils.Query{}

	if ns = c.Query("namespace"); ns != "" {
		query.Namespace = ns
	}
	if name = c.Query("name"); name != "" {
		query.Name = name
	}

	klog.Infof("get /experiment/detail with parameters: namespace=%s, name=%s",
		ns, name)

	peInfos, err := ctrl.experimentHandler.GetExperimentDetail(&query) // will change the content, e.g., query.Pagination.Count = len(dmoJobs)
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to list experiment detail from backend, err=%v", err))
		return
	}
	b, err := json.Marshal(peInfos)
	klog.Infof(string(b))

	utils.Succeed(c, map[string]interface{}{
		"peInfo": peInfos,
		"total":   1,
	})
}