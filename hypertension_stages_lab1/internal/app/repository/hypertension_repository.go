package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"hypertensionstages/internal/app/ds"

	"gorm.io/gorm"
)

type HypertensionRepository struct {
	db *gorm.DB
}

func NewHypertensionRepository(db *gorm.DB) *HypertensionRepository {
	return &HypertensionRepository{
		db: db,
	}
}

// Получить все опубликованные карточки.
func (r *HypertensionRepository) PublishedHypertensionStages() ([]ds.HypertensionStage, error) {
	var services []ds.HypertensionStage

	err := r.db.
		Preload("Likes").
		Where("status = ?", ds.HypertensionPublished).
		Order("id ASC").
		Find(&services).
		Error

	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("опубликованные карточки стадий гипертонии не найдены")
	}

	return services, nil
}

// Получить опубликованную карточку по ID.
// Удаленные и черновые карточки через URL открыть нельзя.
func (r *HypertensionRepository) HypertensionStageByID(id int) (ds.HypertensionStage, error) {
	var service ds.HypertensionStage

	err := r.db.
		Preload("Likes").
		Where("id = ? AND status = ?", id, ds.HypertensionPublished).
		First(&service).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.HypertensionStage{}, fmt.Errorf(
			"опубликованная карточка стадии гипертонии с id=%d не найдена",
			id,
		)
	}

	if err != nil {
		return ds.HypertensionStage{}, err
	}

	return service, nil
}

// GET черновика ничего не создает в БД.
func (r *HypertensionRepository) HypertensionDraftService() (ds.HypertensionStage, error) {
	var stage ds.HypertensionStage
	user := ds.GetCurrentUser()
	err := r.db.Where("user_id = ? AND status = ?", user.ID, ds.HypertensionDraft).
		Order("id DESC").First(&stage).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.HypertensionStage{
			SystolicBP: 140, DiastolicBP: 90,
			ImageName: "draft_stage1.png", VideoName: "draft_stage1.mp4",
		}, nil
	}
	return stage, err
}

// POST создания новой карточки через ORM.
func (r *HypertensionRepository) CreateHypertensionDraft(description string, sbp, dbp int) error {
	user := ds.GetCurrentUser()
	stage := ds.HypertensionStage{
		Title:       ds.HypertensionStageTitleBySBP(sbp),
		Description: strings.TrimSpace(description),
		SystolicBP:  sbp, DiastolicBP: dbp,
		Status: ds.HypertensionDraft,
		UserID: user.ID, ModeratorID: user.ID,
		ImageName: "draft_stage1.png", VideoName: "draft_stage1.mp4",
	}
	return r.db.Create(&stage).Error
}

// Опубликовать черновик через ORM.
func (r *HypertensionRepository) PublishHypertensionDraft(
	id int,
	description string,
	systolicBP int,
	diastolicBP int,
) error {
	updates := map[string]interface{}{
		"title":        ds.HypertensionStageTitleBySBP(systolicBP),
		"description":  strings.TrimSpace(description),
		"systolic_bp":  systolicBP,
		"diastolic_bp": diastolicBP,
		"status":       ds.HypertensionPublished,
	}

	result := r.db.
		Model(&ds.HypertensionStage{}).
		Where("id = ? AND status = ?", id, ds.HypertensionDraft).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("черновик с id=%d не найден", id)
	}

	return nil
}

func (r *HypertensionRepository) DeleteHypertensionStageSQL(
	ctx context.Context,
	id int,
) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	result, err := sqlDB.ExecContext(
		ctx,
		`UPDATE hypertension_stages
		 SET status = $1
		 WHERE id = $2 AND status = $3`,
		ds.HypertensionDeleted,
		id,
		ds.HypertensionPublished,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("опубликованная карточка с id=%d не найдена", id)
	}

	return nil
}

// Следующая карточка: запрос к БД сразу ограничен одной строкой.
func (r *HypertensionRepository) NextPublishedHypertensionStage(currentID int) (ds.HypertensionStage, error) {
	var stage ds.HypertensionStage
	query := r.db.Preload("Likes").Where("status = ? AND id > ?", ds.HypertensionPublished, currentID).
		Order("id ASC").Limit(1).Find(&stage)
	if query.Error != nil {
		return stage, query.Error
	}
	if query.RowsAffected > 0 {
		return stage, nil
	}

	// Переход в начало ленты, если текущая карточка последняя.
	query = r.db.Preload("Likes").Where("status = ?", ds.HypertensionPublished).
		Order("id ASC").Limit(1).Find(&stage)
	if query.Error != nil {
		return stage, query.Error
	}
	if query.RowsAffected == 0 {
		return stage, gorm.ErrRecordNotFound
	}
	return stage, nil
}

