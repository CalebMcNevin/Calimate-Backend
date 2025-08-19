package calibration

import (
	"qc_api/internal/db"
	"time"

	"github.com/google/uuid"
)

func Models() []any {
	return []any{
		&Unit{},
		&Formulation{},
		&LawnService{},
		&CalibrationLog{},
		&CalibrationRecord{},
	}
}

type Unit struct {
	Symbol      string `gorm:"primaryKey" json:"symbol" validate:"required"`    //e.g. "kg"
	Description string `gorm:"not null" json:"description" validate:"required"` //e.g. "Kilogram"
}

type UnitPatch struct {
	Symbol      *string `json:"symbol,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UnitDTO struct {
	Symbol      string `json:"symbol" validate:"required"`
	Description string `json:"description" validate:"required"`
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
	Code                      string      `gorm:"unique;not null" json:"code" validate:"required"` //e.g. "LS01"
	Description               string      `json:"description" validate:"required"`                 //e.g. "Spring Fertilizer"
	FormulationID             uuid.UUID   `json:"formulation_id" validate:"required"`
	Formulation               Formulation `gorm:"foreignKey:FormulationID" json:"formulation"`
	TargetCalibrationValue    float32     `json:"target_calibration_value" validate:"required"`
	TargetCalibrationUnitCode string      `gorm:"not null" json:"target_calibration_unit_code" validate:"required"`
	TargetCalibrationUnit     Unit        `gorm:"foreignKey:TargetCalibrationUnitCode;references:Symbol;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	CalibrationFunction       string      `json:"calibration_function" validate:"required"`
}

type LawnServiceDTO struct {
	Code                      string    `json:"code" validate:"required"`
	Description               string    `json:"description" validate:"required"`
	FormulationID             uuid.UUID `json:"formulation_id" validate:"required"`
	TargetCalibrationValue    float32   `json:"target_calibration_value" validate:"required"`
	TargetCalibrationUnitCode string    `json:"target_calibration_unit_code" validate:"required"`
	CalibrationFunction       string    `json:"calibration_function" validate:"required"`
}

type LawnServicePatch struct {
	Code                      *string    `json:"code,omitempty"`
	Description               *string    `json:"description,omitempty"`
	FormulationID             *uuid.UUID `json:"formulation_id,omitempty"`
	TargetCalibrationValue    *float32   `json:"target_calibration_value,omitempty"`
	TargetCalibrationUnitCode *string    `json:"target_calibration_unit_code,omitempty"`
	CalibrationFunction       *string    `json:"calibration_function,omitempty"`
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
	UserID        *uuid.UUID `json:"user_id,omitempty" query:"user_id"`
	LawnServiceID *uuid.UUID `json:"lawn_service_id,omitempty" query:"lawn_service_id"`
	DateFrom      *time.Time `json:"date_from,omitempty" query:"date_from"`
	DateTo        *time.Time `json:"date_to,omitempty" query:"date_to"`
}

type CalibrationRecord struct {
	db.BaseModel
	CalibrationLogID uuid.UUID      `json:"calibration_log_id"`
	CalibrationLog   CalibrationLog `gorm:"foreignKey:CalibrationLogID" json:"-"`
	MeasurementValue float64        `json:"measurement_value"`
	MeasurementUnit  string         `json:"measurement_unit"`
	MeasurementArea  uint           `json:"measurement_area"`
	UnitSymbol       string         `json:"unit_symbol"`
	Unit             Unit           `gorm:"foreignKey:UnitSymbol;references:Symbol" json:"-"`
	Calibration      float64        `gorm:"-" json:"calibration"`
}

type CalibrationRecordDTO struct {
	Value float32 `json:"measurement_value" validate:"required"`
	Units string  `json:"units" validate:"required"`
}

type CalibrationRecordPatch struct {
	CalibrationLogID *uuid.UUID `json:"calibration_log_id,omitempty"`
	MeasurementValue *float64   `json:"measurement_value,omitempty"`
	MeasurementUnit  *string    `json:"measurement_unit,omitempty"`
	MeasurementArea  *uint      `json:"measurement_area,omitempty"`
	UnitSymbol       *string    `json:"unit_symbol,omitempty"`
}

type CalibrationRecordFilter struct {
	CalibrationLogID *uuid.UUID `json:"calibration_log_id,omitempty" query:"calibration_log_id"`
}
