package dict

import (
	"context"
	"errors"
	"sort"
	"strings"

	"fox-admin/internal/dto"
	"fox-admin/internal/errcode"
	"fox-admin/internal/module/system/entity"
	"fox-admin/internal/module/system/enum"
	"fox-admin/internal/observability/tracing"
	"fox-admin/pkg/ptr"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	tracer                 = otel.Tracer("fox-admin/internal/module/system/dict")
	errDictDataTypeChanged = errors.New("dict data type changed concurrently")
)

// Service 表示字典业务服务。
type Service struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewService 创建字典业务服务。
func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	if db == nil {
		panic("dict service db is nil")
	}
	if logger == nil {
		panic("dict service logger is nil")
	}
	return &Service{db: db, logger: logger}
}

// CreateType 创建字典类型。
func (s *Service) CreateType(ctx context.Context, req *CreateTypeReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.CreateType")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrDictTypeCreateReqNil
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrDictTypeNameRequired
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		return errcode.ErrDictTypeCodeRequired
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDictTypeStatusInvalid
	}
	status := enum.StatusEnabled
	if req.Status != nil {
		status = *req.Status
	}
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(attribute.String("dict.type_code", code), attribute.Int("dict.status", status))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if checkErr := s.validateTypeUnique(tx, 0, name, code); checkErr != nil {
			return checkErr
		}
		item := &entity.DictType{Name: name, Code: code, Status: &status, Remark: remark}
		if createErr := tx.Create(item).Error; createErr != nil {
			s.logger.Error("创建字典类型失败：写入类型失败", zap.String("code", code), zap.Error(createErr))
			return errcode.ErrDictTypeCreateFailed.WithErr(createErr)
		}
		return nil
	})
}

// DeleteTypes 批量删除字典类型；任一类型仍有数据时整体拒绝。
func (s *Service) DeleteTypes(ctx context.Context, req *DeleteTypesReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.DeleteTypes")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil || len(req.IDs) == 0 {
		return errcode.ErrDictTypeDeleteReqNil
	}
	ids, normalizeErr := normalizeIDs(req.IDs, errcode.ErrDictTypeIDInvalid)
	if normalizeErr != nil {
		return normalizeErr
	}
	span.SetAttributes(attribute.Int("dict.batch_size", len(ids)))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batchIDs := ids[start:end]
			types, lockErr := s.lockTypesByIDs(tx, batchIDs)
			if lockErr != nil {
				return lockErr
			}
			codes := make([]string, 0, len(types))
			for i := range types {
				codes = append(codes, types[i].Code)
			}
			var dataCount int64
			if queryErr := tx.Model(&entity.DictData{}).Where("type_code IN ?", codes).Count(&dataCount).Error; queryErr != nil {
				s.logger.Error("删除字典类型失败：查询字典数据失败", zap.Strings("type_codes", codes), zap.Error(queryErr))
				return errcode.ErrDictTypeDataQueryFailed.WithErr(queryErr)
			}
			if dataCount > 0 {
				return errcode.ErrDictTypeHasData
			}
			if deleteErr := tx.Where("id IN ?", batchIDs).Delete(&entity.DictType{}).Error; deleteErr != nil {
				s.logger.Error("删除字典类型失败：删除类型失败", zap.Int64s("type_ids", batchIDs), zap.Error(deleteErr))
				return errcode.ErrDictTypeDeleteFailed.WithErr(deleteErr)
			}
		}
		return nil
	})
}

// UpdateType 更新字典类型；类型编码创建后不可修改。
func (s *Service) UpdateType(ctx context.Context, req *UpdateTypeReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.UpdateType")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrDictTypeUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrDictTypeIDInvalid
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errcode.ErrDictTypeNameRequired
	}
	if req.Status != nil && !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDictTypeStatusInvalid
	}
	remark := normalizeOptionalString(req.Remark)
	span.SetAttributes(attribute.Int64("dict.type_id", req.ID))

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current entity.DictType
		if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).Take(&current).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return errcode.ErrDictTypeNotFound
			}
			s.logger.Error("更新字典类型失败：锁定类型失败", zap.Int64("type_id", req.ID), zap.Error(queryErr))
			return errcode.ErrDictTypeQueryFailed.WithErr(queryErr)
		}
		if checkErr := s.validateTypeName(tx, req.ID, name); checkErr != nil {
			return checkErr
		}
		status := ptr.Value(current.Status)
		if req.Status != nil {
			status = *req.Status
		}
		if updateErr := tx.Model(&entity.DictType{}).Where("id = ?", req.ID).Updates(map[string]any{
			"name": name, "status": status, "remark": remark,
		}).Error; updateErr != nil {
			s.logger.Error("更新字典类型失败：写入类型失败", zap.Int64("type_id", req.ID), zap.Error(updateErr))
			return errcode.ErrDictTypeUpdateFailed.WithErr(updateErr)
		}
		return nil
	})
}