// Фильтр по систолическому давлению.
// Оставлен для обратной совместимости, если где-то ещё вызывается.
func (r *HypertensionRepository) FilterPublishedHypertensionBySBP(
	sbp int,
) ([]ds.HypertensionStage, error) {
	stage := ds.HypertensionStageNumberBySBP(sbp)

	min, max := hypertensionRange(stage)

	var services []ds.HypertensionStage

	err := r.db.
		Preload("Likes").
		Where(
			"status = ? AND systolic_bp BETWEEN ? AND ?",
			ds.HypertensionPublished,
			min,
			max,
		).
		Order("id ASC").
		Find(&services).
		Error

	if err != nil {
		return nil, err
	}

	return services, nil
}

// SBPBounds возвращает минимальное и максимальное систолическое давление
// среди опубликованных карточек. Нужно для границ шкалы двух слайдеров.
func (r *HypertensionRepository) SBPBounds() (int, int, error) {
	var bounds struct {
		Min int
		Max int
	}

	err := r.db.
		Model(&ds.HypertensionStage{}).
		Where("status = ?", ds.HypertensionPublished).
		Select("COALESCE(MIN(systolic_bp), 0) AS min, COALESCE(MAX(systolic_bp), 0) AS max").
		Scan(&bounds).
		Error

	if err != nil {
		return 0, 0, err
	}

	return bounds.Min, bounds.Max, nil
}

// FilterPublishedHypertensionBySBPRange возвращает опубликованные карточки,
// у которых SystolicBP входит в диапазон [min, max].
func (r *HypertensionRepository) FilterPublishedHypertensionBySBPRange(
	min, max int,
) ([]ds.HypertensionStage, error) {
	var services []ds.HypertensionStage

	err := r.db.
		Preload("Likes").
		Where(
			"status = ? AND systolic_bp BETWEEN ? AND ?",
			ds.HypertensionPublished,
			min,
			max,
		).
		Order("systolic_bp ASC").
		Find(&services).
		Error

	if err != nil {
		return nil, err
	}

	return services, nil
}

// Диапазон давления для SQL.
func hypertensionRange(stage int) (int, int) {
	switch stage {
	case -2:
		return 0, 119
	case -1:
		return 120, 129
	case 0:
		return 130, 139
	case 1:
		return 140, 159
	case 2:
		return 160, 179
	default:
		return 180, 300
	}
}

// ============================================================
// API-МЕТОДЫ
// ============================================================

// APIGetStages — опубликованные стадии с признаком авторства текущего пользователя.
func (r *HypertensionRepository) APIGetStages(
	userID uint, filter string,
) ([]ds.HypertensionStage, error) {
	var stages []ds.HypertensionStage

	q := r.db.Where("status = ?", ds.HypertensionPublished).Order("id ASC")
	if filter != "" {
		q = q.Where("title ILIKE ?", "%"+filter+"%")
	}
	if err := q.Find(&stages).Error; err != nil {
		return nil, err
	}

	for i := range stages {
		value := 0
		if stages[i].UserID == userID {
			value = 1
		}
		stages[i].IsOwner = &value
	}
	return stages, nil
}

// APIGetFeed — опубликованные стадии с признаком активного лайка пользователя.
func (r *HypertensionRepository) APIGetFeed(userID uint) ([]ds.HypertensionStage, error) {
	var stages []ds.HypertensionStage
	err := r.db.
		Preload("Likes", "user_id = ? AND value = ?", userID, 1).
		Where("status = ?", ds.HypertensionPublished).
		Order("id ASC").
		Find(&stages).Error
	if err != nil {
		return nil, err
	}
	for i := range stages {
		value := 0
		if len(stages[i].Likes) > 0 {
			value = 1
		}
		stages[i].IsLiked = &value
	}
	return stages, nil
}

// APIPublishedByID — опубликованная стадия по ID с признаком лайка.
func (r *HypertensionRepository) APIPublishedByID(
	id int, userID uint,
) (ds.HypertensionStage, error) {
	var stage ds.HypertensionStage
	err := r.db.
		Preload("Likes", "user_id = ? AND value = ?", userID, 1).
		Where("id = ? AND status = ?", id, ds.HypertensionPublished).
		First(&stage).Error
	if err != nil {
		return stage, err
	}
	value := 0
	if len(stage.Likes) > 0 {
		value = 1
	}
	stage.IsLiked = &value
	return stage, nil
}

