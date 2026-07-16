package loginlog

import (
	"context"
	"errors"
	"strings"
	"time"

	"fox-admin/internal/dto"
	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/internal/observability/tracing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const cleanBatchSize = 500

var tracer = otel.Tracer("fox-admin/internal/module/system/loginlog")

// Service 表示登录日志业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
	now    func() time.Time
}

// NewService 创建登录日志业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("login log service db is nil")
	}
	if logger == nil {
		panic("login log service logger is nil")
	}
	return &Service{db: db, logger: logger, now: time.Now}
}

// Record 写入一次登录结果。
func (s *Service) Record(ctx context.Context, input *RecordInput) (err error) {
	ctx, span := tracer.Start(ctx, "system.loginlog.Record")
	defer func() { tracing.FinishSpan(span, err) }()
	if input == nil {
		return errcode.ErrLoginLogRecordReqNil
	}
	if !enum.IsStatusValid(input.Status) {
		return errcode.ErrLoginLogStatusInvalid
	}

	username := strings.TrimSpace(input.Username)
	item := &entity.LoginLog{
		RequestID:    optionalString(input.RequestID, 120),
		TraceID:      optionalString(input.TraceID, 120),
		UserID:       input.UserID,
		Username:     truncateString(username, 120),
		IP:           optionalString(input.IP, 64),
		UserAgent:    optionalString(input.UserAgent, 500),
		Platform:     optionalString(input.Platform, 32),
		DeviceIDHash: optionalString(input.DeviceIDHash, 64),
		Status:       input.Status,
		BusinessCode: input.BusinessCode,
		Message:      optionalString(input.Message, 255),
	}
	span.SetAttributes(
		attribute.String("system.module", "loginlog"),
		attribute.String("system.operation", "record"),
		attribute.Int("login.status", input.Status),
		attribute.Int("login.business_code", input.BusinessCode),
	)
	if input.UserID != nil {
		span.SetAttributes(attribute.Int64("user.id", *input.UserID))
	}
	if createErr := s.db.WithContext(ctx).Create(item).Error; createErr != nil {
		s.logger.Error("写入登录日志失败", zap.String("username", username), zap.Error(createErr))
		return errcode.ErrLoginLogRecordFailed.WithErr(createErr)
	}
	return nil
}

// List 查询登录日志分页列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp *dto.PageResp[*ListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.loginlog.List")
	defer func() { tracing.FinishSpan(span, err) }()
	page, size := enum.DefaultPage, enum.DefaultSize
	var username, ip string
	var status, businessCode *int
	var startTime, endTime *time.Time
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
		username = strings.TrimSpace(req.Username)
		ip = strings.TrimSpace(req.IP)
		status = req.Status
		businessCode = req.BusinessCode
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrLoginLogStatusInvalid
		}
		var parseErr error
		startTime, endTime, parseErr = parseTimeRange(req.StartTime, req.EndTime)
		if parseErr != nil {
			return nil, parseErr
		}
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}

	query := s.db.WithContext(ctx).Table(entity.LoginLog{}.TableName() + " AS l")
	if username != "" {
		query = query.Where("l.username LIKE ?", "%"+username+"%")
	}
	if ip != "" {
		query = query.Where("l.ip LIKE ?", "%"+ip+"%")
	}
	if status != nil {
		query = query.Where("l.status = ?", *status)
	}
	if businessCode != nil {
		query = query.Where("l.business_code = ?", *businessCode)
	}
	if startTime != nil {
		query = query.Where("l.created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("l.created_at <= ?", *endTime)
	}

	var total int64
	if queryErr := query.Count(&total).Error; queryErr != nil {
		s.logger.Error("查询登录日志列表失败：统计日志失败", zap.Error(queryErr))
		return nil, errcode.ErrLoginLogListQueryFailed.WithErr(queryErr)
	}
	var items []*ListItemResp
	if queryErr := query.Select(loginLogSelectColumns("l")).
		Order("l.created_at DESC, l.id DESC").Limit(size).Offset((page - 1) * size).Find(&items).Error; queryErr != nil {
		s.logger.Error("查询登录日志列表失败：查询日志失败", zap.Int("page", page), zap.Int("size", size), zap.Error(queryErr))
		return nil, errcode.ErrLoginLogListQueryFailed.WithErr(queryErr)
	}
	if items == nil {
		items = make([]*ListItemResp, 0)
	}
	span.SetAttributes(attribute.Int64("loginlog.total", total), attribute.Int("loginlog.count", len(items)))
	return dto.NewPageResp(items, total), nil
}

