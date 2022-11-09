package rbac

import (
	"context"
	"github.com/little-cui/etcdadpt"
)

func AccountExist(ctx context.Context, name string) (bool, error) {
	return etcdadpt.Exist(ctx, GenerateRBACAccountKey(name))
}

func GenerateRBACAccountKey(name string) string {
	return "/cse-sr/accounts/" + name
}
