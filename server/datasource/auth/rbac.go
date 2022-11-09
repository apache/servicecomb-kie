package auth

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/util"
	crbac "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"net/http"
	"os"

	etcd "github.com/apache/servicecomb-kie/server/datasource/etcd/rbac"
	mongo "github.com/apache/servicecomb-kie/server/datasource/mongo/rbac"
)

type DBAC_DB interface {
	GetRole(ctx context.Context, name string) (*crbac.Role, error)
	GenerateRBACRoleKey(name string) string
	AccountExist(ctx context.Context, name string) (bool, error)
	GenerateRBACAccountKey(name string) string
}

var dbacInstance DBAC_DB

func init() {
	d := os.Getenv("DBAC_DB")
	if d == "mongodb" {
		dbacInstance = &mongo.RBAC_Mongo{}
	} else {
		dbacInstance = &etcd.RBAC_ETCD{}
	}
}

func SetContext(req *http.Request) {
	v := req.Header.Get(restful.HeaderAuth)
	util.SetRequestContext(req, restful.HeaderAuth, v)
}
