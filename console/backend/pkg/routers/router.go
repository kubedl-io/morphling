package routers

import (
	clientmgr "github.com/alibaba/morphling/console/backend/pkg/client"
	"github.com/alibaba/morphling/console/backend/pkg/constant"
	"github.com/alibaba/morphling/console/backend/pkg/handlers"
	"github.com/alibaba/morphling/console/backend/pkg/routers/api"
	"github.com/alibaba/morphling/console/backend/pkg/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	//"go.mongodb.org/mongo-driver/x/mongo/driver/auth"
	"net/http"
	"os"
	"path"
	"strings"
)

func init() {

}

type APIController interface {
	RegisterRoutes(routes *gin.RouterGroup)
}

func InitRouter(cmgr *clientmgr.ClientMgr) *gin.Engine {
	r := gin.New()

	r.Use(
		gin.Logger(),
		gin.Recovery(),
		func(context *gin.Context) {
			context.Header("Cache-Control", "no-store,no-cache")
		},
	)
	// No route
	r.NoRoute(
		utils.Redirect500,
		utils.Redirect403,
		utils.Redirect404,
	)

	//Login session
	store := cookie.NewStore([]byte("secret"))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(
		sessions.Sessions("loginSession", store),
	)

	// Bind dist dir
	morphlingDir, err := os.Getwd()
	if err != nil {
		morphlingDir = "."
	}
	distDir := path.Join(morphlingDir, "dist")
	//distDir := path.Join(morphlingDir, "console/frontend/dist")
	r.LoadHTMLFiles(distDir + "/index.html")
	r.Use(static.Serve("/", static.LocalFile(distDir, true)))
	r.Use(func(context *gin.Context) {
		if context.Request.URL == nil || context.Request.URL.Path == "" ||
			!strings.HasPrefix(context.Request.URL.Path, constant.ApiV1Routes) { //Todo: why api/v1
			context.HTML(http.StatusOK, "index.html", gin.H{})
		}
	})

	// Create handlers

	experimentHandler := handlers.NewExperimentHandler(cmgr)
	dataHandler := handlers.NewDataHandler(cmgr)

	// Register api v1 customized routers.
	apiV1Routes := r.Group(constant.ApiV1Routes)
	apiControllers := defaultAPIs(dataHandler, experimentHandler)
	for _, ctrl := range apiControllers {
		ctrl.RegisterRoutes(apiV1Routes)
	}

	return r
}

func defaultAPIs(dataHandler *handlers.DataHandler, experimentHandler *handlers.ExperimentHandler) []APIController {
	return []APIController{
		api.NewDataAPIsController(dataHandler),
		api.NewExperimentAPIsController(experimentHandler),
	}
}
