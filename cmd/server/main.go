// cmd/server/main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Axontik/comin-organization-service/internal/handler"
	"github.com/Axontik/comin-organization-service/internal/middleware"
	"github.com/Axontik/comin-organization-service/internal/repository"
	"github.com/Axontik/comin-organization-service/internal/service"
	"github.com/Axontik/comin-organization-service/pkg/auth"
)

type Application struct {
	db          *gorm.DB
	orgHandler  *handler.OrganizationHandler
	depHandler  *handler.DepartmentHandler
	teamHandler *handler.TeamHandler
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	app := &Application{}

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	app.db = db

	// Initialize dependencies
	app.initializeDependencies()

	// Setup router
	router := setupRouter(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDB() (*gorm.DB, error) {
	dbURL := "postgresql://comin_owner:Ye5rfjcIB7FX@ep-flat-shadow-a8onelva.eastus2.azure.neon.tech/comin?sslmode=require"
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// Run migrations
	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Printf("Warning: Failed to initialize migrations: %v", err)
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}
	}

	return gorm.Open(postgres.Open(dbURL), &gorm.Config{})
}

func (app *Application) initializeDependencies() {
	// Initialize repositories
	orgRepo := repository.NewOrganizationRepository(app.db)
	depRepo := repository.NewDepartmentRepository(app.db)
	teamRepo := repository.NewTeamRepository(app.db)

	// Initialize services
	orgService := service.NewOrganizationService(orgRepo)
	deptService := service.NewDepartmentService(depRepo)
	teamService := service.NewTeamService(teamRepo)

	// Initialize handlers
	app.orgHandler = handler.NewOrganizationHandler(orgService)
	app.depHandler = handler.NewDepartmentHandler(deptService)
	app.teamHandler = handler.NewTeamHandler(teamService)
}

func setupRouter(app *Application) *gin.Engine {
	authClient := auth.NewAuthClient("https://comin.kaveeshagimhana.com/api/v1/auth")
	authMiddleware := auth.NewAuthMiddleware(authClient)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		orgs := api.Group("/organizations")
		orgs.Use(authMiddleware.RequireOrganizationAccess)
		{
			orgs.POST("/", app.orgHandler.Create)
			orgs.GET("/:org_id", app.orgHandler.GetByID)
			orgs.PUT("/:org_id", app.orgHandler.Update)
			orgs.DELETE("/:org_id", app.orgHandler.Delete)
			orgs.GET("/", app.orgHandler.List)

			// Settings endpoints
			// orgs.PUT("/:id/settings", app.handler.UpdateSettings)
			// orgs.PUT("/:id/status", app.handler.UpdateStatus)

			// Department endpoints can be added here
			departments := orgs.Group("/:org_id/departments")
			{
				departments.POST("/", app.depHandler.Create)
				departments.GET("/:dept_id", app.depHandler.GetByID)
				departments.PUT("/:dept_id", app.depHandler.Update)
				departments.DELETE("/:dept_id", app.depHandler.Delete)
				departments.GET("/", app.depHandler.List)
			}

			// Team endpoints can be added here
			teams := orgs.Group("/:org_id/teams")
			{
				teams.POST("/", app.teamHandler.CreateTeam)
				teams.GET("/:team_id", app.teamHandler.GetTeamByID)
				teams.PUT("/:team_id", app.teamHandler.UpdateTeam)
				teams.DELETE("/:team_id", app.teamHandler.DeleteTeam)
				teams.GET("/", app.teamHandler.List)

				// Team member management
				teams.POST("/:team_id/members", app.teamHandler.AddMember)
				teams.DELETE("/:team_id/members/:user_id", app.teamHandler.DeleteMember)
				teams.GET("/:team_id/members", app.teamHandler.ListMembers)

				// Team lead management
				teams.PUT("/:team_id/lead", app.teamHandler.UpdateTeamLead)
			}
		}
	}

	return router
}