// ListTypes 查询字典类型分页列表。
func (s *Service) ListTypes(ctx context.Context, req *ListTypesReq) (resp *dto.PageResp[*TypeListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.dict.ListTypes")
	defer func() { tracing.FinishSpan(span, err) }()
	page, size := enum.DefaultPage, enum.DefaultSize
	var name, code string
	var status *int
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
		name, code, status = strings.TrimSpace(req.Name), strings.TrimSpace(req.Code), req.Status
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrDictTypeStatusInvalid
		}
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}
	query := s.db.WithContext(ctx).Table(entity.DictType{}.TableName()+" AS t").Where("t.deleted_at = ?", 0)
	if name != "" {
		query = query.Where("t.name LIKE ?", "%"+name+"%")
	}
	if code != "" {
		query = query.Where("t.code LIKE ?", "%"+code+"%")
	}
	if status != nil {
		query = query.Where("t.status = ?", *status)
	}
	var total int64
	if queryErr := query.Count(&total).Error; queryErr != nil {
		return nil, errcode.ErrDictTypeListQueryFailed.WithErr(queryErr)
	}
	var items []*TypeListItemResp
	if queryErr := query.Select("t.id, t.name, t.code, t.status, t.remark, t.created_at, t.updated_at").
		Order("t.id DESC").Limit(size).Offset((page - 1) * size).Find(&items).Error; queryErr != nil {
		return nil, errcode.ErrDictTypeListQueryFailed.WithErr(queryErr)
	}
	return dto.NewPageResp(items, total), nil
}

// ListTypeOptions 查询启用字典类型选项。
func (s *Service) ListTypeOptions(ctx context.Context) (resp *TypeOptionsResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dict.ListTypeOptions")
	defer func() { tracing.FinishSpan(span, err) }()
	var items []*TypeOptionItemResp
	if queryErr := s.db.WithContext(ctx).Table(entity.DictType{}.TableName()+" AS t").
		Select("t.id, t.name, t.code").Where("t.deleted_at = ? AND t.status = ?", 0, enum.StatusEnabled).
		Order("t.id ASC").Find(&items).Error; queryErr != nil {
		return nil, errcode.ErrDictTypeListQueryFailed.WithErr(queryErr)
	}
	if items == nil {
		items = []*TypeOptionItemResp{}
	}
	return &TypeOptionsResp{List: items}, nil
}

// DetailType 查询字典类型详情。
func (s *Service) DetailType(ctx context.Context, req *DetailTypeReq) (resp *DetailTypeResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dict.DetailType")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrDictTypeDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrDictTypeIDInvalid
	}
	var detail DetailTypeResp
	if queryErr := s.db.WithContext(ctx).Model(&entity.DictType{}).
		Select("id, name, code, status, remark, created_at, updated_at").Where("id = ?", req.ID).Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrDictTypeNotFound
		}
		return nil, errcode.ErrDictTypeQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// UpdateTypeStatus 批量更新字典类型状态。
func (s *Service) UpdateTypeStatus(ctx context.Context, req *UpdateTypeStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.UpdateTypeStatus")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrDictTypeUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrDictTypeIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDictTypeStatusInvalid
	}
	ids, normalizeErr := normalizeIDs(req.IDs, errcode.ErrDictTypeIDInvalid)
	if normalizeErr != nil {
		return normalizeErr
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			if _, lockErr := s.lockTypesByIDs(tx, batch); lockErr != nil {
				return lockErr
			}
			if updateErr := tx.Model(&entity.DictType{}).Where("id IN ?", batch).Update("status", *req.Status).Error; updateErr != nil {
				s.logger.Error("更新字典类型状态失败：写入状态失败", zap.Int64s("type_ids", batch), zap.Int("status", *req.Status), zap.Error(updateErr))
				return errcode.ErrDictTypeUpdateStatusFailed.WithErr(updateErr)
			}
		}
		return nil
	})
}

