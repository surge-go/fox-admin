package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

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

var (
	tracer           = otel.Tracer("fox-admin/internal/module/system/config")
	configKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
)

// Service 表示系统配置业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建系统配置业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("config service db is nil")
	}
	if logger == nil {
		panic("config service logger is nil")
	}
	return &Service{db: db, logger: logger}
}

// Create 创建自定义配置。
func (s *Service) Create(ctx context.Context, req *CreateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.config.Create")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrConfigCreateReqNil
	}

	name, key, group, valueType, value, validationErr := validateCreateFields(req)
	if validationErr != nil {
		return validationErr
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrConfigStatusInvalid
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(
		attribute.String("system.module", "config"),
		attribute.String("system.operation", "create"),
		attribute.String("config.key", key),
		attribute.String("config.value_type", valueType),
	)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if queryErr := s.ensureKeyUnique(tx, key); queryErr != nil {
			return queryErr
		}
		item := &entity.Config{
			Name:      name,
			Key:       key,
			Value:     value,
			Group:     group,
			ValueType: valueType,
			IsBuiltin: false,
			Status:    &status,
			Remark:    remark,
		}
		if createErr := tx.Create(item).Error; createErr != nil {
			s.logger.Error("创建配置失败：写入配置失败", zap.String("config_key", key), zap.Error(createErr))
			return errcode.ErrConfigCreateFailed.WithErr(createErr)
		}
		return nil
	})
}

// Delete 批量删除非内置配置。
func (s *Service) Delete(ctx context.Context, req *DeleteReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.config.Delete")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrConfigDeleteReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrConfigIDsRequired
	}
	ids, normalizeErr := normalizeIDs(req.IDs)
	if normalizeErr != nil {
		return normalizeErr
	}
	span.SetAttributes(attribute.Int("config.batch_size", len(ids)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			var locked []entity.Config
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Select("id", "is_builtin").Where("id IN ?", batch).Order("id ASC").Find(&locked).Error; queryErr != nil {
				s.logger.Error("删除配置失败：查询配置失败", zap.Int64s("config_ids", batch), zap.Error(queryErr))
				return errcode.ErrConfigQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batch) {
				return errcode.ErrConfigNotFound
			}
			for i := range locked {
				if locked[i].IsBuiltin {
					return errcode.ErrConfigBuiltinDelete
				}
			}
			if deleteErr := tx.Where("id IN ?", batch).Delete(&entity.Config{}).Error; deleteErr != nil {
				s.logger.Error("删除配置失败：删除配置失败", zap.Int64s("config_ids", batch), zap.Error(deleteErr))
				return errcode.ErrConfigDeleteFailed.WithErr(deleteErr)
			}
		}
		return nil
	})
}

// Update 更新配置展示信息和值；配置键和值类型保持不变。
func (s *Service) Update(ctx context.Context, req *UpdateReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.config.Update")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrConfigUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrConfigIDInvalid
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrConfigNameRequired
	}
	group := normalizeGroup(req.Group)
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrConfigStatusInvalid
	}
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(attribute.Int64("config.id", req.ID))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.Config
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).Take(&current).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return errcode.ErrConfigNotFound
			}
			s.logger.Error("更新配置失败：查询配置失败", zap.Int64("config_id", req.ID), zap.Error(queryErr))
			return errcode.ErrConfigQueryFailed.WithErr(queryErr)
		}
		value, valueErr := normalizeValue(current.ValueType, req.Value)
		if valueErr != nil {
			return valueErr
		}
		status := enum.StatusEnabled
		if current.Status != nil {
			status = *current.Status
		}
		if req.Status != nil {
			status = *req.Status
		}
		updates := map[string]any{
			"name":         name,
			"config_value": value,
			"config_group": group,
			"status":       status,
			"remark":       remark,
		}
		if updateErr := tx.Model(&entity.Config{}).Where("id = ?", req.ID).Updates(updates).Error; updateErr != nil {
			s.logger.Error("更新配置失败：写入配置失败", zap.Int64("config_id", req.ID), zap.String("config_key", current.Key), zap.Error(updateErr))
			return errcode.ErrConfigUpdateFailed.WithErr(updateErr)
		}
		return nil
	})
}

