package auth

import (
	"context"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
)

const verbGet, verbCreate, verbUpdate, verbDelete = "get", "create", "update", "delete"

func configPerms(verb string, labels map[string]string) *ResourceScope {
	var labelsList []map[string]string
	if labels != nil {
		labelsList = append(labelsList, labels)
	}
	return &ResourceScope{
		Type:   "config",
		Verb:   verb,
		Labels: labelsList,
	}
}

func FilterKVList(ctx context.Context, kvs []*model.KVDoc) ([]*model.KVDoc, error) {
	if !config.GetRBAC().Enabled {
		return kvs, nil
	}
	// TODO error
	labels, err := CheckPerm(ctx, configPerms(verbGet, nil))
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		// allow all
		return kvs, nil
	}
	return FilterKVs(kvs, labels), nil
}

func CheckGetKV(ctx context.Context, kv *model.KVDoc) error {
	if !config.GetRBAC().Enabled {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbGet, kv.Labels))
	return err
}

func CheckCreateKV(ctx context.Context, kv *model.KVDoc) error {
	if !config.GetRBAC().Enabled {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbCreate, kv.Labels))
	return err
}

func CheckDeleteKV(ctx context.Context, kv *model.KVDoc) error {
	if !config.GetRBAC().Enabled {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbDelete, kv.Labels))
	return err
}

func CheckUpdateKV(ctx context.Context, kv *model.KVDoc) error {
	if !config.GetRBAC().Enabled {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbUpdate, kv.Labels))
	return err
}
