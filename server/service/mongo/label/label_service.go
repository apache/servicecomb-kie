package label

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
)

//Service is db service
type Service struct {
}

//CreateOrUpdate create or update labels
func (s *Service) CreateOrUpdate(ctx context.Context, label *model.LabelDoc) (*model.LabelDoc, error) {
	if label.ID != "" {
		return UpdateLabel(ctx, label)
	}
	return CreateLabel(ctx, label)
}
