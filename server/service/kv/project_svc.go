package kv

import (
	"context"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/openlog"
)

func ListProjects(ctx context.Context, request *model.ListProjectRequest) (int64, *model.ProjectResponse, *errsvc.Error) {
	opts := []datasource.FindOption{
		datasource.WithOffset(request.Offset),
		datasource.WithLimit(request.Limit),
	}
	if len(request.Project) > 0 {
		opts = append(opts, datasource.WithProject(request.Project))
	}
	rev, err := datasource.GetBroker().GetRevisionDao().GetRevision(ctx, request.Domain)
	if err != nil {
		return rev, nil, config.NewError(config.ErrInternal, err.Error())
	}
	kv, err := listProjectsHandle(ctx, request.Domain, opts...)
	if err != nil {
		openlog.Error("common: " + err.Error())
		return rev, nil, config.NewError(config.ErrInternal, common.MsgDBError)
	}
	return rev, kv, nil
}

func listProjectsHandle(ctx context.Context, domain string, options ...datasource.FindOption) (*model.ProjectResponse, error) {
	listSema.Acquire()
	defer listSema.Release()
	return datasource.GetBroker().GetProjectDao().List(ctx, domain, options...)
}
