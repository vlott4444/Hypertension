package main

import (
	"hypertensionstages/internal/app/dsn"
	"hypertensionstages/internal/app/handler"
	"hypertensionstages/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	_ = godotenv.Load()

	db, err := gorm.Open(
		postgres.Open(dsn.FromEnv()),
		&gorm.Config{},
	)
	if err != nil {
		panic("failed to connect database")
	}

	hypertensionRepository := repository.NewHypertensionRepository(db)

	hypertensionHandler := handler.NewHypertensionHandler(hypertensionRepository)

	apiHandler := handler.NewAPIHandler(hypertensionRepository)

	router := gin.Default()

	router.LoadHTMLGlob("./templates/*")
	router.Static("/resources", "./resources")

	router.GET("/feed/*stageID", hypertensionHandler.ShowHypertensionFeed)
	router.GET("/draft", hypertensionHandler.ShowHypertensionDraft)
	router.POST("/draft/publish", hypertensionHandler.PublishHypertensionDraft)
	router.GET("/stages", hypertensionHandler.ShowHypertensionGrid)
	router.POST("/stages/:stageID/delete", hypertensionHandler.DeleteHypertensionStage)

	api := router.Group("/api")
	{

		api.GET("/hypertension_stages", apiHandler.GetStages)
		api.GET("/feed", apiHandler.GetFeed)
		api.GET("/feed/:id", apiHandler.GetFeedByID)
		api.GET("/draft", apiHandler.GetDraft)
		api.POST("/hypertension_stages", apiHandler.CreateStage)
		api.PUT("/hypertension_stages/:id/publish", apiHandler.PublishStage)
		api.DELETE("/hypertension_stages/:id", apiHandler.DeleteStage)
		api.POST("/hypertension_stages/:id/like", apiHandler.LikeStage)

		api.POST("/auth/register", apiHandler.Register)
		api.POST("/auth/login", apiHandler.Login)
		api.POST("/auth/logout", apiHandler.Logout)
	}

	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
