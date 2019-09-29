package mongo

import (
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/history"
	"github.com/apache/servicecomb-kie/server/service/mongo/kv"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
)

func init() {
	openlogging.Info("use mongodb as storage")
	service.DBInit = session.Init
	service.KVService = &kv.Service{}
	service.HistoryService = &history.Service{}
}
