package calibration

import (
	"fmt"
	"qc_api/internal/utils"

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
	result := s.DB.Preload("Formulation").Find(&services)
	return services, result.Error
}

func (s *CalibrationService) CreateLawnService(service *LawnService) error {
	result := s.DB.Create(service)
	if result.Error != nil {
		return result.Error
	}
	return s.DB.Preload("Formulation").Find(service).Error
}

func (s *CalibrationService) UpdateLawnService(id uuid.UUID, patch LawnServicePatch) (*LawnService, error) {
	result := s.DB.Model(&LawnService{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var service LawnService
	result = s.DB.Preload("Formulation").Where("id = ?", id).First(&service)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// === Calibration Logs ===
func (s *CalibrationService) ReadCalibrationLogs(filter CalibrationLogFilter) ([]CalibrationLog, error) {
	var logs []CalibrationLog
	query := utils.ApplyFilter(s.DB.Model(&CalibrationLog{}), filter)
	result := query.Preload("LawnService.Formulation").Preload("Records", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Find(&logs)
	if result.Error != nil {
		return logs, result.Error
	}

	// Calculate calibration for each log
	for i := range logs {
		s.calculateCalibrationForLog(&logs[i])
		// Calculate calibrations for individual records
		s.calculateCalibrationForRecords(&logs[i])
	}

	return logs, nil
}

func evaluateCalibrationFunction(expression string, parameters map[string]any) (float64, error) {
	// Create functions map for future expansion
	functions := make(map[string]govaluate.ExpressionFunction)

	calFunc, err := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
	if err != nil {
		return 0, err
	}

	result, err := calFunc.Evaluate(parameters)
	if err != nil {
		return 0, err
	}

	// Safe type conversion
	switch v := result.(type) {
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("expression returned non-numeric type: %T", result)
	}
}

func (s *CalibrationService) calculateCalibrationForLog(log *CalibrationLog) {
	if len(log.Records) == 0 || log.LawnService.CalibrationFunction == "" {
		return
	}

	firstRecord := log.Records[0]
	lastRecord := log.Records[len(log.Records)-1]

	parameters := map[string]any{
		"first_amount":   firstRecord.MeasurementValue,
		"current_amount": lastRecord.MeasurementValue,
		"first_area":     float64(firstRecord.MeasurementArea),
		"current_area":   float64(lastRecord.MeasurementArea),
	}

	calibration, err := evaluateCalibrationFunction(log.LawnService.CalibrationFunction, parameters)
	if err == nil {
		log.CurrentCalibration = &calibration
	}
}

func (s *CalibrationService) calculateCalibrationForRecords(log *CalibrationLog) {
	if len(log.Records) == 0 || log.LawnService.DifferentialCalibrationFunction == "" {
		return
	}

	// Records are already sorted by created_at due to the Preload order
	for i := 0; i < len(log.Records); i++ {
		currentRecord := log.Records[i]

		// For the first record, use 0 for previous values
		var previousAmount float64 = 0
		var previousArea float64 = 0

		// For subsequent records, use actual previous record values
		if i > 0 {
			previousRecord := log.Records[i-1]
			previousAmount = previousRecord.MeasurementValue
			previousArea = float64(previousRecord.MeasurementArea)
		}

		parameters := map[string]any{
			"current_amount":  currentRecord.MeasurementValue,
			"current_area":    float64(currentRecord.MeasurementArea),
			"previous_amount": previousAmount,
			"previous_area":   previousArea,
		}

		calibration, err := evaluateCalibrationFunction(log.LawnService.DifferentialCalibrationFunction, parameters)
		if err == nil {
			log.Records[i].Calibration = calibration
		}
	}
}

func (s *CalibrationService) ReadCalibrationLog(log_id uuid.UUID) (CalibrationLog, error) {
	var cal_log CalibrationLog
	result := s.DB.Preload("LawnService.Formulation").Preload("Records", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&cal_log, log_id)
	if result.Error != nil {
		return cal_log, result.Error
	}

	// Calculate calibration using the helper function
	s.calculateCalibrationForLog(&cal_log)

	// Calculate calibrations for individual records
	s.calculateCalibrationForRecords(&cal_log)

	return cal_log, nil
}

func (s *CalibrationService) CreateCalibrationLog(log *CalibrationLogDTO, userID uuid.UUID) (*CalibrationLog, error) {
	calibrationLog := &CalibrationLog{
		LawnServiceID: log.LawnServiceID,
		UserID:        userID,
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

func (s *CalibrationService) DeleteCalibrationLog(id uuid.UUID) error {
	result := s.DB.Delete(&CalibrationLog{}, id)
	return result.Error
}

// === Calibration Records ===
func (s *CalibrationService) ReadCalibrationRecords(filter CalibrationRecordFilter) ([]CalibrationRecord, error) {
	var records []CalibrationRecord
	query := utils.ApplyFilter(s.DB.Model(&CalibrationRecord{}), filter)
	result := query.Find(&records)
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
	result = s.DB.Where("id = ?", id).First(&record)
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
