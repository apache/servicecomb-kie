package rbac

import (
	"context"

	crbac "github.com/go-chassis/cari/rbac"
)

type Dao interface {
	GetRole(ctx context.Context, name string) (*crbac.Role, error)
	AccountExist(ctx context.Context, name string) (bool, error)
}