// List 查询配置分页列表。
func (s *Service) List(ctx context.Context, req *ListReq) (resp *dto.PageResp[*ListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.config.List")
	defer func() { tracing.FinishSpan(span, err) }()
	page, size := enum.DefaultPage, enum.DefaultSize
	var name, key, group, valueType string
	var status *int
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
		name = strings.TrimSpace(req.Name)
		key = strings.TrimSpace(req.Key)
		group = strings.TrimSpace(req.Group)
		valueType = strings.ToLower(strings.TrimSpace(req.ValueType))
		status = req.Status
		if valueType != "" && !enum.IsConfigValueTypeValid(valueType) {
			return nil, errcode.ErrConfigValueTypeInvalid
		}
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrConfigStatusInvalid
		}
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}

	query := s.db.WithContext(ctx).Table(entity.Config{}.TableName()+" AS c").Where("c.deleted_at = ?", 0)
	if name != "" {
		query = query.Where("c.name LIKE ?", "%"+name+"%")
	}
	if key != "" {
		query = query.Where("c.config_key LIKE ?", "%"+key+"%")
	}
	if group != "" {
		query = query.Where("c.config_group = ?", group)
	}
	if valueType != "" {
		query = query.Where("c.value_type = ?", valueType)
	}
	if status != nil {
		query = query.Where("c.status = ?", *status)
	}
	var total int64
	if queryErr := query.Count(&total).Error; queryErr != nil {
		s.logger.Error("查询配置列表失败：统计配置失败", zap.Error(queryErr))
		return nil, errcode.ErrConfigListQueryFailed.WithErr(queryErr)
	}
	var items []*ListItemResp
	if queryErr := query.Select("c.id, c.name, c.config_key, c.config_value, c.config_group, c.value_type, c.is_builtin, c.status, c.remark, c.created_at, c.updated_at").
		Order("c.config_group ASC, c.id DESC").Limit(size).Offset((page - 1) * size).Find(&items).Error; queryErr != nil {
		s.logger.Error("查询配置列表失败：查询配置失败", zap.Int("page", page), zap.Int("size", size), zap.Error(queryErr))
		return nil, errcode.ErrConfigListQueryFailed.WithErr(queryErr)
	}
	if items == nil {
		items = make([]*ListItemResp, 0)
	}
	span.SetAttributes(attribute.Int64("config.total", total), attribute.Int("config.count", len(items)))
	return dto.NewPageResp(items, total), nil
}

// Detail 查询配置详情。
func (s *Service) Detail(ctx context.Context, req *DetailReq) (resp *DetailResp, err error) {
	ctx, span := tracer.Start(ctx, "system.config.Detail")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrConfigDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrConfigIDInvalid
	}
	span.SetAttributes(attribute.Int64("config.id", req.ID))
	var detail DetailResp
	if queryErr := s.db.WithContext(ctx).Model(&entity.Config{}).
		Select("id, name, config_key, config_value, config_group, value_type, is_builtin, status, remark, created_at, updated_at").
		Where("id = ?", req.ID).Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrConfigNotFound
		}
		s.logger.Error("查询配置详情失败：查询配置失败", zap.Int64("config_id", req.ID), zap.Error(queryErr))
		return nil, errcode.ErrConfigQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// UpdateStatus 批量更新配置状态。
func (s *Service) UpdateStatus(ctx context.Context, req *UpdateStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.config.UpdateStatus")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrConfigUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrConfigIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrConfigStatusInvalid
	}
	ids, normalizeErr := normalizeIDs(req.IDs)
	if normalizeErr != nil {
		return normalizeErr
	}
	span.SetAttributes(attribute.Int("config.batch_size", len(ids)), attribute.Int("config.status", *req.Status))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			var locked []entity.Config
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").
				Where("id IN ?", batch).Order("id ASC").Find(&locked).Error; queryErr != nil {
				s.logger.Error("更新配置状态失败：查询配置失败", zap.Int64s("config_ids", batch), zap.Error(queryErr))
				return errcode.ErrConfigQueryFailed.WithErr(queryErr)
			}
			if len(locked) != len(batch) {
				return errcode.ErrConfigNotFound
			}
			if updateErr := tx.Model(&entity.Config{}).Where("id IN ?", batch).Update("status", *req.Status).Error; updateErr != nil {
				s.logger.Error("更新配置状态失败：写入配置失败", zap.Int64s("config_ids", batch), zap.Int("status", *req.Status), zap.Error(updateErr))
				return errcode.ErrConfigUpdateStatusFailed.WithErr(updateErr)
			}
		}
		return nil
	})
}

// Get 读取启用配置的原始文本值。
func (s *Service) Get(ctx context.Context, key string) (value string, err error) {
	item, err := s.getEnabled(ctx, key)
	if err != nil {
		return "", err
	}
	return item.Value, nil
}

// GetString 读取字符串配置。
func (s *Service) GetString(ctx context.Context, key string) (value string, err error) {
	item, err := s.getEnabled(ctx, key)
	if err != nil {
		return "", err
	}
	if item.ValueType != enum.ConfigValueTypeString {
		return "", errcode.ErrConfigValueTypeMismatch
	}
	return item.Value, nil
}

// GetInt64 读取整数配置。
func (s *Service) GetInt64(ctx context.Context, key string) (value int64, err error) {
	item, err := s.getEnabled(ctx, key)
	if err != nil {
		return 0, err
	}
	if item.ValueType != enum.ConfigValueTypeInt {
		return 0, errcode.ErrConfigValueTypeMismatch
	}
	value, parseErr := strconv.ParseInt(item.Value, 10, 64)
	if parseErr != nil {
		s.logger.Error("读取整数配置失败：配置值非法", zap.String("config_key", item.Key), zap.Error(parseErr))
		return 0, errcode.ErrConfigValueInvalid.WithErr(parseErr)
	}
	return value, nil
}

