package service

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-kie/pkg/model"
)

//services
var (
	KVService      KV
	HistoryService History
	DBInit         Init
)

//db errors
var (
	ErrKeyNotExists     = errors.New("key with labels does not exits")
	ErrRevisionNotExist = errors.New("label revision not exist")
)

//KV provide api of KV entity
type KV interface {
	CreateOrUpdate(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error)
	Delete(kvID string, labelID string, domain, project string) error
	FindKV(ctx context.Context, domain, project string, options ...FindOption) ([]*model.KVResponse, error)
}

//History provide api of History entity
type History interface {
	GetHistoryByLabelID(ctx context.Context, labelID string) ([]*model.LabelRevisionDoc, error)
}

//Init init db session
type Init func() error