// CreateData 创建字典数据；设置默认项时锁定所属类型并替换原默认项。
func (s *Service) CreateData(ctx context.Context, req *CreateDataReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.CreateData")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrDictDataCreateReqNil
	}
	typeCode, label, value, validationErr := validateDataFields(req.TypeCode, req.Label, req.Value, req.Sort, req.Status)
	if validationErr != nil {
		return validationErr
	}
	sortValue, status, isDefault := enum.DefaultSort, enum.StatusEnabled, false
	if req.Sort != nil {
		sortValue = *req.Sort
	}
	if req.Status != nil {
		status = *req.Status
	}
	if req.IsDefault != nil {
		isDefault = *req.IsDefault
	}
	if isDefault && status != enum.StatusEnabled {
		return errcode.ErrDictDataDefaultDisabled
	}
	remark := normalizeOptionalString(req.Remark)

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, lockErr := s.lockTypeByCode(tx, typeCode); lockErr != nil {
			return lockErr
		}
		if checkErr := s.validateDataValue(tx, 0, typeCode, value); checkErr != nil {
			return checkErr
		}
		if isDefault {
			if updateErr := tx.Model(&entity.DictData{}).Where("type_code = ? AND is_default = ?", typeCode, true).Update("is_default", false).Error; updateErr != nil {
				s.logger.Error("创建字典数据失败：清除原默认项失败", zap.String("type_code", typeCode), zap.Error(updateErr))
				return errcode.ErrDictDataCreateFailed.WithErr(updateErr)
			}
		}
		item := &entity.DictData{TypeCode: typeCode, Label: label, Value: value, Sort: &sortValue, Status: &status, IsDefault: &isDefault, Remark: remark}
		if createErr := tx.Create(item).Error; createErr != nil {
			s.logger.Error("创建字典数据失败：写入数据失败", zap.String("type_code", typeCode), zap.Error(createErr))
			return errcode.ErrDictDataCreateFailed.WithErr(createErr)
		}
		return nil
	})
}

// DeleteData 批量删除字典数据。
func (s *Service) DeleteData(ctx context.Context, req *DeleteDataReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.DeleteData")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil || len(req.IDs) == 0 {
		return errcode.ErrDictDataDeleteReqNil
	}
	ids, normalizeErr := normalizeIDs(req.IDs, errcode.ErrDictDataIDInvalid)
	if normalizeErr != nil {
		return normalizeErr
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			if _, lockErr := s.lockDataByIDs(tx, batch); lockErr != nil {
				return lockErr
			}
			if deleteErr := tx.Where("id IN ?", batch).Delete(&entity.DictData{}).Error; deleteErr != nil {
				s.logger.Error("删除字典数据失败：删除数据失败", zap.Int64s("data_ids", batch), zap.Error(deleteErr))
				return errcode.ErrDictDataDeleteFailed.WithErr(deleteErr)
			}
		}
		return nil
	})
}

