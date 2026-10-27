package api

import (
	"log"

	"hypertensionstages/internal/app/handler"
	"hypertensionstages/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func StartHypertensionServer(db *gorm.DB) {
	log.Println("Starting hypertension stages server")

	repo := repository.NewHypertensionRepository(db)

	// HTML-фронт из лаб 1-3
	htmlHandler := handler.NewHypertensionHandler(repo)

	// REST API из лабы 4
	apiHandler := handler.NewAPIHandler(repo)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/resources", "./resources")

	// ---------- HTML-роуты (не трогаем) ----------
	router.GET("/feed/*stageID", htmlHandler.ShowHypertensionFeed)
	router.GET("/draft", htmlHandler.ShowHypertensionDraft)
	router.GET("/stages", htmlHandler.ShowHypertensionGrid)
	router.POST("/draft/publish", htmlHandler.PublishHypertensionDraft)
	router.POST("/stages/:stageID/delete", htmlHandler.DeleteHypertensionStage)

	// ---------- REST API (лаба 4) ----------
	api := router.Group("/api")
	{
		// Домен стадии гипертонии
		api.GET("/hypertension_stages", apiHandler.GetStages)
		api.GET("/feed", apiHandler.GetFeed)
		api.GET("/feed/:id", apiHandler.GetFeedByID)
		api.GET("/draft", apiHandler.GetDraft)
		api.POST("/hypertension_stages", apiHandler.CreateStage)
		api.PUT("/hypertension_stages/:id/publish", apiHandler.PublishStage)
		api.DELETE("/hypertension_stages/:id", apiHandler.DeleteStage)
		api.POST("/hypertension_stages/:id/like", apiHandler.LikeStage)

		// Домен пользователя
		api.POST("/auth/register", apiHandler.Register)
		api.POST("/auth/login", apiHandler.Login)
		api.POST("/auth/logout", apiHandler.Logout)
	}

	if err := router.Run(":8080"); err != nil {
		logrus.Fatal("сервер остановлен с ошибкой: ", err)
	}
}
