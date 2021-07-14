package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Redirect404(c *gin.Context) {
	c.Redirect(http.StatusFound, "/404")
}

func Redirect403(c *gin.Context) {
	c.Redirect(http.StatusFound, "/403")
}

func Redirect500(c *gin.Context) {
	c.Redirect(http.StatusFound, "/500")
}