// UpdateData 更新字典数据；设置默认项时锁定所属类型并替换原默认项。
func (s *Service) UpdateData(ctx context.Context, req *UpdateDataReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.UpdateData")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrDictDataUpdateReqNil
	}
	if req.ID <= 0 {
		return errcode.ErrDictDataIDInvalid
	}
	typeCode, label, value, validationErr := validateDataFields(req.TypeCode, req.Label, req.Value, req.Sort, req.Status)
	if validationErr != nil {
		return validationErr
	}
	if req.Status != nil && *req.Status == enum.StatusDisabled && req.IsDefault != nil && *req.IsDefault {
		return errcode.ErrDictDataDefaultDisabled
	}
	remark := normalizeOptionalString(req.Remark)

	for attempt := 0; attempt < 3; attempt++ {
		updateErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var snapshot entity.DictData
			if queryErr := tx.Select("id", "type_code").Where("id = ?", req.ID).Take(&snapshot).Error; queryErr != nil {
				if errors.Is(queryErr, gorm.ErrRecordNotFound) {
					return errcode.ErrDictDataNotFound
				}
				s.logger.Error("更新字典数据失败：查询数据类型失败", zap.Int64("data_id", req.ID), zap.Error(queryErr))
				return errcode.ErrDictDataQueryFailed.WithErr(queryErr)
			}

			codes := []string{snapshot.TypeCode, typeCode}
			if lockErr := s.lockTypesByCodes(tx, codes); lockErr != nil {
				return lockErr
			}

			var current entity.DictData
			if queryErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", req.ID).Take(&current).Error; queryErr != nil {
				if errors.Is(queryErr, gorm.ErrRecordNotFound) {
					return errcode.ErrDictDataNotFound
				}
				s.logger.Error("更新字典数据失败：锁定数据失败", zap.Int64("data_id", req.ID), zap.Error(queryErr))
				return errcode.ErrDictDataQueryFailed.WithErr(queryErr)
			}
			if current.TypeCode != snapshot.TypeCode {
				return errDictDataTypeChanged
			}
			if checkErr := s.validateDataValue(tx, req.ID, typeCode, value); checkErr != nil {
				return checkErr
			}

			sortValue, status, isDefault := ptr.Value(current.Sort), ptr.Value(current.Status), ptr.Value(current.IsDefault)
			if req.Sort != nil {
				sortValue = *req.Sort
			}
			if req.Status != nil {
				status = *req.Status
			}
			if req.IsDefault != nil {
				isDefault = *req.IsDefault
			}
			if req.IsDefault != nil && *req.IsDefault && status != enum.StatusEnabled {
				return errcode.ErrDictDataDefaultDisabled
			}
			if status != enum.StatusEnabled {
				isDefault = false
			}
			if isDefault {
				if defaultErr := tx.Model(&entity.DictData{}).
					Where("type_code = ? AND id <> ? AND is_default = ?", typeCode, req.ID, true).
					Update("is_default", false).Error; defaultErr != nil {
					s.logger.Error("更新字典数据失败：清除原默认项失败", zap.Int64("data_id", req.ID), zap.String("type_code", typeCode), zap.Error(defaultErr))
					return errcode.ErrDictDataUpdateFailed.WithErr(defaultErr)
				}
			}
			if writeErr := tx.Model(&entity.DictData{}).Where("id = ?", req.ID).Updates(map[string]any{
				"type_code": typeCode, "label": label, "dict_value": value, "sort": sortValue,
				"status": status, "is_default": isDefault, "remark": remark,
			}).Error; writeErr != nil {
				s.logger.Error("更新字典数据失败：写入数据失败", zap.Int64("data_id", req.ID), zap.String("type_code", typeCode), zap.Error(writeErr))
				return errcode.ErrDictDataUpdateFailed.WithErr(writeErr)
			}
			return nil
		})
		if errors.Is(updateErr, errDictDataTypeChanged) {
			continue
		}
		return updateErr
	}

	s.logger.Warn("更新字典数据失败：所属类型持续发生并发变化", zap.Int64("data_id", req.ID))
	return errcode.ErrDictDataUpdateFailed.WithErr(errDictDataTypeChanged)
}

// ListData 查询字典数据分页列表。
func (s *Service) ListData(ctx context.Context, req *ListDataReq) (resp *dto.PageResp[*DataListItemResp], err error) {
	ctx, span := tracer.Start(ctx, "system.dict.ListData")
	defer func() { tracing.FinishSpan(span, err) }()
	page, size := enum.DefaultPage, enum.DefaultSize
	var typeCode, label, value string
	var status *int
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
		typeCode, label, value, status = strings.TrimSpace(req.TypeCode), strings.TrimSpace(req.Label), strings.TrimSpace(req.Value), req.Status
		if status != nil && !enum.IsStatusValid(*status) {
			return nil, errcode.ErrDictDataStatusInvalid
		}
	}
	if size > enum.MaxSize {
		size = enum.MaxSize
	}
	query := s.db.WithContext(ctx).Table(entity.DictData{}.TableName()+" AS d").Where("d.deleted_at = ?", 0)
	if typeCode != "" {
		query = query.Where("d.type_code = ?", typeCode)
	}
	if label != "" {
		query = query.Where("d.label LIKE ?", "%"+label+"%")
	}
	if value != "" {
		query = query.Where("d.dict_value LIKE ?", "%"+value+"%")
	}
	if status != nil {
		query = query.Where("d.status = ?", *status)
	}
	var total int64
	if queryErr := query.Count(&total).Error; queryErr != nil {
		return nil, errcode.ErrDictDataListQueryFailed.WithErr(queryErr)
	}
	var items []*DataListItemResp
	if queryErr := query.Select("d.id, d.type_code, d.label, d.dict_value AS value, d.sort, d.status, d.is_default, d.remark, d.created_at, d.updated_at").
		Order("d.sort ASC, d.id ASC").Limit(size).Offset((page - 1) * size).Find(&items).Error; queryErr != nil {
		return nil, errcode.ErrDictDataListQueryFailed.WithErr(queryErr)
	}
	return dto.NewPageResp(items, total), nil
}

