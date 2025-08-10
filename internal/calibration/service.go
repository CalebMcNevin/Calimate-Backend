package calibration

import (
	"context"
	"fmt"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CalibrationService struct {
	DB *gorm.DB
}

func NewCalibrationService(db *gorm.DB) *CalibrationService {
	return &CalibrationService{DB: db}
}

// === Units ===
func (s *CalibrationService) ReadUnits() ([]Unit, error) {
	var units []Unit
	result := s.DB.Find(&units)
	return units, result.Error
}

func (s *CalibrationService) CreateUnit(unit *Unit) error {
	return s.DB.Create(unit).Error
}

// === Formulations ===
func (s *CalibrationService) ReadFormulations() ([]Formulation, error) {
	var formulations []Formulation
	result := s.DB.Find(&formulations)
	return formulations, result.Error
}

func (s *CalibrationService) CreateFormulation(formulation *Formulation) error {
	return s.DB.Create(formulation).Error
}

// === Lawn Services ===
func (s *CalibrationService) ReadLawnServices() ([]LawnService, error) {
	var services []LawnService
	result := s.DB.Preload("Formulation").Preload("TargetCalibrationUnit").Find(&services)
	return services, result.Error
}

func (s *CalibrationService) CreateLawnService(service *LawnService) error {
	Result := s.DB.Create(service)
	if Result.Error != nil {
		return Result.Error
	}
	return s.DB.Preload("Formulation").Find(service).Error
}

// === Calibration Logs ===
func (s *CalibrationService) ReadCalibrationLogs() ([]CalibrationLog, error) {
	var logs []CalibrationLog
	result := s.DB.Preload("LawnService.Formulation").Preload("Records").Find(&logs)
	return logs, result.Error
}

func evaluateCalibrationFunction(expression string, parameters map[string]interface{}) (float64, error) {
	// Create a new sandboxed evaluator
	functions := make(map[string]govaluate.ExpressionFunction)

	// Evaluate the expression with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	eval, err := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
	if err != nil {
		return 0, err
	}

	result, err := eval.Evaluate(parameters)
	if err != nil {
		return 0, err
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return 0, fmt.Errorf("calibration function evaluation timed out")
	}

	return result.(float64), nil
}

func (s *CalibrationService) ReadCalibrationLog(log_id uuid.UUID) (CalibrationLog, error) {
	var cal_log CalibrationLog
	result := s.DB.Preload("LawnService.Formulation").Preload("Records", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&cal_log, log_id)
	if result.Error != nil {
		return cal_log, result.Error
	}

	if len(cal_log.Records) > 0 {
		parameters := map[string]interface{}{
			"first_amount": cal_log.Records[0].MeasurementValue,
			"last_amount":  cal_log.Records[len(cal_log.Records)-1].MeasurementValue,
			"last_area":    float64(cal_log.Records[len(cal_log.Records)-1].MeasurementArea),
		}

		calibration, err := evaluateCalibrationFunction(cal_log.LawnService.CalibrationFunction, parameters)
		if err == nil {
			cal_log.CurrentCalibration = &calibration
		}
	}

	return cal_log, nil
}

func (s *CalibrationService) CreateCalibrationLog(log *CalibrationLog) error {
	return s.DB.Create(log).Error
}

// === Calibration Records ===
func (s *CalibrationService) ReadCalibrationRecords(filter CalibrationRecordFilter) ([]CalibrationRecord, error) {
	var records []CalibrationRecord
	query := s.DB.Model(&records)
	if filter.CalibrationLogID != nil {
		query = query.Where("calibration_log_id = ?", *filter.CalibrationLogID)
	}
	result := query.Preload("Unit").Find(&records)
	return records, result.Error
}

func (s *CalibrationService) CreateCalibrationRecord(record *CalibrationRecord) error {
	return s.DB.Create(record).Error
}
