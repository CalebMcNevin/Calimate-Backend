package calibration_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qc_api/internal/calibration"
	"qc_api/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(calibration.Models()...)
	return db
}

func TestPostUnit(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Request
	reqBody := `{"symbol": "kg", "description": "Kilogram"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostUnitHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPostUnitMissingSymbol(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Request
	reqBody := `{"description": "Kilogram"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostUnitHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostUnitMissingDescription(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Request
	reqBody := `{"symbol": "kg"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostUnitHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostUnitExistingSymbol(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a unit
	unit := &calibration.Unit{Symbol: "kg", Description: "Kilogram"}
	db.Create(unit)

	// Request
	reqBody := `{"symbol": "kg", "description": "Kilogram"}`
	req := httptest.NewRequest(http.MethodPost, "/units", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostUnitHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestPostFormulation(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Request
	reqBody := `{"name": "GRANULE"}`
	req := httptest.NewRequest(http.MethodPost, "/formulations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostFormulationHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPostFormulationMissingName(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Request
	reqBody := `{}`
	req := httptest.NewRequest(http.MethodPost, "/formulations", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostFormulationHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostLawnService(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a formulation and unit
	formulation := &calibration.Formulation{Name: "GRANULE"}
	db.Create(formulation)
	unit := &calibration.Unit{Symbol: "kg", Description: "Kilogram"}
	db.Create(unit)

	// Request
	reqBody := `{"code": "LS01", "description": "Spring Fertilizer", "formulation_id": "` + formulation.ID.String() + `", "target_calibration_value": 1.5, "target_calibration_unit_code": "kg", "calibration_function": "last_amount / last_area"}`
	req := httptest.NewRequest(http.MethodPost, "/lawnservices", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostLawnServiceHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPostLawnServiceMissingCode(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a formulation and unit
	formulation := &calibration.Formulation{Name: "GRANULE"}
	db.Create(formulation)
	unit := &calibration.Unit{Symbol: "kg", Description: "Kilogram"}
	db.Create(unit)

	// Request
	reqBody := `{"description": "Spring Fertilizer", "formulation_id": "` + formulation.ID.String() + `", "target_calibration_value": 1.5, "target_calibration_unit_code": "kg", "calibration_function": "last_amount / last_area"}`
	req := httptest.NewRequest(http.MethodPost, "/lawnservices", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostLawnServiceHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostCalibrationLog(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a lawn service
	formulation := &calibration.Formulation{Name: "GRANULE"}
	db.Create(formulation)
	lawnService := &calibration.LawnService{
		Code:                      "LS01",
		Description:               "Spring Fertilizer",
		FormulationID:             formulation.ID,
		TargetCalibrationValue:    1.5,
		TargetCalibrationUnitCode: "kg",
		CalibrationFunction:       "last_amount / last_area",
	}
	db.Create(lawnService)

	// Request
	reqBody := `{"lawn_service_id": "` + lawnService.ID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/calibrationlogs", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostCalibrationLogHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPostCalibrationLogEmptyRecords(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a lawn service
	formulation := &calibration.Formulation{Name: "GRANULE"}
	db.Create(formulation)
	lawnService := &calibration.LawnService{
		Code:                      "LS01",
		Description:               "Spring Fertilizer",
		FormulationID:             formulation.ID,
		TargetCalibrationValue:    1.5,
		TargetCalibrationUnitCode: "kg",
		CalibrationFunction:       "last_amount / last_area",
	}
	db.Create(lawnService)

	// Request
	reqBody := `{"lawn_service_id": "` + lawnService.ID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/calibrationlogs", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Handler
	err := calibService.PostCalibrationLogHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), `"records":[]`)
}

func TestPostCalibrationRecord(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a calibration log
	calibLog := &calibration.CalibrationLog{UserID: uuid.New(), LawnServiceID: uuid.New()}
	db.Create(calibLog)

	// Request
	reqBody := `{"measurement_value": 1.5, "units": "kg"}`
	req := httptest.NewRequest(http.MethodPost, "/calibrationlogs/"+calibLog.ID.String()+"/records", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(calibLog.ID.String())

	// Handler
	err := calibService.PostCalibrationRecordHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestPostCalibrationRecordInvalidBody(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a calibration log
	calibLog := &calibration.CalibrationLog{UserID: uuid.New(), LawnServiceID: uuid.New()}
	db.Create(calibLog)

	// Request
	reqBody := `{"measurement_value": "invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/calibrationlogs/"+calibLog.ID.String()+"/records", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(calibLog.ID.String())

	// Handler
	err := calibService.PostCalibrationRecordHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCalibrationLogNoRecords(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a calibration log
	calibLog := &calibration.CalibrationLog{UserID: uuid.New(), LawnServiceID: uuid.New()}
	db.Create(calibLog)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/calibrationlogs/"+calibLog.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(calibLog.ID.String())

	// Handler
	err := calibService.GetCalibrationLogHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), `"current_calibration"`)
}

func TestGetCalibrationLogNotFound(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Request
	req := httptest.NewRequest(http.MethodGet, "/calibrationlogs/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())

	// Handler
	err := calibService.GetCalibrationLogHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetCalibrationLogInvalidFunction(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a formulation and lawn service with an invalid calibration function
	formulation := &calibration.Formulation{Name: "GRANULE"}
	db.Create(formulation)
	lawnService := &calibration.LawnService{
		Code:                      "LS01",
		Description:               "Spring Fertilizer",
		FormulationID:             formulation.ID,
		TargetCalibrationValue:    1.5,
		TargetCalibrationUnitCode: "kg",
		CalibrationFunction:       "invalid_function +",
	}
	db.Create(lawnService)

	// Create a calibration log
	calibLog := &calibration.CalibrationLog{UserID: uuid.New(), LawnServiceID: lawnService.ID}
	db.Create(calibLog)

	// Create a calibration record
	record := &calibration.CalibrationRecord{
		CalibrationLogID: calibLog.ID,
		MeasurementValue: 1.5,
		MeasurementUnit:  "kg",
		MeasurementArea:  100,
	}
	db.Create(record)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/calibrationlogs/"+calibLog.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(calibLog.ID.String())

	// Handler
	err := calibService.GetCalibrationLogHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// Should not contain current_calibration due to invalid function
	assert.NotContains(t, rec.Body.String(), `"current_calibration"`)
}

func TestGetCalibrationLogWithRecords(t *testing.T) {
	// Setup
	db := setupTestDB()
	calibService := calibration.NewCalibrationService(db)
	e := echo.New()
	e.Validator = utils.NewValidator()

	// Create a formulation and lawn service
	formulation := &calibration.Formulation{Name: "GRANULE"}
	db.Create(formulation)
	lawnService := &calibration.LawnService{
		Code:                      "LS01",
		Description:               "Spring Fertilizer",
		FormulationID:             formulation.ID,
		TargetCalibrationValue:    1.5,
		TargetCalibrationUnitCode: "kg",
		CalibrationFunction:       "last_amount / last_area",
	}
	db.Create(lawnService)

	// Create a calibration log
	calibLog := &calibration.CalibrationLog{UserID: uuid.New(), LawnServiceID: lawnService.ID}
	db.Create(calibLog)

	// Create a calibration record
	record := &calibration.CalibrationRecord{
		CalibrationLogID: calibLog.ID,
		MeasurementValue: 1.5,
		MeasurementUnit:  "kg",
		MeasurementArea:  100,
	}
	db.Create(record)

	// Request
	req := httptest.NewRequest(http.MethodGet, "/calibrationlogs/"+calibLog.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(calibLog.ID.String())

	// Handler
	err := calibService.GetCalibrationLogHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"current_calibration":0.015`)
}
