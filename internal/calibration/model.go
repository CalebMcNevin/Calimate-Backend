package calibration

import (
	"qc_api/internal/db"
	"qc_api/internal/utils"

	"github.com/google/uuid"
)

func Models() []any {
	return []any{
		&Formulation{},
		&LawnService{},
		&CalibrationLog{},
		&CalibrationRecord{},
	}
}

type Formulation struct {
	db.BaseModel
	Name string `gorm:"unique;not null" json:"name" validate:"required"` //e.g. "GRANULE"
}

type FormulationDTO struct {
	Name string `json:"name" validate:"required"`
}

type FormulationPatch struct {
	Name *string `json:"name,omitempty"`
}

type LawnService struct {
	db.BaseModel
	Code                            string      `gorm:"unique;not null" json:"code" validate:"required"` //e.g. "LS01"
	Description                     string      `json:"description" validate:"required"`                 //e.g. "Spring Fertilizer"
	FormulationID                   uuid.UUID   `json:"formulation_id" validate:"required"`
	Formulation                     Formulation `gorm:"foreignKey:FormulationID" json:"formulation"`
	TargetCalibrationValue          float32     `json:"target_calibration_value" validate:"required"`
	TargetCalibrationUnit           string      `gorm:"not null" json:"target_calibration_unit" validate:"required"`
	MeasurementUnit                 string      `json:"measurement_unit" validate:"required"`
	CalibrationFunction             string      `json:"calibration_function" validate:"required"`
	DifferentialCalibrationFunction string      `json:"differential_calibration_function"`
}

type LawnServiceDTO struct {
	Code                            string    `json:"code" validate:"required"`
	Description                     string    `json:"description" validate:"required"`
	FormulationID                   uuid.UUID `json:"formulation_id" validate:"required"`
	TargetCalibrationValue          float32   `json:"target_calibration_value" validate:"required"`
	TargetCalibrationUnit           string    `json:"target_calibration_unit" validate:"required"`
	MeasurementUnit                 string    `json:"measurement_unit" validate:"required"`
	CalibrationFunction             string    `json:"calibration_function" validate:"required"`
	DifferentialCalibrationFunction string    `json:"differential_calibration_function"`
}

type LawnServicePatch struct {
	Code                            *string    `json:"code,omitempty"`
	Description                     *string    `json:"description,omitempty"`
	FormulationID                   *uuid.UUID `json:"formulation_id,omitempty"`
	TargetCalibrationValue          *float32   `json:"target_calibration_value,omitempty"`
	TargetCalibrationUnit           *string    `json:"target_calibration_unit,omitempty"`
	MeasurementUnit                 *string    `json:"measurement_unit,omitempty"`
	CalibrationFunction             *string    `json:"calibration_function,omitempty"`
	DifferentialCalibrationFunction *string    `json:"differential_calibration_function,omitempty"`
}

type CalibrationLog struct {
	db.BaseModel
	UserID             uuid.UUID           `json:"user_id"`
	LawnServiceID      uuid.UUID           `json:"lawn_service_id"`
	LawnService        LawnService         `gorm:"foreignKey:LawnServiceID" json:"-"`
	CurrentCalibration *float64            `gorm:"-" json:"current_calibration,omitempty"`
	Records            []CalibrationRecord `json:"records"`
}

type CalibrationLogDTO struct {
	LawnServiceID uuid.UUID `json:"lawn_service_id"`
}

type CalibrationLogPatch struct {
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	LawnServiceID *uuid.UUID `json:"lawn_service_id,omitempty"`
}

type CalibrationLogFilter struct {
	UserID        *uuid.UUID        `json:"user_id,omitempty" query:"user_id"`
	LawnServiceID *uuid.UUID        `json:"lawn_service_id,omitempty" query:"lawn_service_id"`
	DateFrom      *utils.SimpleDate `json:"date_from,omitempty" query:"date_from" filter:"created_at"`
	DateTo        *utils.SimpleDate `json:"date_to,omitempty" query:"date_to" filter:"created_at"`
}

type CalibrationRecord struct {
	db.BaseModel
	CalibrationLogID uuid.UUID      `json:"calibration_log_id"`
	CalibrationLog   CalibrationLog `gorm:"foreignKey:CalibrationLogID" json:"-"`
	MeasurementValue float64        `json:"measurement_value"`
	MeasurementUnit  string         `json:"measurement_unit"`
	MeasurementArea  uint           `json:"measurement_area"`
	Calibration      float64        `gorm:"-" json:"calibration"`
}

type CalibrationRecordDTO struct {
	Value *float32 `json:"measurement_value" validate:"required"`
	Units string   `json:"units" validate:"required"`
	Area  *uint    `json:"measurement_area" validate:"required"`
}

type CalibrationRecordPatch struct {
	CalibrationLogID *uuid.UUID `json:"calibration_log_id,omitempty"`
	MeasurementValue *float64   `json:"measurement_value,omitempty"`
	MeasurementUnit  *string    `json:"measurement_unit,omitempty"`
	MeasurementArea  *uint      `json:"measurement_area,omitempty"`
}

type CalibrationRecordFilter struct {
	CalibrationLogID *uuid.UUID `json:"calibration_log_id,omitempty" query:"calibration_log_id"`
}
