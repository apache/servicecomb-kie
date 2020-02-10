package v1

import (
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	"net/http"
)

//LabelResource is label API
type LabelResource struct {
}

//PutLabel update by label_id , only can modify alias
// create return 201 / update return 200
func (r *LabelResource) PutLabel(context *restful.Context) {
	var err error
	entity := new(model.LabelDoc)
	if err = readRequest(context, entity); err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	entity.Project = context.ReadPathParameter("project")
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	entity.Domain = domain.(string)
	res, err := service.LabelService.CreateOrUpdate(context.Ctx, entity)
	if err != nil {
		if err == service.ErrRevisionNotExist {
			WriteErrResponse(context, http.StatusNotFound, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	if res == nil {
		WriteErrResponse(context, http.StatusNotFound, "put alias fail", common.ContentTypeText)
		return
	}
	if entity.ID == "" {
		context.WriteHeader(http.StatusCreated)
	}
	err = writeResponse(context, res)
	if err != nil {
		openlogging.Error(err.Error())
	}
}

//URLPatterns defined config operations
func (r *LabelResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodPut,
			Path:         "/v1/{project}/kie/label",
			ResourceFunc: r.PutLabel,
			FuncDesc:     "put alias for label or create new label",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKeyID,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "update success",
				},
				{
					Code:    http.StatusCreated,
					Message: "create success",
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		},
	}
}
