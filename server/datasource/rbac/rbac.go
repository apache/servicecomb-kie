package rbac

import (
	"context"
	crbac "github.com/go-chassis/cari/rbac"
)

type DBAC_DB interface {
	GetRole(ctx context.Context, name string) (*crbac.Role, error)
	GenerateRBACRoleKey(name string) string
	AccountExist(ctx context.Context, name string) (bool, error)
	GenerateRBACAccountKey(name string) string
}
