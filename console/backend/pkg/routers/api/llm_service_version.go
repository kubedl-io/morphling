package api

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/morphling/console/backend/pkg/handlers"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

func NewLLMServiceVersionAPIsController(llmServiceVersionHandler *handlers.LLMServiceVersionHandler) *LLMServiceVersionAPIsController {
	return &LLMServiceVersionAPIsController{
		llmServiceVersionHandler: llmServiceVersionHandler,
	}
}

type LLMServiceVersionAPIsController struct {
	llmServiceVersionHandler *handlers.LLMServiceVersionHandler
}

func (ctrl *LLMServiceVersionAPIsController) RegisterRoutes(routes *gin.RouterGroup) {
	llmServiceVersion := routes.Group("/llm-service-version")
	llmServiceVersion.POST("", ctrl.createLLMServiceVersion)
	llmServiceVersion.GET("", ctrl.getLLMServiceVersions)
}

func (ctrl *LLMServiceVersionAPIsController) createLLMServiceVersion(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to get raw posted data from request"))
		return
	}

	klog.Infof("")

	// Unmarshal the data to LLMServiceVersion and github repo info
	var llmServiceVersionRequest utils.LLMServiceVersionRequest
	err = json.Unmarshal(data, &llmServiceVersionRequest)
	if err != nil {
		handleErr(c, fmt.Sprintf("failed to unmarshal llmServiceVersionRequest in json format"))
		return
	}

	klog.Infof("Received LLMServiceVersionRequest: %+v", llmServiceVersionRequest)

	if err := ctrl.llmServiceVersionHandler.CreateLLMServiceVersion(&llmServiceVersionRequest); err != nil {
		handleErr(c, fmt.Sprintf("Failed to create LLM service version: %v", err))
		return
	}

	utils.Succeed(c, nil)
}

func (ctrl *LLMServiceVersionAPIsController) getLLMServiceVersions(c *gin.Context) {
	// todo

	utils.Succeed(c, nil)
}