// DetailData 查询字典数据详情。
func (s *Service) DetailData(ctx context.Context, req *DetailDataReq) (resp *DetailDataResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dict.DetailData")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrDictDataDetailReqNil
	}
	if req.ID <= 0 {
		return nil, errcode.ErrDictDataIDInvalid
	}
	var detail DetailDataResp
	if queryErr := s.db.WithContext(ctx).Model(&entity.DictData{}).
		Select("id, type_code, label, dict_value AS value, sort, status, is_default, remark, created_at, updated_at").
		Where("id = ?", req.ID).Take(&detail).Error; queryErr != nil {
		if errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrDictDataNotFound
		}
		return nil, errcode.ErrDictDataQueryFailed.WithErr(queryErr)
	}
	return &detail, nil
}

// UpdateDataStatus 批量更新字典数据状态；禁用默认项时同步取消默认标记。
func (s *Service) UpdateDataStatus(ctx context.Context, req *UpdateDataStatusReq) (err error) {
	ctx, span := tracer.Start(ctx, "system.dict.UpdateDataStatus")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return errcode.ErrDictDataUpdateStatusReqNil
	}
	if len(req.IDs) == 0 {
		return errcode.ErrDictDataIDsRequired
	}
	if req.Status == nil || !enum.IsStatusValid(*req.Status) {
		return errcode.ErrDictDataStatusInvalid
	}
	ids, normalizeErr := normalizeIDs(req.IDs, errcode.ErrDictDataIDInvalid)
	if normalizeErr != nil {
		return normalizeErr
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(ids); start += enum.BatchSize {
			end := min(start+enum.BatchSize, len(ids))
			batch := ids[start:end]
			if _, lockErr := s.lockDataByIDs(tx, batch); lockErr != nil {
				return lockErr
			}
			updates := map[string]any{"status": *req.Status}
			if *req.Status == enum.StatusDisabled {
				updates["is_default"] = false
			}
			if updateErr := tx.Model(&entity.DictData{}).Where("id IN ?", batch).Updates(updates).Error; updateErr != nil {
				s.logger.Error("更新字典数据状态失败：写入状态失败", zap.Int64s("data_ids", batch), zap.Int("status", *req.Status), zap.Error(updateErr))
				return errcode.ErrDictDataUpdateStatusFailed.WithErr(updateErr)
			}
		}
		return nil
	})
}

