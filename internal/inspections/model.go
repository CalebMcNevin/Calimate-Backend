package inspections

import (
	"qc_api/internal/db"

	"github.com/google/uuid"
)

func Models() []any {
	return []any{
		&Inspection{},
	}
}

// InspectionDTO represents the data transfer object for creating an inspection.
type InspectionDTO struct {
	Report                 string    `json:"report" validate:"required"`
	UniformPPEGood         bool      `json:"uniform_ppe_good"`
	PICPresent             bool      `json:"pic_present"`
	MotiveLoggedIn         bool      `json:"motive_logged_in"`
	PodiumLoggedIn         bool      `json:"podium_logged_in"`
	SpillAdsorbtionPresent bool      `json:"spill_adsorbtion_present"`
	Calibration            float32   `json:"calibration"`
	EmployeeID             uuid.UUID `json:"employee_id" validate:"required"`
}

// Inspection represents the inspection model.
type Inspection struct {
	db.BaseModel
	Report                 string    `json:"report"`
	UniformPPEGood         bool      `json:"uniform_ppe_good"`
	PICPresent             bool      `json:"pic_present"`
	MotiveLoggedIn         bool      `json:"motive_logged_in"`
	PodiumLoggedIn         bool      `json:"podium_logged_in"`
	SpillAdsorbtionPresent bool      `json:"spill_adsorbtion_present"`
	Calibration            float32   `json:"calibration"`
	EmployeeID             uuid.UUID `gorm:"type:string" json:"employee_id"`
}

// InspectionFilter represents the filter for retrieving inspections.
type InspectionFilter struct {
	EmployeeID *uuid.UUID `json:"employee_id,omitempty" query:"employee_id"`
	Status     *string    `json:"status,omitempty" query:"status"`
	DateFrom   *string    `json:"date_from,omitempty" query:"date_from"`
	DateTo     *string    `json:"date_to,omitempty" query:"date_to"`
}

// InspectionPatch represents the fields that can be updated in an inspection.
type InspectionPatch struct {
	Report                 *string    `json:"report,omitempty"`
	UniformPPEGood         *bool      `json:"uniform_ppe_good,omitempty"`
	PICPresent             *bool      `json:"pic_present,omitempty"`
	MotiveLoggedIn         *bool      `json:"motive_logged_in,omitempty"`
	PodiumLoggedIn         *bool      `json:"podium_logged_in,omitempty"`
	SpillAdsorbtionPresent *bool      `json:"spill_adsorbtion_present,omitempty"`
	Calibration            *float32   `json:"calibration,omitempty"`
	EmployeeID             *uuid.UUID `json:"employee_id,omitempty"`
}
