package calibration

import (
	"errors"
	"net/http"
	"qc_api/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// === Formulations ===

// GetFormulationsHandler godoc
// @Summary Get all formulations
// @Description Retrieve all formulation types
// @Tags calibration
// @Accept json
// @Produce json
// @Success 200 {array} Formulation
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /formulations [get]
func (s *CalibrationService) GetFormulationsHandler(c echo.Context) error {
	formulations, err := s.ReadFormulations()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, formulations)
}

// PostFormulationHandler godoc
// @Summary Create a new formulation
// @Description Create a new formulation type
// @Tags calibration
// @Accept json
// @Produce json
// @Param formulation body FormulationDTO true "Formulation data"
// @Success 201 {object} Formulation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /formulations [post]
func (s *CalibrationService) PostFormulationHandler(c echo.Context) error {
	var formulationDTO FormulationDTO
	if err := c.Bind(&formulationDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&formulationDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	formulation := &Formulation{
		Name: formulationDTO.Name,
	}

	if err := s.CreateFormulation(formulation); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, formulation)
}

// === Lawn Services ===

// GetLawnServicesHandler godoc
// @Summary Get all lawn services
// @Description Retrieve all lawn service configurations
// @Tags calibration
// @Accept json
// @Produce json
// @Success 200 {array} LawnService
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /lawnservices [get]
func (s *CalibrationService) GetLawnServicesHandler(c echo.Context) error {
	services, err := s.ReadLawnServices()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, services)
}

// PostLawnServiceHandler godoc
// @Summary Create a new lawn service
// @Description Create a new lawn service configuration
// @Tags calibration
// @Accept json
// @Produce json
// @Param lawnservice body LawnServiceDTO true "Lawn service data"
// @Success 201 {object} LawnService
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /lawnservices [post]
func (s *CalibrationService) PostLawnServiceHandler(c echo.Context) error {
	var lawnServiceDTO LawnServiceDTO
	if err := c.Bind(&lawnServiceDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&lawnServiceDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	lawnService := &LawnService{
		Code:                            lawnServiceDTO.Code,
		Description:                     lawnServiceDTO.Description,
		FormulationID:                   lawnServiceDTO.FormulationID,
		TargetCalibrationValue:          lawnServiceDTO.TargetCalibrationValue,
		TargetCalibrationUnit:           lawnServiceDTO.TargetCalibrationUnit,
		MeasurementUnit:                 lawnServiceDTO.MeasurementUnit,
		CalibrationFunction:             lawnServiceDTO.CalibrationFunction,
		DifferentialCalibrationFunction: lawnServiceDTO.DifferentialCalibrationFunction,
	}

	if err := s.CreateLawnService(lawnService); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, lawnService)
}

// === Calibration Logs ===

// GetCalibrationLogsHandler godoc
// @Summary Get all calibration logs
// @Description Retrieve all calibration logs with optional filtering
// @Tags calibration
// @Accept json
// @Produce json
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param lawn_service_id query string false "Filter by lawn service ID (UUID)"
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Success 200 {array} CalibrationLog
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs [get]
func (s *CalibrationService) GetCalibrationLogsHandler(c echo.Context) error {
	var filter CalibrationLogFilter
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	logs, err := s.ReadCalibrationLogs(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, logs)
}

// GetCalibrationLogHandler godoc
// @Summary Get calibration log by ID
// @Description Retrieve a specific calibration log by its ID
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Calibration Log ID"
// @Success 200 {object} CalibrationLog
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs/{id} [get]
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

