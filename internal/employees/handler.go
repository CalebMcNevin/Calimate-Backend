package employees

import (
	"log"
	"net/http"
	"qc_api/internal/utils"

	"github.com/labstack/echo/v4"
)

func (s *EmployeeService) CreateEmployeeHandler(c echo.Context) error {
	var empReq Employee
	if err := c.Bind(&empReq); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
	}
	employee, err := NewEmployee(empReq.CommonName, empReq.FirstName, empReq.LastName, empReq.EmployeeNumber)
	if err != nil {
		log.Printf("Error constructing Employee struct: %v", err)
		return err
	}
	err = s.CreateEmployee(employee)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create employee",
		})
	}
	return nil
}

func (s *EmployeeService) GetEmployeesHandler(c echo.Context) error {
	employees, err := s.GetEmployees()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: "failed to retrieve employees"})
	}
	return c.JSON(http.StatusOK, employees)
}

func (s *EmployeeService) GetEmployeeByIDHandler(c echo.Context) error {
	employee, err := s.GetEmployeeByID(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, employee)
}
