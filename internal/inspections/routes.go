package inspections

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, inspectionService *InspectionService) {
	g.GET("/employees/:id/inspections", inspectionService.GetByEmployeeHandler)
	g.POST("/inspections", inspectionService.PostInspectionHandler)
	g.GET("/inspections", inspectionService.GetInspectionsHandler)
	g.GET("/inspections/:id", inspectionService.GetInspectionHandler)
	g.PATCH("/inspections/:id", inspectionService.PatchInspectionHandler)
	g.DELETE("/inspections/:id", inspectionService.DeleteInspectionHandler)
}