// PostCalibrationLogHandler godoc
// @Summary Create a new calibration log
// @Description Create a new calibration log entry
// @Tags calibration
// @Accept json
// @Produce json
// @Param calibrationlog body CalibrationLogDTO true "Calibration log data"
// @Success 201 {object} CalibrationLog
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs [post]
func (s *CalibrationService) PostCalibrationLogHandler(c echo.Context) error {
	var logDTO CalibrationLogDTO
	if err := c.Bind(&logDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid request body"})
	}
	if err := c.Validate(&logDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	log, err := s.CreateCalibrationLog(&logDTO, c.Get("user_id").(uuid.UUID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	if log.Records == nil {
		log.Records = []CalibrationRecord{}
	}
	return c.JSON(http.StatusCreated, log)
}

// === Calibration Records ===

// GetCalibrationRecordsHandler godoc
// @Summary Get calibration records for a log
// @Description Retrieve all calibration records for a specific calibration log
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Calibration Log ID"
// @Success 200 {array} CalibrationRecord
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs/{id}/records [get]
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

// PostCalibrationRecordHandler godoc
// @Summary Create a new calibration record
// @Description Create a new calibration record for a specific calibration log
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Calibration Log ID"
// @Param record body CalibrationRecordDTO true "Calibration record data"
// @Success 201 {object} CalibrationRecord
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs/{id}/records [post]
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
		MeasurementValue: float64(*recordDTO.Value),
		MeasurementArea:  *recordDTO.Area,
		MeasurementUnit:  recordDTO.Units,
	}

	if err := s.CreateCalibrationRecord(record); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusCreated, record)
}

// === PATCH Handlers ===

// PatchFormulationHandler godoc
// @Summary Update formulation by ID
// @Description Update specific fields of a formulation by its ID
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Formulation ID"
// @Param formulation body FormulationPatch true "Formulation update data"
// @Success 200 {object} Formulation
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /formulations/{id} [patch]
func (s *CalibrationService) PatchFormulationHandler(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid formulation id"})
	}
	var patch FormulationPatch
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	formulation, err := s.UpdateFormulation(id, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "update to database failed"})
	}
	return c.JSON(http.StatusOK, formulation)
}

// PatchLawnServiceHandler godoc
// @Summary Update lawn service by ID
// @Description Update specific fields of a lawn service by its ID
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Lawn Service ID"
// @Param lawnservice body LawnServicePatch true "Lawn service update data"
// @Success 200 {object} LawnService
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /lawnservices/{id} [patch]
func (s *CalibrationService) PatchLawnServiceHandler(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid lawn service id"})
	}
	var patch LawnServicePatch
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	service, err := s.UpdateLawnService(id, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "update to database failed"})
	}
	return c.JSON(http.StatusOK, service)
}

// PatchCalibrationLogHandler godoc
// @Summary Update calibration log by ID
// @Description Update specific fields of a calibration log by its ID
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Calibration Log ID"
// @Param calibrationlog body CalibrationLogPatch true "Calibration log update data"
// @Success 200 {object} CalibrationLog
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs/{id} [patch]
func (s *CalibrationService) PatchCalibrationLogHandler(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid calibration log id"})
	}
	var patch CalibrationLogPatch
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	log, err := s.UpdateCalibrationLog(id, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "update to database failed"})
	}
	return c.JSON(http.StatusOK, log)
}

// PatchCalibrationRecordHandler godoc
// @Summary Update calibration record by ID
// @Description Update specific fields of a calibration record by its ID
// @Tags calibration
// @Accept json
// @Produce json
// @Param logId path string true "Calibration Log ID"
// @Param id path string true "Calibration Record ID"
// @Param record body CalibrationRecordPatch true "Calibration record update data"
// @Success 200 {object} CalibrationRecord
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs/{logId}/records/{id} [patch]
func (s *CalibrationService) PatchCalibrationRecordHandler(c echo.Context) error {
	recordId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid calibration record id"})
	}
	var patch CalibrationRecordPatch
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	record, err := s.UpdateCalibrationRecord(recordId, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "update to database failed"})
	}
	return c.JSON(http.StatusOK, record)
}

// DeleteCalibrationLogHandler godoc
// @Summary Delete calibration log by ID
// @Description Soft-delete a calibration log by its ID
// @Tags calibration
// @Accept json
// @Produce json
// @Param id path string true "Calibration Log ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /calibrationlogs/{id} [delete]
func (s *CalibrationService) DeleteCalibrationLogHandler(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid calibration log id"})
	}

	if err := s.DeleteCalibrationLog(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "calibration log not found"})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "delete failed"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *CalibrationService) DeleteCalibrationRecordHandler(c echo.Context) error {
	recordId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid calibration record id"})
	}

	if err := s.DeleteCalibrationRecord(recordId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "calibration record not found"})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "delete failed"})
	}

	return c.NoContent(http.StatusNoContent)
}
