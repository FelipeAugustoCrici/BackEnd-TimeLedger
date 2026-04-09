package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"controltasks/internal/db"
	"controltasks/internal/handler"
	"controltasks/internal/middleware"
	"controltasks/internal/repository"
	"controltasks/internal/service"
)

func main() {
	_ = godotenv.Load()

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}
	defer database.Close()
	log.Println("✓ Banco de dados conectado")

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Erro ao rodar migrations: %v", err)
	}

	// ─── Repositórios ─────────────────────────────────────────────────────────
	entryRepo    := repository.NewEntryRepository(database)
	settingsRepo := repository.NewSettingsRepository(database)
	authRepo     := repository.NewAuthRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	timerRepo    := repository.NewTimerRepository(database)

	// ─── Services ─────────────────────────────────────────────────────────────
	entrySvc    := service.NewEntryService(entryRepo)
	settingsSvc := service.NewSettingsService(settingsRepo)
	authSvc     := service.NewAuthService(authRepo)
	timerSvc    := service.NewTimerService(timerRepo)

	// ─── Handlers ─────────────────────────────────────────────────────────────
	entryH    := handler.NewEntryHandler(entrySvc, settingsSvc)
	settingsH := handler.NewSettingsHandler(settingsSvc)
	authH     := handler.NewAuthHandler(authSvc)
	categoryH := handler.NewCategoryHandler(categoryRepo)
	timerH    := handler.NewTimerHandler(timerSvc)

	// ─── Router ───────────────────────────────────────────────────────────────
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	// ─── Rotas públicas (sem autenticação) ────────────────────────────────────
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login",    authH.Login)

	// ─── Rotas protegidas ─────────────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.Auth(authSvc))
	{
		protected.POST("/auth/logout", authH.Logout)
		protected.GET("/auth/me",      authH.Me)

		protected.GET("/dashboard", entryH.Dashboard)

		entries := protected.Group("/entries")
		{
			entries.GET("",                   entryH.List)
			entries.POST("",                  entryH.Create)
			entries.POST("/apply-rate",       entryH.ApplyRate)
			entries.GET("/:id",               entryH.GetByID)
			entries.PUT("/:id",               entryH.Update)
			entries.DELETE("/:id",            entryH.Delete)
			entries.GET("/meta/projects",     entryH.ListProjects)
			entries.GET("/meta/categories",   entryH.ListCategories)
		}

		protected.GET("/settings", settingsH.Get)
		protected.PUT("/settings", settingsH.Update)

		categories := protected.Group("/categories")
		{
			categories.GET("",        categoryH.List)
			categories.POST("",       categoryH.Create)
			categories.DELETE("/:id", categoryH.Delete)
		}

		timer := protected.Group("/timer")
		{
			timer.GET("",       timerH.Get)
			timer.POST("",      timerH.Start)
			timer.PATCH("/pause",  timerH.Pause)
			timer.PATCH("/resume", timerH.Resume)
			timer.DELETE("",    timerH.Delete)
		}
	}

	// Railway injeta PORT; fallback para APP_PORT e depois 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	log.Printf("✓ API rodando em http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
