package handler

import (
	"github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/core/invocation"
)

//RecordHandler is a handle help record polling data
type RecordHandler struct{}

//Handle to make the wait group done after return response
func (bk *RecordHandler) Handle(chain *handler.Chain, inv *invocation.Invocation, cb invocation.ResponseCallBack) {
	chain.Next(inv, cb)
	v1.Wg.Done()
}

func newRecorder() handler.Handler {
	return &RecordHandler{}
}

//Name is handler name
func (bk *RecordHandler) Name() string {
	return "record-handler"
}
func init() {
	handler.RegisterHandler("record-handler", newRecorder)
}
