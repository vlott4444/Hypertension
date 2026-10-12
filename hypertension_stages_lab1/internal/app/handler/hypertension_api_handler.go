package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"hypertensionstages/internal/app/ds"
	"hypertensionstages/internal/app/repository"

	"github.com/gin-gonic/gin"
)

// APIHandler — REST-хендлер для /api/*.
// Всё общение — только JSON, без HTML-шаблонов.
type APIHandler struct {
	Repo *repository.HypertensionRepository
}

func NewAPIHandler(repo *repository.HypertensionRepository) *APIHandler {
	return &APIHandler{Repo: repo}
}

// ============================================================
// ДОМЕН СТАДИИ ГИПЕРТОНИИ
// ============================================================

// GET /api/hypertension_stages?filter=...
// Список опубликованных с фильтрацией.
// Поле "is_owner" 0/1 — если создатель совпадает с текущим пользователем.
func (h *APIHandler) GetStages(c *gin.Context) {
	user := ds.GetCurrentUser()
	filter := c.Query("filter")

	services, err := h.Repo.APIGetStages(user.ID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hypertension_stages": services})
}

// GET /api/feed
// Лента — только опубликованные.
func (h *APIHandler) GetFeed(c *gin.Context) {
	user := ds.GetCurrentUser()
	services, err := h.Repo.APIGetFeed(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feed": services})
}

// GET /api/feed/:id?next=true
// Публикация по ID. Если next=true — следующая по ленте.
func (h *APIHandler) GetFeedByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	next := c.Query("next") == "true"
	user := ds.GetCurrentUser()

	var service ds.HypertensionStage
	if next {
		service, err = h.Repo.APINextPublished(id, user.ID)
	} else {
		service, err = h.Repo.APIPublishedByID(id, user.ID)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hypertension_stage": service})
}

// GET /api/draft
// Черновик текущего пользователя. Не более одного. ID не указывается.
func (h *APIHandler) GetDraft(c *gin.Context) {
	user := ds.GetCurrentUser()

	service, err := h.Repo.APIGetDraft(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"draft": service})
}

// POST /api/hypertension_stages
// Создание стадии гипертонии с загрузкой файлов (multipart/form-data).
// Название стадии, описание, САД, ДАД — из формы.
// Файлы: image, video.
// Системные поля (ID, статус, создатель, даты) с клиента игнорируются.
func (h *APIHandler) CreateStage(c *gin.Context) {
	user := ds.GetCurrentUser()

	title := c.PostForm("title")
	description := c.PostForm("description")
	sbp, _ := strconv.Atoi(c.PostForm("systolic_bp"))
	dbp, _ := strconv.Atoi(c.PostForm("diastolic_bp"))

	imageName, err := saveUploadedFile(c, "image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	videoName, err := saveUploadedFile(c, "video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.Repo.APICreateStage(
		user.ID, title, description, sbp, dbp, imageName, videoName,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"hypertension_stage": service})
}

// PUT /api/hypertension_stages/:id/publish
// Публикация черновика. Только свои услуги.
func (h *APIHandler) PublishStage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user := ds.GetCurrentUser()

	service, err := h.Repo.APIPublishStage(id, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"hypertension_stage": service})
}

// DELETE /api/hypertension_stages/:id
// Soft delete. Только свои услуги.
func (h *APIHandler) DeleteStage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	user := ds.GetCurrentUser()

	if err := h.Repo.APISoftDeleteStage(id, user.ID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// POST /api/hypertension_stages/:id/like
// Тело: {"value": 0|1}
func (h *APIHandler) LikeStage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body struct {
		Value int `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Value != 0 && body.Value != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be 0 or 1"})
		return
	}

	user := ds.GetCurrentUser()

	like, err := h.Repo.APIToggleLike(id, user.ID, body.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"like": like})
}

// ============================================================
// ДОМЕН ПОЛЬЗОВАТЕЛЯ
// ============================================================

// POST /api/auth/register
func (h *APIHandler) Register(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	user, err := h.Repo.APIRegisterUser(body.Username, body.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// POST /api/auth/login — заглушка
func (h *APIHandler) Login(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "stub"})
}

// POST /api/auth/logout — заглушка
func (h *APIHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "stub"})
}

// saveUploadedFile принимает файл из формы, генерирует латинское имя
// для последующей загрузки в MinIO. Возвращает имя файла (для сохранения в БД).
func saveUploadedFile(c *gin.Context, field string) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return "", fmt.Errorf("файл %s не передан", field)
	}

	// TODO: здесь подключишь Minio-клиент для пользовательских файлов
	// и загрузишь файл. Пока просто возвращаем имя.
	return file.Filename, nil
}
