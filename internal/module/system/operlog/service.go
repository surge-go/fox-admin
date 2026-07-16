package operlog

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

var tracer = otel.Tracer("fox-admin/internal/module/system/operlog")

// Service 表示操作日志业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
	now    func() time.Time
}

// NewService 创建操作日志业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("operation log service db is nil")
	}
	if logger == nil {
		panic("operation log service logger is nil")
	}
	return &Service{db: db, logger: logger, now: time.Now}
}

// Record 写入一次操作审计结果。
func (s *Service) Record(ctx context.Context, input *RecordInput) (err error) {
	ctx, span := tracer.Start(ctx, "system.operlog.Record")
	defer func() { tracing.FinishSpan(span, err) }()
	if input == nil {
		return errcode.ErrOperLogRecordReqNil
	}
	if !enum.IsStatusValid(input.Status) {
		return errcode.ErrOperLogStatusInvalid
	}
	module := strings.TrimSpace(input.Module)
	if module == "" {
		return errcode.ErrOperLogModuleRequired
	}
	action := strings.TrimSpace(input.Action)
	if action == "" {
		return errcode.ErrOperLogActionRequired
	}
	method := strings.ToUpper(strings.TrimSpace(input.Method))
	if method == "" {
		return errcode.ErrOperLogMethodRequired
	}
	path := strings.TrimSpace(input.Path)
	if path == "" {
		return errcode.ErrOperLogPathRequired
	}
	username := strings.TrimSpace(input.Username)
	if username == "" && input.UserID != nil {
		var user entity.User
		queryErr := s.db.WithContext(ctx).Unscoped().Select("username").Where("id = ?", *input.UserID).Take(&user).Error
		if queryErr == nil {
			username = user.Username
		} else if !errors.Is(queryErr, gorm.ErrRecordNotFound) {
			s.logger.Warn("写入操作日志：查询操作用户失败", zap.Int64("user_id", *input.UserID), zap.Error(queryErr))
		}
	}
	statusCode := input.StatusCode
	if statusCode <= 0 {
		statusCode = 200
	}
	item := &entity.OperLog{
		RequestID:    optionalString(input.RequestID, 120),
		TraceID:      optionalString(input.TraceID, 120),
		UserID:       input.UserID,
		Username:     optionalString(username, 120),
		Module:       truncateString(module, 120),
		Action:       truncateString(action, 120),
		Method:       truncateString(method, 16),
		Path:         truncateString(path, 500),
		IP:           optionalString(input.IP, 64),
		UserAgent:    optionalString(input.UserAgent, 500),
		RequestData:  optionalString(input.RequestData, 4096),
		Status:       input.Status,
		StatusCode:   statusCode,
		BusinessCode: input.BusinessCode,
		CostMillis:   max(input.CostMillis, 0),
		ErrorMessage: optionalString(input.ErrorMessage, 500),
	}
	span.SetAttributes(
		attribute.String("system.module", "operlog"),
		attribute.String("system.operation", "record"),
		attribute.String("operation.module", module),
		attribute.String("operation.action", action),
		attribute.Int("operation.status", input.Status),
		attribute.Int("operation.business_code", input.BusinessCode),
	)
	if input.UserID != nil {
		span.SetAttributes(attribute.Int64("user.id", *input.UserID))
	}
	if createErr := s.db.WithContext(ctx).Create(item).Error; createErr != nil {
		s.logger.Error("写入操作日志失败", zap.String("module", module), zap.String("action", action), zap.Error(createErr))
		return errcode.ErrOperLogRecordFailed.WithErr(createErr)
	}
	return nil
}

// List 查询操作日志分页列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp *dto.PageResp[*ListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.operlog.List")
	defer func() { tracing.FinishSpan(span, err) }()
	page, size := enum.DefaultPage, enum.DefaultSize
	var username, module, action, method, path, ip string
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
		module = strings.TrimSpace(req.Module)
		action = strings.TrimSpace(req.Action)
		method = strings.ToUpper(strings.TrimSpace(req.Method))
		path = strings.TrimSpace(req.Path)
		ip = strings.TrimSpace(req.IP)
		status = req.Status
		businessCode = req.BusinessCode
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrOperLogStatusInvalid
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
	query := s.db.WithContext(ctx).Table(entity.OperLog{}.TableName() + " AS o")
	if username != "" {
		query = query.Where("o.username LIKE ?", "%"+username+"%")
	}
	if module != "" {
		query = query.Where("o.module = ?", module)
	}
	if action != "" {
		query = query.Where("o.action = ?", action)
	}
	if method != "" {
		query = query.Where("o.method = ?", method)
	}
	if path != "" {
		query = query.Where("o.path LIKE ?", "%"+path+"%")
	}
	if ip != "" {
		query = query.Where("o.ip LIKE ?", "%"+ip+"%")
	}
	if status != nil {
		query = query.Where("o.status = ?", *status)
	}
	if businessCode != nil {
		query = query.Where("o.business_code = ?", *businessCode)
	}
	if startTime != nil {
		query = query.Where("o.created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("o.created_at <= ?", *endTime)
	}
	var total int64
	if queryErr := query.Count(&total).Error; queryErr != nil {
		s.logger.Error("查询操作日志列表失败：统计日志失败", zap.Error(queryErr))
		return nil, errcode.ErrOperLogListQueryFailed.WithErr(queryErr)
	}
	var items []*ListItemResp
	if queryErr := query.Select(operLogSelectColumns("o")).Order("o.created_at DESC, o.id DESC").
		Limit(size).Offset((page - 1) * size).Find(&items).Error; queryErr != nil {
		s.logger.Error("查询操作日志列表失败：查询日志失败", zap.Int("page", page), zap.Int("size", size), zap.Error(queryErr))
		return nil, errcode.ErrOperLogListQueryFailed.WithErr(queryErr)
	}
	if items == nil {
		items = make([]*ListItemResp, 0)
	}
	span.SetAttributes(attribute.Int64("operlog.total", total), attribute.Int("operlog.count", len(items)))
	return dto.NewPageResp(items, total), nil
}