// Detail 查询登录日志详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.loginlog.Detail")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrLoginLogDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrLoginLogIDInvalid
	}
	span.SetAttributes(attribute.Int64("loginlog.id", req.ID))
	var detail DetailResp
	if queryErr := s.db.WithContext(ctx).Table(entity.LoginLog{}.TableName()+" AS l").
		Select(loginLogSelectColumns("l")).Where("l.id = ?", req.ID).Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrLoginLogNotFound
		}
		s.logger.Error("查询登录日志详情失败", zap.Int64("login_log_id", req.ID), zap.Error(queryErr))
		return nil, errcode.ErrLoginLogQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// Delete 批量硬删除登录日志。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.loginlog.Delete")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrLoginLogDeleteReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrLoginLogIDsRequired
	}
	ids, normalizeErr := normalizeIDs(req.IDs)
	if normalizeErr != nil {
		return normalizeErr
	}
	span.SetAttributes(attribute.Int("loginlog.batch_size", len(ids)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			var locked []entity.LoginLog
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").
				Where("id IN ?", batch).Order("id ASC").Find(&locked).Error; queryErr != nil {
				s.logger.Error("删除登录日志失败：查询日志失败", zap.Int64s("login_log_ids", batch), zap.Error(queryErr))
				return errcode.ErrLoginLogQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batch) {
				return errcode.ErrLoginLogNotFound
			}
			if deleteErr := tx.Where("id IN ?", batch).Delete(&entity.LoginLog{}).Error; deleteErr != nil {
				s.logger.Error("删除登录日志失败", zap.Int64s("login_log_ids", batch), zap.Error(deleteErr))
				return errcode.ErrLoginLogDeleteFailed.WithErr(deleteErr)
			}
		}
		return nil
	})
}

// Clean 分批清理截止时间之前的登录日志。
func (s *Service) Clean(ctx context.Context, req *CleanReq) (resp *CleanResp, err error) {
	ctx, span := tracer.Start(ctx, "system.loginlog.Clean")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrLoginLogCleanReqNil
	}
	beforeText := strings.TrimSpace(req.Before)
	if beforeText == "" {
		return nil, errcode.ErrLoginLogCleanBeforeRequired
	}
	before, parseErr := time.Parse(time.RFC3339, beforeText)
	if parseErr != nil {
		return nil, errcode.ErrLoginLogTimeInvalid
	}
	if !before.Before(s.now()) {
		return nil, errcode.ErrLoginLogCleanBeforeFuture
	}
	span.SetAttributes(attribute.String("loginlog.clean_before", before.Format(time.RFC3339)))

	var deleted int64
	for {
		var batchDeleted int64
		batchErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var ids []int64
			if queryErr := tx.Model(&entity.LoginLog{}).Select("id").Where("created_at < ?", before).
				Order("created_at ASC, id ASC").Limit(cleanBatchSize).Pluck("id", &ids).Error; queryErr != nil {
				return queryErr
			}
			if len(ids) == 0 {
				return nil
			}
			result := tx.Where("id IN ?", ids).Delete(&entity.LoginLog{})
			if result.Error != nil {
				return result.Error
			}
			batchDeleted = result.RowsAffected
			return nil
		})
		if batchErr != nil {
			s.logger.Error("清理登录日志失败", zap.Time("before", before), zap.Error(batchErr))
			return nil, errcode.ErrLoginLogCleanFailed.WithErr(batchErr)
		}
		deleted += batchDeleted
		if batchDeleted == 0 {
			break
		}
	}
	span.SetAttributes(attribute.Int64("loginlog.deleted", deleted))
	return &CleanResp{Deleted: deleted}, nil
}

func parseTimeRange(startText, endText string) (*time.Time, *time.Time, error) {
	var startTime, endTime *time.Time
	startText = strings.TrimSpace(startText)
	if startText != "" {
		parsed, err := time.Parse(time.RFC3339, startText)
		if err != nil {
			return nil, nil, errcode.ErrLoginLogTimeInvalid
		}
		startTime = &parsed
	}
	endText = strings.TrimSpace(endText)
	if endText != "" {
		parsed, err := time.Parse(time.RFC3339, endText)
		if err != nil {
			return nil, nil, errcode.ErrLoginLogTimeInvalid
		}
		endTime = &parsed
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, nil, errcode.ErrLoginLogTimeRangeInvalid
	}
	return startTime, endTime, nil
}

func loginLogSelectColumns(alias string) string {
	return alias + ".id, " + alias + ".request_id, " + alias + ".trace_id, " +
		alias + ".user_id, " + alias + ".username, " + alias + ".ip, " +
		alias + ".location, " + alias + ".browser, " + alias + ".os, " +
		alias + ".user_agent, " + alias + ".platform, " + alias + ".device_id_hash, " +
		alias + ".status, " + alias + ".business_code, " + alias + ".message, " + alias + ".created_at"
}

func optionalString(value string, maxLength int) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	value = truncateString(value, maxLength)
	return &value
}

func truncateString(value string, maxLength int) string {
	runes := []rune(value)
	if len(runes) <= maxLength {
		return value
	}
	return string(runes[:maxLength])
}

func normalizeIDs(input []int64) ([]int64, error) {
	ids := make([]int64, 0, len(input))
	seen := make(map[int64]struct{}, len(input))
	for _, id := range input {
		if id <= 0 {
			return nil, errcode.ErrLoginLogIDInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}
