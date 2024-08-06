package dao

import (
	"context"
	"fmt"
	"github.com/gnasnik/titan-workerd-api/core/generated/model"
)

func AddProject(ctx context.Context, project *model.Project) error {
	_, err := DB.NamedExecContext(ctx, fmt.Sprintf(`
		INSERT INTO project ( user_id, project_id, name, area_id, region, bundle_url, status, replicas, cpu_cores, memory, expiration, version, created_at, updated_at)
			VALUES (:user_id, :project_id, :name, :area_id, :region, :bundle_url, :status, :replicas, :cpu_cores, :memory, :expiration, :version, now(), now());`,
	), project)
	return err
}

func UpdateProject(ctx context.Context, project *model.Project) error {
	_, err := DB.ExecContext(ctx, `UPDATE project set name = ?, bundle_url = ?, replicas = ? WHERE project_id = ?`, project.Name, project.BundleUrl, project.Replicas, project.ProjectID)
	return err
}

func GetProjectById(ctx context.Context, projectId string) (*model.Project, error) {
	var out model.Project
	if err := DB.GetContext(ctx, &out, `SELECT * FROM project WHERE project_id = ?`, projectId); err != nil {
		return nil, err
	}
	return &out, nil
}

func DeleteProjectById(ctx context.Context, projectId string) error {
	_, err := DB.ExecContext(ctx, `DELETE FROM project WHERE project_id = ?`, projectId)
	return err
}

func GetProjectByUserId(ctx context.Context, option QueryOption) (int64, []*model.Project, error) {
	var total int64
	var out []*model.Project

	limit := option.PageSize
	offset := option.Page
	if option.PageSize <= 0 {
		limit = 50
	}
	if option.Page > 0 {
		offset = limit * (option.Page - 1)
	}

	err := DB.GetContext(ctx, &total, `SELECT count(*) FROM project`)
	if err != nil {
		return 0, nil, err
	}

	err = DB.SelectContext(ctx, &out, `SELECT * FROM project where user_id = ? order by created_at DESC LIMIT ? OFFSET ?`, option.UserID, limit, offset)
	if err != nil {
		return 0, nil, err
	}

	return total, out, err
}
