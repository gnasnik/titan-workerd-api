package api

import (
	"fmt"
	"github.com/Filecoin-Titan/titan/api/types"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/gnasnik/titan-workerd-api/core/dao"
	"github.com/gnasnik/titan-workerd-api/core/errors"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
	"github.com/gnasnik/titan-workerd-api/pkg/iptool"
	"github.com/gnasnik/titan-workerd-api/utils"
	"github.com/google/uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type DeployReq struct {
	Name       string `db:"name" json:"name"`
	AreaID     string `db:"area_id" json:"area_id"`
	Region     string `db:"region" json:"region"`
	BundleUrl  string `db:"bundle_url" json:"bundle_url"`
	Replicas   int64  `db:"replicas" json:"replicas"`
	CpuCores   int32  `db:"cpu_cores" json:"cpu_cores"`
	Memory     int64  `db:"memory" json:"memory"`
	Expiration string `db:"expiration" json:"expiration"`
	NodeIds    string `db:"node_ids" json:"node_ids"`
	Version    int64  `db:"version" json:"version"`
}

func DeployProjectHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	username := claims[identityKey].(string)
	var params DeployReq

	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}

	var (
		schedulers []*Scheduler
	)

	clientIP := iptool.GetClientIP(c.Request)

	if params.AreaID == "" && params.NodeIds == "" {
		s, err := GetNearestScheduler(c.Request.Context(), clientIP)
		if err != nil {
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}
		schedulers = append(schedulers, s)
	} else if params.NodeIds != "" {
		nodes := strings.Split(params.NodeIds, ",")
		for _, id := range nodes {
			s, err := GetSchedulerByNodeId(id)
			if err != nil {
				log.Errorf("get scheduler by node id: %s %v", id, err)
				continue
			}
			schedulers = append(schedulers, s)
		}

	} else if params.AreaID != "" {
		s, err := GetSchedulerByAreaId(params.AreaID)
		if err != nil {
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}
		schedulers = append(schedulers, s)
	} else {
		c.JSON(http.StatusOK, respError(errors.ErrInvalidParams))
		return
	}

	if len(schedulers) == 0 {
		c.JSON(http.StatusOK, respError(errors.ErrNoAvailableScheduler))
		return
	}

	projectId := uuid.NewString()
	expirationT, _ := time.Parse(time.DateTime, params.Expiration)
	if expirationT.IsZero() {
		expirationT = time.Now().AddDate(1, 0, 0)
	}

	var areaIds []string

	for _, scheduler := range schedulers {
		areaIds = append(areaIds, scheduler.AreaId)
		var nodeIds []string
		if params.NodeIds != "" {
			nodeIds = strings.Split(params.NodeIds, ",")
		}
		err := scheduler.Api.DeployProject(c.Request.Context(), &types.DeployProjectReq{
			UUID:      projectId,
			Name:      params.Name,
			BundleURL: params.BundleUrl,
			UserID:    username,
			Replicas:  params.Replicas,
			Requirement: types.ProjectRequirement{
				CPUCores: int64(params.CpuCores),
				Memory:   params.Memory,
				AreaID:   params.Region,
				NodeIDs:  nodeIds,
				Version:  params.Version,
			},
			Expiration: expirationT,
		})
		if err != nil {
			log.Errorf("api: failed to deploy project: %v", err)
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}
	}

	if params.AreaID == "" {
		params.AreaID = strings.Join(areaIds, ",")
	}

	err := dao.AddProject(c.Request.Context(), &model.Project{
		ProjectID:  projectId,
		UserID:     username,
		Name:       params.Name,
		AreaID:     params.AreaID,
		Region:     params.Region,
		BundleUrl:  params.BundleUrl,
		Replicas:   params.Replicas,
		CpuCores:   params.CpuCores,
		Memory:     params.Memory,
		Expiration: expirationT,
		NodeIds:    params.NodeIds,
		Version:    params.Version,
	})
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

		project.Name = projectInfo.Name
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

	var projectInfos []*types.ProjectInfo

	areaIds := strings.Split(project.AreaID, ",")
	for _, areaId := range areaIds {
		scheduler, err := GetSchedulerByAreaId(areaId)
		if err != nil {
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}

		projectInfo, err := scheduler.Api.GetProjectInfo(c.Request.Context(), projectId)
		if err != nil {
			log.Errorf("api: failed to get project info: %v", err)
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}

		projectInfos = append(projectInfos, projectInfo)
	}

	c.JSON(http.StatusOK, respJSON(projectInfos))
}

func GetProjectTunnelsHandler(c *gin.Context) {
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

	externalIP := utils.GetClientIP(c.Request)

	location, err := utils.GetLocationByIP(externalIP)
	if err != nil {
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	var tuns []*model.ProjectTunnel
	maybeBestScheduler, err := GetMaybeBestScheduler(GetAreaIdFromLocation(location))
	if err != nil {
		if err != errors.ErrNoAvailableScheduler {
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}

		scheduler, err := GetSchedulerByAreaId(project.AreaID)
		if err != nil {
			c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
			return
		}

		maybeBestScheduler = scheduler
	}

	tunRes, err := maybeBestScheduler.Api.GetTunserverURLFromUser(c.Request.Context(), &types.TunserverReq{
		IP:     externalIP,
		AreaID: GetAreaIdFromLocation(location),
	})

	if err != nil {
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	for _, tunnel := range []*types.TunserverRsp{tunRes} {
		tuns = append(tuns, &model.ProjectTunnel{
			ProjectId:   projectId,
			TunnelIndex: int64(len(tuns)),
			WsURL:       tunnel.URL,
			NodeID:      tunnel.NodeID,
		})
	}

	if err != nil {
		log.Errorf("api: GetTunserverURLFromUser: %v", err)
		c.JSON(http.StatusOK, respErrorWrapMessage(errors.ErrInternalServer, err.Error()))
		return
	}

	c.JSON(http.StatusOK, respJSON(JsonObject{
		"tunnels": tuns,
	}))
}

func GetAreaIdFromLocation(l *model.Location) string {
	return fmt.Sprintf("%s-%s-%s-%s", l.Continent, l.Country, l.Province, l.City)
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
