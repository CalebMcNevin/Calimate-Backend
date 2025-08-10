package inspections

import (
	"log"
	"net/http"
	"qc_api/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *InspectionService) PostInspectionHandler(c echo.Context) error {
	var NewInspection Inspection
	if err := c.Bind(&NewInspection); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid JSON"})
	}
	inspection, err := s.CreateInspection(&NewInspection)
	if err != nil {
		log.Print(err.Error())
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, inspection)
}

func (s *InspectionService) GetInspectionsHandler(c echo.Context) error {
	inspections, err := s.GetInspections()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, inspections)
}

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

func (s *InspectionService) GetByEmployeeHandler(c echo.Context) error {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "user_id not found in context"})
	}
	employee_id, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	inspections, err := s.GetInspectionsByEmployee(employee_id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, inspections)
}

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
