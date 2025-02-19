/*
 * @Author: Bin
 * @Date: 2023-03-05
 * @FilePath: /gpt-zmide-server/routers/routers.go
 */
package routers

import (
	"github.com/gin-gonic/gin"

	"gpt-zmide-server/controllers"
	"gpt-zmide-server/controllers/apis"
	"gpt-zmide-server/middleware"
)

func BuildRouter(r *gin.Engine) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()

	r.GET("/", new(controllers.Index).Index)
	r.GET("/install", middleware.InstallMiddleware(), new(controllers.Install).Index)
	r.POST("/install/config", middleware.InstallMiddleware(), new(controllers.Install).Config)

	r.GET("/admin", middleware.BasicAuth(), new(controllers.Admin).Index)
	r.GET("/admin/signout", new(controllers.Admin).SignOut)

	// r.GET("/test", new(controllers.InstallController).Test) // 测试路由

	api := r.Group("/api")
	{

		apisCtlApp := new(apis.Application)
		apisCtlOpen := new(apis.Open)
		apisCtlConfig := new(apis.Config)
		apisCtlChat := new(apis.Chat)

		notDefault := func(ctx *gin.Context) {
			apis.APIDefaultController.Fail(ctx, "404 route not found.")
		}

		api.GET("/", notDefault)
		api.Any("/:route/*no", notDefault)

		// 开放接口
		openApis := api.Group("/open", middleware.BasicAuthOpen())
		openApis.POST("/", apisCtlOpen.Index)
		openApis.POST("/query", apisCtlOpen.Query)
		openApis.POST("/chat", apisCtlOpen.Chat)
		openApis.POST("/chat/raw", apisCtlOpen.ChatRaw)

		adminApis := api.Group("/admin", middleware.BasicAuthAdmin())

		// 系统配置
		adminConfig := adminApis.Group("/config")
		adminConfig.POST("/update/password", apisCtlConfig.UpdatePassword)
		adminConfig.GET("/system/info", apisCtlConfig.SystemInfo)
		adminConfig.GET("/ping/openai", apisCtlConfig.PingOpenAI)
		adminConfig.GET("/system/config", apisCtlConfig.ConfigInfo)
		adminConfig.POST("/system/config", apisCtlConfig.ConfigInfoSave)
		adminConfig.GET("/system/log", apisCtlConfig.GetSystemLogs)

		// 后台管理应用接口
		adminApp := adminApis.Group("/application")
		adminApp.GET("/", apisCtlApp.Index)
		adminApp.POST("/create", apisCtlApp.Create)
		adminApp.POST("/:id/update", apisCtlApp.Update)
		adminApp.POST("/:id/apikey/reset", apisCtlApp.RestApiKey)

		// 后台管理应用接口
		adminChat := adminApis.Group("/chat")
		adminChat.GET("/", apisCtlChat.Index)
	}

	return r
}