// GetBool 读取布尔配置。
func (s *Service) GetBool(ctx context.Context, key string) (value bool, err error) {
	item, err := s.getEnabled(ctx, key)
	if err != nil {
		return false, err
	}
	if item.ValueType != enum.ConfigValueTypeBool {
		return false, errcode.ErrConfigValueTypeMismatch
	}
	switch item.Value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		s.logger.Error("读取布尔配置失败：配置值非法", zap.String("config_key", item.Key), zap.String("config_value", item.Value))
		return false, errcode.ErrConfigValueInvalid
	}
}

// DecodeJSON 将 JSON 配置解码到 dst。
func (s *Service) DecodeJSON(ctx context.Context, key string, dst any) (err error) {
	if dst == nil {
		return errcode.ErrConfigValueInvalid
	}
	item, err := s.getEnabled(ctx, key)
	if err != nil {
		return err
	}
	if item.ValueType != enum.ConfigValueTypeJSON {
		return errcode.ErrConfigValueTypeMismatch
	}
	if decodeErr := json.Unmarshal([]byte(item.Value), dst); decodeErr != nil {
		s.logger.Error("读取 JSON 配置失败：配置值非法", zap.String("config_key", item.Key), zap.Error(decodeErr))
		return errcode.ErrConfigValueInvalid.WithErr(decodeErr)
	}
	return nil
}

func (s *Service) getEnabled(ctx context.Context, key string) (*entity.Config, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errcode.ErrConfigKeyRequired
	}
	if !configKeyPattern.MatchString(key) {
		return nil, errcode.ErrConfigKeyInvalid
	}
	var item entity.Config
	if queryErr := s.db.WithContext(ctx).Where("config_key = ? AND status = ?", key, enum.StatusEnabled).Take(&item).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrConfigNotFound
		}
		s.logger.Error("读取配置失败：查询配置失败", zap.String("config_key", key), zap.Error(queryErr))
		return nil, errcode.ErrConfigQueryFailed.WithErr(queryErr)
	}
	return &item, nil
}

func (s *Service) ensureKeyUnique(tx *gorm.DB, key string) error {
	var count int64
	if queryErr := tx.Model(&entity.Config{}).Where("config_key = ?", key).Count(&count).Error; queryErr != nil {
		s.logger.Error("校验配置键失败：查询配置键失败", zap.String("config_key", key), zap.Error(queryErr))
		return errcode.ErrConfigKeyQueryFailed.WithErr(queryErr)
	}
	if count > 0 {
		return errcode.ErrConfigKeyExists
	}
	return nil
}

func validateCreateFields(req *CreateReq) (name, key, group, valueType, value string, err error) {
	name = strings.TrimSpace(req.Name)
	if name == "" {
		return "", "", "", "", "", errcode.ErrConfigNameRequired
	}
	key = strings.TrimSpace(req.Key)
	if key == "" {
		return "", "", "", "", "", errcode.ErrConfigKeyRequired
	}
	if !configKeyPattern.MatchString(key) {
		return "", "", "", "", "", errcode.ErrConfigKeyInvalid
	}
	group = normalizeGroup(req.Group)
	valueType = strings.ToLower(strings.TrimSpace(req.ValueType))
	if valueType == "" {
		valueType = enum.ConfigValueTypeString
	}
	if !enum.IsConfigValueTypeValid(valueType) {
		return "", "", "", "", "", errcode.ErrConfigValueTypeInvalid
	}
	value, err = normalizeValue(valueType, req.Value)
	if err != nil {
		return "", "", "", "", "", err
	}
	return name, key, group, valueType, value, nil
}

func normalizeValue(valueType, value string) (string, error) {
	switch valueType {
	case enum.ConfigValueTypeString:
		return value, nil
	case enum.ConfigValueTypeInt:
		parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return "", errcode.ErrConfigValueInvalid
		}
		return strconv.FormatInt(parsed, 10), nil
	case enum.ConfigValueTypeBool:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "true":
			return "true", nil
		case "false":
			return "false", nil
		default:
			return "", errcode.ErrConfigValueInvalid
		}
	case enum.ConfigValueTypeJSON:
		if !json.Valid([]byte(value)) {
			return "", errcode.ErrConfigValueInvalid
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, []byte(value)); err != nil {
			return "", errcode.ErrConfigValueInvalid
		}
		return compact.String(), nil
	default:
		return "", errcode.ErrConfigValueTypeInvalid
	}
}

func normalizeGroup(group string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return enum.DefaultConfigGroup
	}
	return group
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeIDs(input []int64) ([]int64, error) {
	ids := make([]int64, 0, len(input))
	seen := make(map[int64]struct{}, len(input))
	for _, id := range input {
		if id <= 0 {
			return nil, errcode.ErrConfigIDInvalid
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}
