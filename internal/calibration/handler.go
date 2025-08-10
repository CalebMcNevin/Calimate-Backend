package calibration

import (
	"errors"
	"net/http"
	"qc_api/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// === Units ===
func (s *CalibrationService) GetUnitsHandler(c echo.Context) error {
	units, err := s.ReadUnits()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, units)
}

func (s *CalibrationService) PostUnitHandler(c echo.Context) error {
	var unit Unit
	if err := c.Bind(&unit); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&unit); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	if err := s.CreateUnit(&unit); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, unit)
}

// === Formulations ===
func (s *CalibrationService) GetFormulationsHandler(c echo.Context) error {
	formulations, err := s.ReadFormulations()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, formulations)
}

func (s *CalibrationService) PostFormulationHandler(c echo.Context) error {
	var formulation Formulation
	if err := c.Bind(&formulation); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&formulation); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	if err := s.CreateFormulation(&formulation); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, formulation)
}

// === Lawn Services ===
func (s *CalibrationService) GetLawnServicesHandler(c echo.Context) error {
	services, err := s.ReadLawnServices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, services)
}

func (s *CalibrationService) PostLawnServiceHandler(c echo.Context) error {
	var lawnServiceDTO LawnServiceDTO
	if err := c.Bind(&lawnServiceDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&lawnServiceDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	lawnService := &LawnService{
		Code:                      lawnServiceDTO.Code,
		Description:               lawnServiceDTO.Description,
		FormulationID:             lawnServiceDTO.FormulationID,
		TargetCalibrationValue:    lawnServiceDTO.TargetCalibrationValue,
		TargetCalibrationUnitCode: lawnServiceDTO.TargetCalibrationUnitCode,
		CalibrationFunction:       lawnServiceDTO.CalibrationFunction,
	}

	if err := s.CreateLawnService(lawnService); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, lawnService)
}

// === Calibration Logs ===
func (s *CalibrationService) GetCalibrationLogsHandler(c echo.Context) error {
	logs, err := s.ReadCalibrationLogs()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, logs)
}

func (s *CalibrationService) GetCalibrationLogHandler(c echo.Context) error {
	log_id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid calibration_log id"})
	}
	log, err := s.ReadCalibrationLog(log_id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "Calibration log not found"})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, log)
}

func (s *CalibrationService) PostCalibrationLogHandler(c echo.Context) error {
	var log CalibrationLog
	if err := c.Bind(&log); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := s.CreateCalibrationLog(&log); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	if log.Records == nil {
		log.Records = []CalibrationRecord{}
	}
	return c.JSON(http.StatusCreated, log)
}

// === Calibration Records ===
func (s *CalibrationService) GetCalibrationRecordsHandler(c echo.Context) error {
	LogIDStr := c.Param("id")
	LogID, err := uuid.Parse(LogIDStr)
	if err != nil {
		return err
	}
	filter := CalibrationRecordFilter{
		CalibrationLogID: &LogID,
	}
	records, err := s.ReadCalibrationRecords(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, records)
}

func (s *CalibrationService) PostCalibrationRecordHandler(c echo.Context) error {
	// will post at /calibrationlogs/:id/records, so we have the log id
	log_id_str := c.Param("id")
	log_id, err := uuid.Parse(log_id_str)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid Calibration Log ID"})
	}
	var recordDTO CalibrationRecordDTO
	if err := c.Bind(&recordDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&recordDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	record := &CalibrationRecord{
		CalibrationLogID: log_id,
		MeasurementValue: float64(recordDTO.Value),
		MeasurementUnit:  recordDTO.Units,
	}

	if err := s.CreateCalibrationRecord(record); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, record)
}
