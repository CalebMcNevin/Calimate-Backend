package calibration

import (
	"qc_api/internal/db"

	"github.com/google/uuid"
)

func Models() []interface{} {
	return []interface{}{
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

type Formulation struct {
	db.BaseModel
	Name string `gorm:"unique;not null" json:"name" validate:"required"` //e.g. "GRANULE"
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

type CalibrationLog struct {
	db.BaseModel
	UserID             uuid.UUID           `json:"user_id"`
	LawnServiceID      uuid.UUID           `json:"lawn_service_id"`
	LawnService        LawnService         `gorm:"foreignKey:LawnServiceID" json:"-"`
	CurrentCalibration *float64            `gorm:"-" json:"current_calibration,omitempty"`
	Records            []CalibrationRecord `json:"records"`
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

type CalibrationRecordFilter struct {
	CalibrationLogID *uuid.UUID
}

type CalibrationRecordDTO struct {
	Value float32 `json:"measurement_value" validate:"required"`
	Units string  `json:"units" validate:"required"`
}
