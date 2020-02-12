package record

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
)

//Service is db service
type Service struct {
}

//RecordSuccess record success
func (s *Service) RecordSuccess(ctx context.Context, detail *model.PollingDetail, respStatus int, respData, respHeader string) (*model.PollingDetail, error) {
	detail.ResponseCode = respStatus
	detail.ResponseBody = respData
	detail.ResponseHeader = respHeader
	return CreateOrUpdateRecord(ctx, detail)
}

//RecordSuccess record success
func (s *Service) RecordFailed(ctx context.Context, detail *model.PollingDetail, respStatus int, respData string) (*model.PollingDetail, error) {
	detail.ResponseCode = respStatus
	detail.ResponseBody = respData
	return CreateOrUpdateRecord(ctx, detail)
}
