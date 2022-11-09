package rbac

import "github.com/apache/servicecomb-kie/pkg/model"

func FilterKVList(kvs []*model.KVDoc, labelsList []map[string]string) []*model.KVDoc {
	var permKVs []*model.KVDoc
	for _, kv := range kvs {
		for _, labels := range labelsList {
			if !matchOne(kv, labels) {
				continue
			}
			permKVs = append(permKVs, kv)
			break
		}
	}
	return permKVs
}

func matchOne(kv *model.KVDoc, labels map[string]string) bool {
	for lk, lv := range labels {
		if v, ok := kv.Labels[lk]; ok && v != lv {
			return false
		}
	}
	return true
}

func MatchLabelsList(kv *model.KVDoc, labelsList []map[string]string) bool {
	for _, labels := range labelsList {
		if !matchOne(kv, labels) {
			continue
		}
		return true
	}
	return false
}
