package handler

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"hypertensionstages/internal/app/auth"
	"hypertensionstages/internal/app/media"
	"net/http"
	"strconv"
	"strings"

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
	user := currentUser(c)
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
	user := currentUser(c)
	service, err := h.Repo.APIGetFeed(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feed": service})
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
	user := currentUser(c)

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
	user := currentUser(c)

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
	user := currentUser(c)

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

	user := currentUser(c)

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

	user := currentUser(c)

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

	user := currentUser(c)

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
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	body.Username = strings.TrimSpace(body.Username)
	if len(body.Username) < 3 || len(body.Password) < 8 {
		c.JSON(400, gin.H{"error": "username requires 3 characters, password requires 8"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "password hashing failed"})
		return
	}
	user, err := h.Repo.APIRegisterUser(body.Username, string(hash))
	if err != nil {
		c.JSON(409, gin.H{"error": "registration failed"})
		return
	}
	c.JSON(201, gin.H{"user": user})
}
func (h *APIHandler) Login(c *gin.Context) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	u, err := h.Repo.UserByUsername(body.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(body.Password)) != nil {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	token, claims, err := auth.Issue(u.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "token creation failed"})
		return
	}
	c.JSON(200, gin.H{"access_token": token, "token_type": "Bearer", "expires_at": claims.ExpiresAt})
}
func (h *APIHandler) Logout(c *gin.Context) {
	claims, ok := c.Get("jwt_claims")
	if !ok {
		c.JSON(401, gin.H{"error": "authentication required"})
		return
	}
	if err := auth.Revoke(c.Request.Context(), claims.(auth.Claims)); err != nil {
		c.JSON(503, gin.H{"error": "redis unavailable"})
		return
	}
	c.JSON(200, gin.H{"status": "logged_out"})
}

// Optional authentication for public endpoints, required authentication for mutations.
func AuthMiddleware(required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			if required {
				c.AbortWithStatusJSON(401, gin.H{"error": "Bearer token required"})
				return
			}
			c.Next()
			return
		}
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization header"})
			return
		}
		id, claims, err := auth.Verify(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token"})
			return
		}
		revoked, err := auth.Revoked(c.Request.Context(), claims.ID)
		if err != nil {
			c.AbortWithStatusJSON(503, gin.H{"error": "redis unavailable"})
			return
		}
		if revoked {
			c.AbortWithStatusJSON(401, gin.H{"error": "token revoked"})
			return
		}
		c.Set("current_user_id", id)
		c.Set("jwt_claims", claims)
		c.Next()
	}
}
func currentUser(c *gin.Context) *ds.User {
	v, ok := c.Get("current_user_id")
	if !ok {
		return &ds.User{}
	}
	return &ds.User{ID: v.(uint)}
}

// saveUploadedFile validates and uploads the file into MinIO.
func saveUploadedFile(c *gin.Context, field string) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return "", fmt.Errorf("file %s is missing", field)
	}
	return media.Upload(file, field)
}
