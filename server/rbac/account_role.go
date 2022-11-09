package rbac

import (
	"context"
	"errors"
	"fmt"
	rbacdb "github.com/apache/servicecomb-kie/server/datasource/etcd/rbac"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/security/authr"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
	"strings"
)

const (
	RootName = "root"
)

var ErrNoRoles = errors.New("no role found in token")

func GetRoleFromReq(rctx *restful.Context) (*rbacmodel.Account, error) {
	v := rctx.ReadHeader(restful.HeaderAuth)
	if v == "" {
		return nil, rbacmodel.NewError(rbacmodel.ErrNoAuthHeader, "")
	}
	s := strings.Split(v, " ")
	if len(s) != 2 {
		return nil, rbacmodel.ErrInvalidHeader
	}
	to := s[1]

	claims, err := authr.Authenticate(rctx.Ctx, to)
	if err != nil {
		return nil, err
	}

	m, ok := claims.(map[string]interface{})
	if !ok {
		openlog.Error("claims convert failed", openlog.WithErr(rbacmodel.ErrConvert))
		return nil, rbacmodel.ErrConvert
	}
	account, err := rbacmodel.GetAccount(m)
	if err != nil {
		openlog.Error("get account from token failed", openlog.WithErr(err))
		return nil, err
	}
	if len(account.Roles) == 0 {
		openlog.Error("no role found in token")
		return nil, ErrNoRoles
	}
	return account, nil
}

func accountExist(ctx context.Context, user string) error {
	// if root should pass, cause of root initialization
	if user == RootName {
		return nil
	}
	exist, err := rbacdb.AccountExist(ctx, user)
	if err != nil {
		return err
	}
	if !exist {
		msg := fmt.Sprintf("account [%s] is deleted", user)
		return rbacmodel.NewError(rbacmodel.ErrTokenOwnedAccountDeleted, msg)
	}
	return nil
}

func filterRoles(roleList []string) (hasAdmin bool, normalRoles []string) {
	for _, r := range roleList {
		if r == rbacmodel.RoleAdmin {
			hasAdmin = true
			return
		}
		normalRoles = append(normalRoles, r)
	}
	return
}
