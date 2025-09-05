package calibration

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, calibrationService *CalibrationService) {
	g.POST("/formulations", calibrationService.PostFormulationHandler)
	g.GET("/formulations", calibrationService.GetFormulationsHandler)
	g.PATCH("/formulations/:id", calibrationService.PatchFormulationHandler)
	g.POST("/lawnservices", calibrationService.PostLawnServiceHandler)
	g.GET("/lawnservices", calibrationService.GetLawnServicesHandler)
	g.PATCH("/lawnservices/:id", calibrationService.PatchLawnServiceHandler)
	g.POST("/calibrationlogs", calibrationService.PostCalibrationLogHandler)
	g.GET("/calibrationlogs", calibrationService.GetCalibrationLogsHandler)
	g.GET("/calibrationlogs/:id", calibrationService.GetCalibrationLogHandler)
	g.DELETE("/calibrationlogs/:id", calibrationService.DeleteCalibrationLogHandler)
	g.PATCH("/calibrationlogs/:id", calibrationService.PatchCalibrationLogHandler)
	g.POST("/calibrationlogs/:id/records", calibrationService.PostCalibrationRecordHandler)
	g.GET("/calibrationlogs/:id/records", calibrationService.GetCalibrationRecordsHandler)
	g.PATCH("/calibrationlogs/:logId/records/:id", calibrationService.PatchCalibrationRecordHandler)
	g.DELETE("/calibrationlogs/:logId/records/:id", calibrationService.DeleteCalibrationRecordHandler)
}
