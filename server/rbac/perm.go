package rbac

import (
	"github.com/go-chassis/go-chassis/v2/server/restful"
)

func GetPermLabels(rctx *restful.Context, targetResource *ResourceScope) ([]map[string]string, error) {
	account, err := GetRoleFromReq(rctx)
	if err != nil {
		return nil, err
	}

	err = accountExist(rctx.Ctx, account.Name)
	if err != nil {
		return nil, err
	}

	matchedLabels, err := checkPerm(account.Roles, targetResource, rctx)
	if err != nil {
		return nil, err
	}

	return matchedLabels, nil
}

// this method decouple business code and perm checks
func checkPerm(roleList []string, targetResource *ResourceScope, rctx *restful.Context) ([]map[string]string, error) {
	hasAdmin, normalRoles := filterRoles(roleList)
	if hasAdmin {
		return nil, nil
	}

	//TODO add project
	project := rctx.ReadQueryParameter(":project")
	return Allow(rctx.Ctx, project, normalRoles, targetResource)
}
