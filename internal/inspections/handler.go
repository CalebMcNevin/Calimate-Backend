package inspections

import (
	"errors"
	"log"
	"net/http"
	"qc_api/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// PostInspectionHandler godoc
// @Summary Create a new inspection
// @Description Create a new quality control inspection
// @Tags inspections
// @Accept json
// @Produce json
// @Param inspection body InspectionDTO true "Inspection data"
// @Success 200 {object} Inspection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /inspections [post]
func (s *InspectionService) PostInspectionHandler(c echo.Context) error {
	var inspectionDTO InspectionDTO
	if err := c.Bind(&inspectionDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid JSON"})
	}
	if err := c.Validate(&inspectionDTO); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	inspection := &Inspection{
		Report:                 inspectionDTO.Report,
		UniformPPEGood:         inspectionDTO.UniformPPEGood,
		PICPresent:             inspectionDTO.PICPresent,
		MotiveLoggedIn:         inspectionDTO.MotiveLoggedIn,
		PodiumLoggedIn:         inspectionDTO.PodiumLoggedIn,
		SpillAdsorbtionPresent: inspectionDTO.SpillAdsorbtionPresent,
		Calibration:            inspectionDTO.Calibration,
		EmployeeID:             inspectionDTO.EmployeeID,
	}

	inspection, err := s.CreateInspection(inspection)
	if err != nil {
		log.Print(err.Error())
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, inspection)
}

// GetInspectionsHandler godoc
// @Summary Get all inspections
// @Description Retrieve all quality control inspections with optional filtering
// @Tags inspections
// @Accept json
// @Produce json
// @Param employee_id query string false "Filter by employee ID (UUID)"
// @Param status query string false "Filter by status"
// @Param date_from query string false "Filter by date from (YYYY-MM-DD)"
// @Param date_to query string false "Filter by date to (YYYY-MM-DD)"
// @Success 200 {array} Inspection
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /inspections [get]
func (s *InspectionService) GetInspectionsHandler(c echo.Context) error {
	var filter InspectionFilter
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	inspections, err := s.GetInspections(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to retrieve inspections"})
	}
	return c.JSON(http.StatusOK, inspections)
}

// GetInspectionHandler godoc
// @Summary Get inspection by ID
// @Description Retrieve a specific inspection by its ID
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 200 {object} Inspection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /inspections/{id} [get]
func (s *InspectionService) GetInspectionHandler(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	inspection, err := s.GetInspectionByID(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, inspection)
}

// GetByEmployeeHandler godoc
// @Summary Get inspections by employee ID
// @Description Retrieve all inspections for a specific employee
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {array} Inspection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /employees/{id}/inspections [get]
func (s *InspectionService) GetByEmployeeHandler(c echo.Context) error {
	employeeIDStr := c.Param("id")
	employee_id, err := uuid.Parse(employeeIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "Invalid employee ID"})
	}
	inspections, err := s.GetInspectionsByEmployee(employee_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, inspections)
}

// PatchInspectionHandler godoc
// @Summary Update inspection by ID
// @Description Update specific fields of an inspection by its ID
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Param inspection body InspectionPatch true "Inspection update data"
// @Success 200 {object} Inspection
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /inspections/{id} [patch]
func (s *InspectionService) PatchInspectionHandler(c echo.Context) error {
	// parse and validate inspection_id
	inspection_id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid inspection id"})
	}
	// bind json payload to InpectionPatch struct
	var patch InspectionPatch
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	inspection, err := s.UpdateInspection(inspection_id, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "update to database failed"})
	}
	return c.JSON(http.StatusOK, inspection)
}

// DeleteInspectionHandler godoc
// @Summary Delete inspection by ID
// @Description Soft-delete an inspection by its ID
// @Tags inspections
// @Accept json
// @Produce json
// @Param id path string true "Inspection ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /inspections/{id} [delete]
func (s *InspectionService) DeleteInspectionHandler(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid inspection id"})
	}

	if err := s.DeleteInspection(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{Error: "inspection not found"})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "delete failed"})
	}

	return c.NoContent(http.StatusNoContent)
}
