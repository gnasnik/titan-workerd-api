package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-workerd-api/config"
)

func RegisterRouter(r *gin.Engine, cfg config.Config) {
	apiV1 := r.Group("/api/v1")
	authMiddleware, err := jwtGinMiddleware(cfg.SecretKey)
	if err != nil {
		log.Fatalf("jwt auth middleware: %v", err)
	}

	err = authMiddleware.MiddlewareInit()
	if err != nil {
		log.Fatalf("authMiddleware.MiddlewareInit: %v", err)
	}

	user := apiV1.Group("/user")
	user.POST("/login", authMiddleware.LoginHandler)
	user.POST("/logout", authMiddleware.LogoutHandler)
	user.GET("/refresh_token", authMiddleware.RefreshHandler)

	user.Use(authMiddleware.MiddlewareFunc())
	user.POST("/info", GetUserInfoHandler)

	project := apiV1.Group("project")
	project.Use(authMiddleware.MiddlewareFunc())
	project.POST("/create", DeployProjectHandler)
	project.GET("/info", GetProjectInfoHandler)
	project.GET("/list", GetProjectsHandler)
	project.POST("/delete", DeleteProjectHandler)
	project.POST("/update", UpdateProjectHandler)
	project.GET("/regions", GetRegionsHandler)
	project.GET("/region/nodes", GetNodesByRegionHandler)
}
