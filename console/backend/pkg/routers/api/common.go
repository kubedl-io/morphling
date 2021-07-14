package api

import (
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
)

func handleErr(c *gin.Context, msg string) {
	formattedMsg := msg
	klog.Error(formattedMsg)
	utils.Failed(c, msg)
}
