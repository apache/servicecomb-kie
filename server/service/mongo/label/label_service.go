package label

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
)

type Service struct {
}

func (s *Service) CreateOrUpdate(ctx context.Context, label *model.LabelDoc) (*model.LabelDoc, error) {
	if label.ID != "" {
		return UpdateLabel(ctx, label)
	}
	return CreateLabel(ctx, label)
}
