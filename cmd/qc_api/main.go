package main

import (
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"qc_api/internal/ReportGenerator"
	"qc_api/internal/auth"
	"qc_api/internal/calibration"
	"qc_api/internal/config"
	"qc_api/internal/employees"
	"qc_api/internal/inspections"
	"qc_api/internal/jobqueue"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	delay_ms time.Duration = 100 * time.Millisecond
)

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

// DelayMiddleware delays each request by 1 second and logs the path
func DelayMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Path()
		c.Logger().Infof("-> %s %s", c.Request().Method, path)
		time.Sleep(delay_ms)
		return next(c)
	}
}

func main() {

	cfg := config.NewConfig()
	db := InitDB("data/db/qc_api.db")

	authService := auth.NewAuthService(db, cfg.JWTSecret)
	employeeService := employees.NewEmployeeService(db)
	inspectionService := inspections.NewInspectionService(db)
	calibrationService := calibration.NewCalibrationService(db)

	go jobqueue.Worker()

	e := echo.New()

	e.Logger.SetLevel(log.INFO)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins: []string{"https://dev.calebm.ddns.net", "https://api.calebm.ddns.net"},
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.Use(DelayMiddleware)

	protected := e.Group("", authService.AuthMiddleware)

	e.POST("/login", authService.LoginHandler)
	e.POST("/register", authService.RegisterHandler)

	protected.POST("/employees", employeeService.CreateEmployeeHandler)
	protected.GET("/employees", employeeService.GetEmployeesHandler)
	protected.GET("/employees/:id", employeeService.GetEmployeeByIDHandler)
	protected.GET("/employees/inspections", inspectionService.GetByEmployeeHandler)
	protected.POST("/inspections", inspectionService.PostInspectionHandler)
	protected.GET("/inspections", inspectionService.GetInspectionsHandler)
	protected.GET("/inspections/:id", inspectionService.GetInspectionHandler)
	protected.PATCH("/inspections/:id", inspectionService.PatchInspectionHandler)
	protected.POST("/units", calibrationService.PostUnitHandler)
	protected.GET("/units", calibrationService.GetUnitsHandler)
	protected.POST("/formulations", calibrationService.PostFormulationHandler)
	protected.GET("/formulations", calibrationService.GetFormulationsHandler)
	protected.POST("/lawnservices", calibrationService.PostLawnServiceHandler)
	protected.GET("/lawnservices", calibrationService.GetLawnServicesHandler)
	protected.POST("/calibrationlogs", calibrationService.PostCalibrationLogHandler)
	protected.GET("/calibrationlogs", calibrationService.GetCalibrationLogsHandler)
	protected.GET("/calibrationlogs/:id", calibrationService.GetCalibrationLogHandler)
	protected.POST("/calibrationlogs/:id/records", calibrationService.PostCalibrationRecordHandler)
	protected.GET("/calibrationlogs/:id/records", calibrationService.GetCalibrationRecordsHandler)

	// e.POST("/upload", authService.AuthMiddleware(uploadHandler))
	// e.GET("/uploads", authService.AuthMiddleware(updloadsHandler))
	// http.Handle("/api/uploads", http.StripPrefix("/api/uploads", http.FileServer(http.Dir("uploads"))))

	fmt.Println("API server running at http://0.0.0.0:2847")
	// log.Fatal(http.ListenAndServe("localhost:2847", nil))
	e.Logger.Fatal(e.Start(":2847"))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(20 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	os.MkdirAll("uploads", 0755)
	os.MkdirAll("transcriptions", 0755)

	filePath := filepath.Join("uploads", handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	io.Copy(dst, file)

	// Background processing
	go func(filename string) {
		cmd := exec.Command("/home/caleb/dev/python/qc_ai/venv/bin/python3", "/home/caleb/dev/python/qc_ai/main.py", filename, "../transcriptions/"+filename+".txt")
		cmd.Dir = "./uploads"

		tranJob := &jobqueue.Job{
			Type:        jobqueue.Transcription,
			Description: filename,
			Run: func() error {
				if out, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf("Error processing %s: %v\nOutput:\n%s\n", filename, err, out)
				} else {
					fmt.Printf("Finished processing %s\n", filename)
				}

				reportInput := filepath.Join("transcriptions", filename+".txt")
				reportOutput := filepath.Join("uploads", filename+".txt")
				genJob := &jobqueue.Job{
					Type:        jobqueue.ReportGeneration,
					Description: filename,
					Run: func() error {
						ReportGenerator.GenerateReport(reportInput, reportOutput)
						return nil
					},
				}
				jobqueue.Enqueue(genJob)
				return err
			},
		}
		jobqueue.Enqueue(tranJob)
	}(handler.Filename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{
		Status:   "uploaded",
		Filename: handler.Filename,
	})
}

func updloadsHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir("./uploads")
	if err != nil {
		http.Error(w, "Error reading directory: "+err.Error(), http.StatusInternalServerError)
	}
	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
