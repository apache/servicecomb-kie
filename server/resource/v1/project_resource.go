package v1

import (
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service/kv"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/foundation/validator"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
	"net/http"
	"strconv"
)

type ProjectResource struct {
}

func (r *ProjectResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/project",
			ResourceFunc: r.List,
			FuncDesc:     "list projects",
			Parameters: []*restful.Parameters{
				DocQueryProjectParameters, DocQueryLimitParameters, DocQueryOffsetParameters,
			},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.ProjectResponse{},
					Headers: map[string]goRestful.Header{
						common.HeaderRevision: DocHeaderRevision,
					},
				}, {
					Code:    http.StatusNotModified,
					Message: "empty body",
				},
			},
			Produces: []string{goRestful.MIME_JSON},
		},
	}
}

//List response Projects list
func (r *ProjectResource) List(rctx *restful.Context) {
	var err error
	request := &model.ListProjectRequest{
		Project: rctx.ReadQueryParameter(common.PathParameterProject),
	}
	request.Domain = ReadDomain(rctx.Ctx)
	offsetStr := rctx.ReadQueryParameter(common.QueryParamOffset)
	limitStr := rctx.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	if limit == 0 {
		limit = 20
	}
	request.Offset = offset
	request.Limit = limit
	err = validator.Validate(request)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	returnProjectsData(rctx, request)
}

func returnProjectsData(rctx *restful.Context, request *model.ListProjectRequest) {
	rev, projects, queryErr := kv.ListProjects(rctx.Ctx, request)
	if queryErr != nil {
		WriteErrResponse(rctx, queryErr.Code, queryErr.Message)
		return
	}
	rctx.ReadResponseWriter().Header().Set(common.HeaderRevision, strconv.FormatInt(rev, 10))
	err := writeResponse(rctx, projects)
	rctx.ReadRestfulRequest().SetAttribute(common.RespBodyContextKey, projects.Data)
	if err != nil {
		openlog.Error(err.Error())
	}
}
