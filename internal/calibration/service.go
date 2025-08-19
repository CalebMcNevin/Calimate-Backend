package calibration

import (
	"context"
	"fmt"
	"qc_api/internal/utils"
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

func (s *CalibrationService) UpdateUnit(symbol string, patch UnitPatch) (*Unit, error) {
	result := s.DB.Model(&Unit{}).Where("symbol = ?", symbol).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var unit Unit
	result = s.DB.Where("symbol = ?", symbol).First(&unit)
	if result.Error != nil {
		return nil, result.Error
	}
	return &unit, nil
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

func (s *CalibrationService) UpdateFormulation(id uuid.UUID, patch FormulationPatch) (*Formulation, error) {
	result := s.DB.Model(&Formulation{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var formulation Formulation
	result = s.DB.Where("id = ?", id).First(&formulation)
	if result.Error != nil {
		return nil, result.Error
	}
	return &formulation, nil
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

func (s *CalibrationService) UpdateLawnService(id uuid.UUID, patch LawnServicePatch) (*LawnService, error) {
	result := s.DB.Model(&LawnService{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var service LawnService
	result = s.DB.Preload("Formulation").Preload("TargetCalibrationUnit").Where("id = ?", id).First(&service)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// === Calibration Logs ===
func (s *CalibrationService) ReadCalibrationLogs(filter CalibrationLogFilter) ([]CalibrationLog, error) {
	var logs []CalibrationLog
	query := utils.ApplyFilter(s.DB.Model(&CalibrationLog{}), filter)
	result := query.Preload("LawnService.Formulation").Preload("Records").Find(&logs)
	return logs, result.Error
}

func evaluateCalibrationFunction(expression string, parameters map[string]any) (float64, error) {
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
		parameters := map[string]any{
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

func (s *CalibrationService) CreateCalibrationLog(log *CalibrationLogDTO) (*CalibrationLog, error) {
	calibrationLog := &CalibrationLog{
		LawnServiceID: log.LawnServiceID,
	}
	result := s.DB.Create(calibrationLog)
	if result.Error != nil {
		return nil, result.Error
	}
	return calibrationLog, nil
}

func (s *CalibrationService) UpdateCalibrationLog(id uuid.UUID, patch CalibrationLogPatch) (*CalibrationLog, error) {
	result := s.DB.Model(&CalibrationLog{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var log CalibrationLog
	result = s.DB.Preload("LawnService.Formulation").Preload("Records").Where("id = ?", id).First(&log)
	if result.Error != nil {
		return nil, result.Error
	}
	return &log, nil
}

// === Calibration Records ===
func (s *CalibrationService) ReadCalibrationRecords(filter CalibrationRecordFilter) ([]CalibrationRecord, error) {
	var records []CalibrationRecord
	query := utils.ApplyFilter(s.DB.Model(&CalibrationRecord{}), filter)
	result := query.Preload("Unit").Find(&records)
	return records, result.Error
}

func (s *CalibrationService) CreateCalibrationRecord(record *CalibrationRecord) error {
	return s.DB.Create(record).Error
}

func (s *CalibrationService) UpdateCalibrationRecord(id uuid.UUID, patch CalibrationRecordPatch) (*CalibrationRecord, error) {
	result := s.DB.Model(&CalibrationRecord{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var record CalibrationRecord
	result = s.DB.Preload("Unit").Where("id = ?", id).First(&record)
	if result.Error != nil {
		return nil, result.Error
	}
	return &record, nil
}

func (s *CalibrationService) DeleteCalibrationRecord(id uuid.UUID) error {
	result := s.DB.Delete(&CalibrationRecord{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
