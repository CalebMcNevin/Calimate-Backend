package employees

import (
	"log"
	"net/http"
	"qc_api/internal/utils"

	"github.com/labstack/echo/v4"
)

// PostEmployeeHandler godoc
// @Summary Create a new employee
// @Description Create a new employee record
// @Tags employees
// @Accept json
// @Produce json
// @Param employee body EmployeeDTO true "Employee data"
// @Success 201 {object} Employee
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /employees [post]
func (s *EmployeeService) PostEmployeeHandler(c echo.Context) error {
	var empReq EmployeeDTO
	if err := c.Bind(&empReq); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid JSON"})
	}
	if err := c.Validate(&empReq); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	employee, err := NewEmployee(empReq.CommonName, empReq.FirstName, empReq.LastName, empReq.EmployeeNumber)
	if err != nil {
		log.Printf("Error constructing Employee struct: %v", err)
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	err = s.CreateEmployee(employee)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create employee",
		})
	}
	return c.JSON(http.StatusCreated, employee)
}

// GetEmployeesHandler godoc
// @Summary Get all employees
// @Description Retrieve all employee records with optional filtering
// @Tags employees
// @Accept json
// @Produce json
// @Param active query bool false "Filter by active status"
// @Param employee_number query string false "Filter by employee number"
// @Success 200 {array} Employee
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /employees [get]
func (s *EmployeeService) GetEmployeesHandler(c echo.Context) error {
	var filter EmployeeFilter
	if err := c.Bind(&filter); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}

	employees, err := s.GetEmployees(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to retrieve employees"})
	}
	return c.JSON(http.StatusOK, employees)
}

// GetEmployeeByIDHandler godoc
// @Summary Get employee by ID
// @Description Retrieve a specific employee by their ID
// @Tags employees
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} Employee
// @Failure 404 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /employees/{id} [get]
func (s *EmployeeService) GetEmployeeByIDHandler(c echo.Context) error {
	employee, err := s.GetEmployeeByID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, employee)
}

// PatchEmployeeHandler godoc
// @Summary Update employee by ID
// @Description Update specific fields of an employee by their ID
// @Tags employees
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param employee body EmployeePatch true "Employee update data"
// @Success 200 {object} Employee
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Security BearerAuth
// @Router /employees/{id} [patch]
func (s *EmployeeService) PatchEmployeeHandler(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid employee ID"})
	}
	var patch EmployeePatch
	if err := c.Bind(&patch); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
	}
	employee, err := s.UpdateEmployee(id, patch)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "update to database failed"})
	}
	return c.JSON(http.StatusOK, employee)
}
