package main

import (
	"fmt"
	stdLog "log"
	"os"
	"time"

	"qc_api/internal/auth"
	"qc_api/internal/calibration"
	"qc_api/internal/config"
	"qc_api/internal/employees"
	"qc_api/internal/inspections"
	"qc_api/internal/jobqueue"
	"qc_api/internal/utils"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "qc_api/docs"
)

// @title QC API
// @version 1.0
// @description Quality Control API for lawn care calibration and inspection management
// @termsOfService http://swagger.io/terms/

// @contact.name Caleb McNevin
// @contact.email caleb@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host api.calebm.ddns.net
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

type UploadResponse struct {
	Status   string `json:"status"`
	Filename string `json:"filename"`
}

func InitDB(ConnectionString string) *gorm.DB {
	// Create logger
	newLogger := logger.New(
		stdLog.New(os.Stdout, "\r\n", stdLog.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	// Connect
	db, err := gorm.Open(sqlite.Open(ConnectionString), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate
	// Collect models
	models := []interface{}{}
	models = append(models, auth.Models()...)
	models = append(models, employees.Models()...)
	models = append(models, inspections.Models()...)
	models = append(models, calibration.Models()...)

	// Migrate all
	if err := db.AutoMigrate(models...); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	return db
}

// DelayMiddleware delays each request and logs the path
func DelayMiddleware(delay time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			c.Logger().Infof("-> %s %s", c.Request().Method, path)
			time.Sleep(delay)
			return next(c)
		}
	}
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found, using system environment variables")
	}

	cfg := config.NewConfig()

	utils.GetMotiveUser("dishant.patel", cfg)

	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/db/qc_api.db"
	}
	db := InitDB(dbPath)

	authService := auth.NewAuthService(db, cfg.JWTSecret)
	employeeService := employees.NewEmployeeService(db)
	inspectionService := inspections.NewInspectionService(db)
	calibrationService := calibration.NewCalibrationService(db)

	go jobqueue.Worker()

	// Get delay from environment or use default
	delayMs := os.Getenv("DELAY_MS")
	delay := 100 * time.Millisecond
	if delayMs != "" {
		if parsedDelay, err := time.ParseDuration(delayMs + "ms"); err == nil {
			delay = parsedDelay
		}
	}

	// Get motive API key
	motiveKey := os.Getenv("MOTIVE_KEY")
	if motiveKey == "" {
		log.Warn("No Motive API Key specified")
	}

	e := echo.New()
	e.Validator = utils.NewValidator()

	e.Logger.SetLevel(log.INFO)

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins: []string{"https://dev.calebm.ddns.net", "https://api.calebm.ddns.net"},
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.Logger())

	e.Use(DelayMiddleware(delay))

	protected := e.Group("", authService.AuthMiddleware)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Register module routes
	auth.RegisterRoutes(e, authService)
	employees.RegisterRoutes(protected, employeeService)
	inspections.RegisterRoutes(protected, inspectionService)
	calibration.RegisterRoutes(protected, calibrationService)

	// e.POST("/upload", authService.AuthMiddleware(uploadHandler))
	// e.GET("/uploads", authService.AuthMiddleware(updloadsHandler))
	// http.Handle("/api/uploads", http.StripPrefix("/api/uploads", http.FileServer(http.Dir("uploads"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "2847"
	}

	fmt.Printf("API server running at http://0.0.0.0:%s\n", port)
	e.Logger.Fatal(e.Start(":" + port))
}