// APINextPublished — одна опубликованная карточка из БД и признак лайка.
func (r *HypertensionRepository) APINextPublished(id int, userID uint) (ds.HypertensionStage, error) {
	var stage ds.HypertensionStage
	query := r.db.Preload("Likes", "user_id = ? AND value = ?", userID, 1).
		Where("status = ? AND id > ?", ds.HypertensionPublished, id).
		Order("id ASC").Limit(1).Find(&stage)
	if query.Error != nil {
		return stage, query.Error
	}
	if query.RowsAffected == 0 {
		query = r.db.Preload("Likes", "user_id = ? AND value = ?", userID, 1).
			Where("status = ?", ds.HypertensionPublished).
			Order("id ASC").Limit(1).Find(&stage)
		if query.Error != nil {
			return stage, query.Error
		}
		if query.RowsAffected == 0 {
			return stage, gorm.ErrRecordNotFound
		}
	}
	liked := 0
	if len(stage.Likes) > 0 {
		liked = 1
	}
	stage.IsLiked = &liked
	return stage, nil
}

// APIGetDraft — черновик пользователя.
func (r *HypertensionRepository) APIGetDraft(
	userID uint,
) (ds.HypertensionStage, error) {
	var s ds.HypertensionStage
	err := r.db.
		Where("user_id = ? AND status = ?", userID, ds.HypertensionDraft).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.HypertensionStage{}, nil
	}
	return s, err
}

// APICreateStage — создание услуги.
//
// ВАЖНО: moderator_id обязателен по методичке, и на него есть FK
// на таблицу users. На этапе лабы модератором ставим самого создателя.
func (r *HypertensionRepository) APICreateStage(
	userID uint,
	title, description string,
	sbp, dbp int,
	imageName, videoName string,
) (ds.HypertensionStage, error) {

	s := ds.HypertensionStage{
		Title:       title,
		Description: description,
		SystolicBP:  sbp,
		DiastolicBP: dbp,
		Status:      ds.HypertensionDraft,
		UserID:      userID,
		ModeratorID: userID,
		ImageName:   imageName,
		VideoName:   videoName,
	}

	if err := r.db.Create(&s).Error; err != nil {
		return ds.HypertensionStage{}, err
	}
	return s, nil
}

// APIPublishStage — публикация черновика пользователя.
func (r *HypertensionRepository) APIPublishStage(
	id int, userID uint,
) (ds.HypertensionStage, error) {

	now := time.Now()

	result := r.db.
		Model(&ds.HypertensionStage{}).
		Where("id = ? AND user_id = ? AND status = ?",
			id, userID, ds.HypertensionDraft).
		Updates(map[string]interface{}{
			"status":    ds.HypertensionPublished,
			"formed_at": now,
		})

	if result.Error != nil {
		return ds.HypertensionStage{}, result.Error
	}
	if result.RowsAffected == 0 {
		return ds.HypertensionStage{}, fmt.Errorf("черновик не найден или не ваш")
	}

	var s ds.HypertensionStage
	r.db.First(&s, id)
	return s, nil
}

// APISoftDeleteStage — soft delete только своих.
func (r *HypertensionRepository) APISoftDeleteStage(
	id int, userID uint,
) error {

	result := r.db.
		Model(&ds.HypertensionStage{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("status", ds.HypertensionDeleted)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("услуга не найдена или не ваша")
	}
	return nil
}

// APIToggleLike — поставить/снять лайк.
func (r *HypertensionRepository) APIToggleLike(
	stageID int, userID uint, value int,
) (ds.HypertensionLike, error) {

	var like ds.HypertensionLike
	err := r.db.
		Where("user_id = ? AND hypertension_stage_id = ?", userID, stageID).
		First(&like).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		like = ds.HypertensionLike{
			UserID:              userID,
			HypertensionStageID: uint(stageID),
			Value:               value,
		}
		return like, r.db.Create(&like).Error
	}
	if err != nil {
		return ds.HypertensionLike{}, err
	}

	like.Value = value
	return like, r.db.Save(&like).Error
}

// APIRegisterUser — регистрация.
func (r *HypertensionRepository) APIRegisterUser(
	username, password string,
) (ds.User, error) {

	user := ds.User{Username: username, Password: password}
	if err := r.db.Create(&user).Error; err != nil {
		return ds.User{}, err
	}
	return user, nil
}
