package main

import (
	"hypertensionstages/internal/app/dsn"
	"os"

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
	router.POST("/draft/create", handler.AuthMiddleware(true), hypertensionHandler.CreateHypertensionDraft)
	router.POST("/draft/publish", handler.AuthMiddleware(true), hypertensionHandler.PublishHypertensionDraft)
	router.GET("/stages", hypertensionHandler.ShowHypertensionGrid)
	router.POST("/stages/:stageID/delete", handler.AuthMiddleware(true), hypertensionHandler.DeleteHypertensionStage)

	router.GET("/swagger", func(c *gin.Context) { c.Header("Content-Type", "text/html; charset=utf-8"); c.String(200, swaggerHTML) })
	router.GET("/swagger/openapi.json", func(c *gin.Context) { c.File("./docs/openapi.json") })

	api := router.Group("/api")
	{

		api.GET("/hypertension_stages", handler.AuthMiddleware(false), apiHandler.GetStages)
		api.GET("/feed", handler.AuthMiddleware(false), apiHandler.GetFeed)
		api.GET("/feed/:id", handler.AuthMiddleware(false), apiHandler.GetFeedByID)
		api.GET("/draft", handler.AuthMiddleware(true), apiHandler.GetDraft)
		api.POST("/hypertension_stages", handler.AuthMiddleware(true), apiHandler.CreateStage)
		api.PUT("/hypertension_stages/:id/publish", handler.AuthMiddleware(true), apiHandler.PublishStage)
		api.DELETE("/hypertension_stages/:id", handler.AuthMiddleware(true), apiHandler.DeleteStage)
		api.POST("/hypertension_stages/:id/like", handler.AuthMiddleware(true), apiHandler.LikeStage)

		api.POST("/auth/register", apiHandler.Register)
		api.POST("/auth/login", apiHandler.Login)
		api.POST("/auth/logout", handler.AuthMiddleware(true), apiHandler.Logout)
	}

	if os.Getenv("JWT_SECRET") == "" {
		panic("JWT_SECRET is required; copy .env.lab4.example to .env and configure it")
	}
	err = router.Run(":8080")
	if err != nil {
		panic(err)
	}
}

const swaggerHTML = `<!DOCTYPE html><html><head><meta charset="utf-8"><title>Hypertension Stages API</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/swagger/openapi.json',dom_id:'#swagger-ui',persistAuthorization:true})</script></body></html>`