// Detail 查询操作日志详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.operlog.Detail")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrOperLogDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrOperLogIDInvalid
	}
	span.SetAttributes(attribute.Int64("operlog.id", req.ID))
	var detail DetailResp
	if queryErr := s.db.WithContext(ctx).Table(entity.OperLog{}.TableName()+" AS o").
		Select(operLogSelectColumns("o")).Where("o.id = ?", req.ID).Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrOperLogNotFound
		}
		s.logger.Error("查询操作日志详情失败", zap.Int64("oper_log_id", req.ID), zap.Error(queryErr))
		return nil, errcode.ErrOperLogQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// Delete 批量硬删除操作日志。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.operlog.Delete")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrOperLogDeleteReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrOperLogIDsRequired
	}
	ids, normalizeErr := normalizeIDs(req.IDs)
	if normalizeErr != nil {
		return normalizeErr
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			var locked []entity.OperLog
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id IN ?", batch).
				Order("id ASC").Find(&locked).Error; queryErr != nil {
				return errcode.ErrOperLogQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batch) {
				return errcode.ErrOperLogNotFound
			}
			if deleteErr := tx.Where("id IN ?", batch).Delete(&entity.OperLog{}).Error; deleteErr != nil {
				return errcode.ErrOperLogDeleteFailed.WithErr(deleteErr)
			}
		}
		return nil
	})
}

// Clean 分批清理截止时间之前的操作日志。
func (s *Service) Clean(ctx context.Context, req *CleanReq) (resp *CleanResp, err error) {
	ctx, span := tracer.Start(ctx, "system.operlog.Clean")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrOperLogCleanReqNil
	}
	beforeText := strings.TrimSpace(req.Before)
	if beforeText == "" {
		return nil, errcode.ErrOperLogCleanBeforeRequired
	}
	before, parseErr := time.Parse(time.RFC3339, beforeText)
	if parseErr != nil {
		return nil, errcode.ErrOperLogTimeInvalid
	}
	if !before.Before(s.now()) {
		return nil, errcode.ErrOperLogCleanBeforeFuture
	}
	var deleted int64
	for {
		var batchDeleted int64
		batchErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var ids []int64
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&entity.OperLog{}).
				Select("id").Where("created_at < ?", before).Order("created_at ASC, id ASC").
				Limit(cleanBatchSize).Pluck("id", &ids).Error; queryErr != nil {
				return queryErr
			}
			if len(ids) == 0 {
				return nil
			}
			result := tx.Where("id IN ?", ids).Delete(&entity.OperLog{})
			if result.Error != nil {
				return result.Error
			}
			batchDeleted = result.RowsAffected
			return nil
		})
		if batchErr != nil {
			s.logger.Error("清理操作日志失败", zap.Time("before", before), zap.Error(batchErr))
			return nil, errcode.ErrOperLogCleanFailed.WithErr(batchErr)
		}
		deleted += batchDeleted
		if batchDeleted == 0 {
			break
		}
	}
	span.SetAttributes(attribute.Int64("operlog.deleted", deleted))
	return &CleanResp{Deleted: deleted}, nil
}

func parseTimeRange(startText, endText string) (*time.Time, *time.Time, error) {
	var startTime, endTime *time.Time
	if startText = strings.TrimSpace(startText); startText != "" {
		parsed, err := time.Parse(time.RFC3339, startText)
		if err != nil {
			return nil, nil, errcode.ErrOperLogTimeInvalid
		}
		startTime = &parsed
	}
	if endText = strings.TrimSpace(endText); endText != "" {
		parsed, err := time.Parse(time.RFC3339, endText)
		if err != nil {
			return nil, nil, errcode.ErrOperLogTimeInvalid
		}
		endTime = &parsed
	}
	if startTime != nil && endTime != nil && startTime.After(*endTime) {
		return nil, nil, errcode.ErrOperLogTimeRangeInvalid
	}
	return startTime, endTime, nil
}

func operLogSelectColumns(alias string) string {
	return alias + ".id, " + alias + ".request_id, " + alias + ".trace_id, " + alias + ".user_id, " +
		alias + ".username, " + alias + ".module, " + alias + ".action, " + alias + ".method, " +
		alias + ".path, " + alias + ".ip, " + alias + ".user_agent, " + alias + ".request_data, " +
		alias + ".status, " + alias + ".status_code, " + alias + ".business_code, " +
		alias + ".cost_millis, " + alias + ".error_message, " + alias + ".created_at"
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
			return nil, errcode.ErrOperLogIDInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}
