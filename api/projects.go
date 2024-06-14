package api

import (
	"github.com/Filecoin-Titan/titan/api/types"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-workerd-api/core/dao"
	"github.com/gnasnik/titan-workerd-api/core/errors"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
)

func DeployProjectHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	var params model.Project

	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}

	params.UserID = username

	var (
		scheduler *Scheduler
		err       error
	)

	if params.AreaID == "" {
		scheduler, err = GetRandomSchedulerAPI()
	} else {
		scheduler, err = GetSchedulerByAreaId(params.AreaID)
	}

	if err != nil {
		c.JSON(http.StatusOK, respError(err))
		return
	}

	params.AreaID = scheduler.AreaId
	projectId := uuid.NewString()

	err = scheduler.Api.DeployProject(c.Request.Context(), &types.DeployProjectReq{
		UUID:      projectId,
		Name:      params.Name,
		BundleURL: params.BundleUrl,
		UserID:    params.UserID,
		Replicas:  params.Replicas,
		CPUCores:  int64(params.CpuCores),
		Memory:    params.Memory,
		AreaID:    params.Region,
		NodeIDs:   strings.Split(params.NodeIds, ","),
	})
	if err != nil {
		log.Errorf("api: failed to deploy project: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	params.ProjectID = projectId

	err = dao.AddProject(c.Request.Context(), &params)
	if err != nil {
		log.Errorf("add project: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetProjectsHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)

	page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	size, _ := strconv.ParseInt(c.Query("size"), 10, 64)
	option := dao.QueryOption{
		Page:     int(page),
		PageSize: int(size),
		UserID:   username,
	}

	total, projects, err := dao.GetProjectByUserId(c.Request.Context(), option)
	if err != nil {
		log.Errorf("failed to get projects: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	for _, project := range projects {
		scheduler, err := GetSchedulerByAreaId(project.AreaID)
		if err != nil {
			continue
		}

		projectInfo, err := scheduler.Api.GetProjectInfo(c.Request.Context(), project.ProjectID)
		if err != nil {
			log.Errorf("api GetProjectInfo: %v", err)
			continue
		}

		project.Status = projectInfo.State
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list":  projects,
		"total": total,
	}))
}

func GetProjectInfoHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	projectId := c.Query("project_id")

	project, err := dao.GetProjectById(c.Request.Context(), projectId)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrProjectNotExists))
		return
	}

	if project.UserID != username {
		c.JSON(http.StatusOK, respError(errors.ErrProjectNotExists))
		return
	}

	scheduler, err := GetSchedulerByAreaId(project.AreaID)
	if err != nil {
		c.JSON(http.StatusOK, respError(err))
		return
	}

	projectInfo, err := scheduler.Api.GetProjectInfo(c.Request.Context(), projectId)
	if err != nil {
		log.Errorf("api: failed to get project info: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	c.JSON(http.StatusOK, respJSON(projectInfo))
}

func UpdateProjectHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	var params model.Project

	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}

	project, err := dao.GetProjectById(c.Request.Context(), params.ProjectID)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrProjectNotExists))
		return
	}

	if project.UserID != username {
		c.JSON(http.StatusOK, respError(errors.ErrProjectNotExists))
		return
	}

	if params.Name == "" {
		params.Name = project.Name
	}

	if params.BundleUrl == "" {
		params.BundleUrl = project.BundleUrl
	}

	if params.Replicas == 0 {
		params.Replicas = project.Replicas
	}

	var (
		scheduler *Scheduler
	)

	if params.AreaID == "" {
		scheduler, err = GetRandomSchedulerAPI()
	} else {
		scheduler, err = GetSchedulerByAreaId(params.AreaID)
	}

	if err != nil {
		c.JSON(http.StatusOK, respError(err))
		return
	}

	err = scheduler.Api.UpdateProject(c.Request.Context(), &types.ProjectReq{
		UUID:      project.ProjectID,
		Name:      params.Name,
		BundleURL: params.BundleUrl,
		Replicas:  params.Replicas,
	})
	if err != nil {
		log.Errorf("api: failed to deploy project: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	err = dao.UpdateProject(c.Request.Context(), &params)
	if err != nil {
		log.Errorf("update project: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func DeleteProjectHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	projectId := c.Query("project_id")

	project, err := dao.GetProjectById(c.Request.Context(), projectId)
	if err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrProjectNotExists))
		return
	}

	if project.UserID != username {
		c.JSON(http.StatusOK, respError(errors.ErrProjectNotExists))
		return
	}

	var (
		scheduler *Scheduler
	)

	if project.AreaID == "" {
		scheduler, err = GetRandomSchedulerAPI()
	} else {
		scheduler, err = GetSchedulerByAreaId(project.AreaID)
	}

	if err != nil {
		c.JSON(http.StatusOK, respError(err))
		return
	}

	err = scheduler.Api.DeleteProject(c.Request.Context(), &types.ProjectReq{UUID: projectId})
	if err != nil {
		log.Errorf("api: failed to delete project: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	err = dao.DeleteProjectById(c.Request.Context(), projectId)
	if err != nil {
		log.Errorf("failed to delete project: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	c.JSON(http.StatusOK, respJSON(nil))
}

func GetRegionsHandler(c *gin.Context) {
	region := c.Query("region")
	schedulers := GlobalServer.GetSchedulers()

	type Region struct {
		AreaId string         `json:"area_id"`
		Region map[string]int `json:"region"`
	}

	var out []*Region
	for _, scheduler := range schedulers {
		regions, err := scheduler.Api.GetCurrentRegionInfos(c.Request.Context(), region)
		if err != nil {
			log.Errorf("api: GetCurrentRegionInfos: %v", err)
			continue
		}

		out = append(out, &Region{AreaId: scheduler.AreaId, Region: regions})
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list": out,
	}))
}

func GetNodesByRegionHandler(c *gin.Context) {
	areaId := c.Query("area_id")
	region := c.Query("region")

	scheduler, err := GetSchedulerByAreaId(areaId)
	if err != nil {
		c.JSON(http.StatusOK, respError(err))
		return
	}

	nodes, err := scheduler.Api.GetNodesFromRegion(c.Request.Context(), region)
	if err != nil {
		log.Errorf("api: GetCurrentRegionInfos: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"list": nodes,
	}))
}
