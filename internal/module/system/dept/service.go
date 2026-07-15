package dept

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service 表示部门业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建部门业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("dept service db is nil")
	}
	if logger == nil {
		panic("dept service logger is nil")
	}

	return &Service{
		db:     db,
		logger: logger,
	}
}

// Create 创建部门。
func (s *Service) Create(ctx context.Context, req *CreateReq) error {
	return nil
}

// Delete 删除部门。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) error {
	return nil
}

// Update 更新部门。
func (s *Service) Update(ctx context.Context, req *UpdateReq) error {
	return nil
}

// Tree 查询部门管理树。
func (s *Service) Tree(ctx context.Context, req *TreeReq) ([]*TreeResp, error) {
	return []*TreeResp{}, nil
}

// Options 查询启用部门选项树。
func (s *Service) Options(ctx context.Context) ([]*OptionsResp, error) {
	return []*OptionsResp{}, nil
}

// Detail 查询部门详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (*DetailResp, error) {
	return nil, nil
}

// UpdateStatus 批量更新部门状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) error {
	return nil
}