// ListValues 按类型编码查询业务侧可用的启用字典值。
func (s *Service) ListValues(ctx context.Context, req *ListValuesReq) (resp []*ValueResp, err error) {
	ctx, span := tracer.Start(ctx, "system.dict.ListValues")
	defer func() { tracing.FinishSpan(span, err) }()
	if req == nil {
		return nil, errcode.ErrDictValuesReqNil
	}
	typeCode := strings.TrimSpace(req.TypeCode)
	if typeCode == "" {
		return nil, errcode.ErrDictDataTypeCodeRequired
	}
	returnValues := make([]*ValueResp, 0)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var dictType entity.DictType
		if queryErr := tx.Select("id", "status").Where("code = ?", typeCode).Take(&dictType).Error; queryErr != nil {
			if errors.Is(queryErr, gorm.ErrRecordNotFound) {
				return errcode.ErrDictDataTypeNotFound
			}
			return errcode.ErrDictDataTypeQueryFailed.WithErr(queryErr)
		}
		if dictType.Status == nil || *dictType.Status != enum.StatusEnabled {
			return errcode.ErrDictTypeDisabled
		}
		var rows []entity.DictData
		if queryErr := tx.Where("type_code = ? AND status = ?", typeCode, enum.StatusEnabled).
			Order("sort ASC, id ASC").Find(&rows).Error; queryErr != nil {
			return errcode.ErrDictValuesQueryFailed.WithErr(queryErr)
		}
		for i := range rows {
			returnValues = append(returnValues, &ValueResp{Label: rows[i].Label, Value: rows[i].Value, IsDefault: ptr.Value(rows[i].IsDefault)})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return returnValues, nil
}

func (s *Service) validateTypeUnique(tx *gorm.DB, id int64, name, code string) error {
	if err := s.validateTypeName(tx, id, name); err != nil {
		return err
	}
	query := tx.Model(&entity.DictType{}).Where("code = ?", code)
	if id > 0 {
		query = query.Where("id <> ?", id)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return errcode.ErrDictTypeCodeQueryFailed.WithErr(err)
	}
	if count > 0 {
		return errcode.ErrDictTypeCodeExists
	}
	return nil
}

func (s *Service) validateTypeName(tx *gorm.DB, id int64, name string) error {
	query := tx.Model(&entity.DictType{}).Where("name = ?", name)
	if id > 0 {
		query = query.Where("id <> ?", id)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return errcode.ErrDictTypeNameQueryFailed.WithErr(err)
	}
	if count > 0 {
		return errcode.ErrDictTypeNameExists
	}
	return nil
}

func (s *Service) validateDataValue(tx *gorm.DB, id int64, typeCode, value string) error {
	query := tx.Model(&entity.DictData{}).Where("type_code = ? AND dict_value = ?", typeCode, value)
	if id > 0 {
		query = query.Where("id <> ?", id)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return errcode.ErrDictDataValueQueryFailed.WithErr(err)
	}
	if count > 0 {
		return errcode.ErrDictDataValueExists
	}
	return nil
}

func (s *Service) lockTypesByIDs(tx *gorm.DB, ids []int64) ([]entity.DictType, error) {
	var rows []entity.DictType
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "code", "status").Where("id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, errcode.ErrDictTypeQueryFailed.WithErr(err)
	}
	if len(rows) != len(ids) {
		return nil, errcode.ErrDictTypeNotFound
	}
	return rows, nil
}

func (s *Service) lockTypeByCode(tx *gorm.DB, code string) (*entity.DictType, error) {
	var row entity.DictType
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "code", "status").Where("code = ?", code).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrDictDataTypeNotFound
		}
		return nil, errcode.ErrDictDataTypeQueryFailed.WithErr(err)
	}
	return &row, nil
}

func (s *Service) lockTypesByCodes(tx *gorm.DB, codes []string) error {
	seen := make(map[string]struct{}, len(codes))
	unique := make([]string, 0, len(codes))
	for _, code := range codes {
		if _, ok := seen[code]; !ok {
			seen[code] = struct{}{}
			unique = append(unique, code)
		}
	}
	var rows []entity.DictType
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "code").Where("code IN ?", unique).Order("code ASC").Find(&rows).Error; err != nil {
		return errcode.ErrDictDataTypeQueryFailed.WithErr(err)
	}
	if len(rows) != len(unique) {
		return errcode.ErrDictDataTypeNotFound
	}
	return nil
}

func (s *Service) lockDataByIDs(tx *gorm.DB, ids []int64) ([]entity.DictData, error) {
	var rows []entity.DictData
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "type_code", "is_default").Where("id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, errcode.ErrDictDataQueryFailed.WithErr(err)
	}
	if len(rows) != len(ids) {
		return nil, errcode.ErrDictDataNotFound
	}
	return rows, nil
}

func validateDataFields(typeCode, label, value string, sortValue, status *int) (string, string, string, error) {
	typeCode, label, value = strings.TrimSpace(typeCode), strings.TrimSpace(label), strings.TrimSpace(value)
	if typeCode == "" {
		return "", "", "", errcode.ErrDictDataTypeCodeRequired
	}
	if label == "" {
		return "", "", "", errcode.ErrDictDataLabelRequired
	}
	if value == "" {
		return "", "", "", errcode.ErrDictDataValueRequired
	}
	if sortValue != nil && *sortValue < 0 {
		return "", "", "", errcode.ErrDictDataSortInvalid
	}
	if status != nil && !enum.IsStatusValid(*status) {
		return "", "", "", errcode.ErrDictDataStatusInvalid
	}
	return typeCode, label, value, nil
}

func normalizeIDs(input []int64, invalidErr error) ([]int64, error) {
	ids := make([]int64, 0, len(input))
	seen := make(map[int64]struct{}, len(input))
	for _, id := range input {
		if id <= 0 {
			return nil, invalidErr
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
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
