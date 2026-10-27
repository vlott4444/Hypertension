package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"hypertensionstages/internal/app/ds"
	"hypertensionstages/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type HypertensionHandler struct {
	HypertensionRepository *repository.HypertensionRepository
}

type HypertensionCardView struct {
	Service         ds.HypertensionStage
	StageTitle      string
	StageRomanTitle string
	StageRange      string
	LikesCount      int
}

func NewHypertensionHandler(
	hypertensionRepository *repository.HypertensionRepository,
) *HypertensionHandler {
	return &HypertensionHandler{
		HypertensionRepository: hypertensionRepository,
	}
}

// GET — лента
func (h *HypertensionHandler) ShowHypertensionFeed(ctx *gin.Context) {
	service, err := h.getHypertensionStageForFeed(ctx)
	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "hypertension_feed.html", gin.H{
		"Card": h.BuildHypertensionCardView(service),
	})
}

// GET — черновик
func (h *HypertensionHandler) ShowHypertensionDraft(ctx *gin.Context) {
	draft, err := h.HypertensionRepository.HypertensionDraftService()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "hypertension_draft.html", gin.H{
		"Card": h.BuildHypertensionCardView(draft),
	})
}

// POST — публикация через ORM
func (h *HypertensionHandler) PublishHypertensionDraft(ctx *gin.Context) {
	stageID, systolicBP, diastolicBP, err := parseDraftForm(ctx)
	if err != nil {
		ctx.String(http.StatusBadRequest, "некорректные данные")
		return
	}

	err = h.HypertensionRepository.PublishHypertensionDraft(
		stageID,
		ctx.PostForm("description"),
		systolicBP,
		diastolicBP,
	)

	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/stages")
}

// GET — плитка и поиск
func (h *HypertensionHandler) ShowHypertensionGrid(ctx *gin.Context) {
	services, boundsMin, boundsMax, sbpMin, sbpMax, filterError, err :=
		h.getHypertensionStagesForGrid(ctx)

	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	cards := make([]HypertensionCardView, 0, len(services))

	for _, service := range services {
		cards = append(cards, h.BuildHypertensionCardView(service))
	}

	ctx.HTML(http.StatusOK, "hypertension_grid.html", gin.H{
		"Cards": cards,
		"SBPBounds": gin.H{
			"Min": boundsMin,
			"Max": boundsMax,
		},
		"SBPMin":      sbpMin,
		"SBPMax":      sbpMax,
		"SBPQuery":    fmt.Sprintf("%d–%d", sbpMin, sbpMax),
		"FilterError": filterError,
	})
}

// POST — удаление чистый SQL
func (h *HypertensionHandler) DeleteHypertensionStage(ctx *gin.Context) {
	stageID, err := strconv.Atoi(ctx.Param("stageID"))
	if err != nil {
		ctx.String(http.StatusBadRequest, "некорректный id")
		return
	}

	err = h.HypertensionRepository.DeleteHypertensionStageSQL(
		ctx.Request.Context(),
		stageID,
	)

	if err != nil {
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/stages")
}

// --------------------
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// --------------------

func (h *HypertensionHandler) getHypertensionStageForFeed(
	ctx *gin.Context,
) (ds.HypertensionStage, error) {

	rawID := strings.Trim(ctx.Param("stageID"), "/")

	if rawID == "" {
		services, err :=
			h.HypertensionRepository.PublishedHypertensionStages()

		if err != nil {
			return ds.HypertensionStage{}, err
		}

		return services[0], nil
	}

	stageID, err := strconv.Atoi(rawID)
	if err != nil {
		return ds.HypertensionStage{}, err
	}

	if ctx.Query("next") == "true" {
		return h.HypertensionRepository.
			NextPublishedHypertensionStage(stageID)
	}

	return h.HypertensionRepository.
		HypertensionStageByID(stageID)
}

func parseDraftForm(
	ctx *gin.Context,
) (int, int, int, error) {

	stageID, err := strconv.Atoi(ctx.PostForm("service_id"))
	if err != nil {
		return 0, 0, 0, err
	}

	systolicBP, err := strconv.Atoi(ctx.PostForm("sbp"))
	if err != nil {
		return 0, 0, 0, err
	}

	diastolicBP, err := strconv.Atoi(ctx.PostForm("dbp"))
	if err != nil {
		return 0, 0, 0, err
	}

	return stageID, systolicBP, diastolicBP, nil
}

// getHypertensionStagesForGrid возвращает карточки, отфильтрованные
// по диапазону [sbp_min, sbp_max], а также границы шкалы и текущие
// значения ползунков для шаблона.
func (h *HypertensionHandler) getHypertensionStagesForGrid(
	ctx *gin.Context,
) (
	services []ds.HypertensionStage,
	boundsMin, boundsMax int,
	sbpMin, sbpMax int,
	filterError string,
	err error,
) {
	// 1. Границы шкалы — минимум и максимум САД по всем
	//    опубликованным карточкам.
	boundsMin, boundsMax, err = h.HypertensionRepository.SBPBounds()
	if err != nil {
		return nil, 0, 0, 0, 0, "", err
	}

	if boundsMin == 0 && boundsMax == 0 {
		boundsMin, boundsMax = 100, 200
	}

	// 2. Текущие значения ползунков. По умолчанию — границы шкалы.
	sbpMin = boundsMin
	sbpMax = boundsMax

	if raw := strings.TrimSpace(ctx.Query("sbp_min")); raw != "" {
		v, convErr := strconv.Atoi(raw)
		if convErr != nil {
			return nil, boundsMin, boundsMax, boundsMin, boundsMax,
				"Минимальное САД должно быть целым числом", nil
		}
		sbpMin = v
	}

	if raw := strings.TrimSpace(ctx.Query("sbp_max")); raw != "" {
		v, convErr := strconv.Atoi(raw)
		if convErr != nil {
			return nil, boundsMin, boundsMax, boundsMin, boundsMax,
				"Максимальное САД должно быть целым числом", nil
		}
		sbpMax = v
	}

	// 3. Без JS ползунки могут перескочить друг через друга —
	//    просто меняем местами, чтобы фильтр остался корректным.
	if sbpMin > sbpMax {
		sbpMin, sbpMax = sbpMax, sbpMin
	}

	// 4. Фильтрация по диапазону.
	services, err = h.HypertensionRepository.
		FilterPublishedHypertensionBySBPRange(sbpMin, sbpMax)

	return services, boundsMin, boundsMax, sbpMin, sbpMax, "", err
}

func (h *HypertensionHandler) BuildHypertensionCardView(
	service ds.HypertensionStage,
) HypertensionCardView {

	return HypertensionCardView{
		Service: service,

		StageTitle: ds.HypertensionStageTitleBySBP(
			service.SystolicBP,
		),

		StageRomanTitle: ds.HypertensionStageRomanTitleBySBP(
			service.SystolicBP,
		),

		StageRange: ds.HypertensionStageRangeBySBP(
			service.SystolicBP,
		),

		LikesCount: len(service.Likes),
	}
}
