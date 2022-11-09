package rbac

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
)

func CheckGetOneKV(ctx context.Context, kv *model.KVDoc) error {
	var labelsList []map[string]string
	if kv.Labels != nil {
		labelsList = append(labelsList, kv.Labels)
	}
	scop := &ResourceScope{
		Type:   "config",
		Verb:   "get",
		Labels: labelsList,
	}
	_, err := CheckPermByReq(ctx, scop)
	return err
}

func FilterKVList(ctx context.Context, kvs []*model.KVDoc) ([]*model.KVDoc, error) {
	scop := &ResourceScope{
		Type: "config",
		Verb: "get",
	}
	labels, err := CheckPermByReq(ctx, scop)
	if err != nil {
		return nil, err
	}

	return FilterKV(kvs, labels), err
}

func CheckCreateOneKV(ctx context.Context, kv *model.KVDoc) error {
	var labelsList []map[string]string
	if kv.Labels != nil {
		labelsList = append(labelsList, kv.Labels)
	}
	scop := &ResourceScope{
		Type:   "config",
		Verb:   "create",
		Labels: labelsList,
	}
	_, err := CheckPermByReq(ctx, scop)
	return err
}

func CheckDeleteOneKV(ctx context.Context, kv *model.KVDoc) error {
	var labelsList []map[string]string
	if kv.Labels != nil {
		labelsList = append(labelsList, kv.Labels)
	}
	scop := &ResourceScope{
		Type:   "config",
		Verb:   "delete",
		Labels: labelsList,
	}
	_, err := CheckPermByReq(ctx, scop)
	return err
}

func CheckUpdateOneKV(ctx context.Context, kv *model.KVDoc) error {
	var labelsList []map[string]string
	if kv.Labels != nil {
		labelsList = append(labelsList, kv.Labels)
	}
	scop := &ResourceScope{
		Type:   "config",
		Verb:   "update",
		Labels: labelsList,
	}
	_, err := CheckPermByReq(ctx, scop)
	return err
}
