package employees

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, employeeService EmployeeService) {
	g.POST("/employees", employeeService.PostEmployeeHandler)
	g.GET("/employees", employeeService.GetEmployeesHandler)
	g.GET("/employees/:id", employeeService.GetEmployeeByIDHandler)
	g.PATCH("/employees/:id", employeeService.PatchEmployeeHandler)
}
