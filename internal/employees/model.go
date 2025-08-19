package employees

import (
	"errors"
	"qc_api/internal/db"
	"qc_api/internal/inspections"
)

func Models() []any {
	return []any{
		&Employee{},
	}
}

type Employee struct {
	db.BaseModel
	CommonName     string                   `gorm:"not null" json:"common_name"`
	FirstName      string                   `gorm:"not null" json:"first_name"`
	LastName       string                   `gorm:"not null" json:"last_name"`
	EmployeeNumber string                   `gorm:"" json:"employee_number"`
	Active         bool                     `gorm:"default:true" json:"active"`
	Inspections    []inspections.Inspection `gorm:"foreignKey:EmployeeID" json:"-"`
}

// EmployeeDTO represents the data transfer object for creating an employee.
type EmployeeDTO struct {
	CommonName     string `json:"common_name" validate:"required"`
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	EmployeeNumber string `json:"employee_number" validate:"required"`
}

type EmployeePatch struct {
	CommonName     *string `json:"common_name,omitempty"`
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	EmployeeNumber *string `json:"employee_number,omitempty"`
	Active         *bool   `json:"active,omitempty"`
}

type EmployeeFilter struct {
	Active         *bool   `json:"active,omitempty" query:"active"`
	EmployeeNumber *string `json:"employee_number,omitempty" query:"employee_number"`
}

func NewEmployee(CommonName, FirstName, LastName, EmployeeNumber string) (*Employee, error) {
	// Business logic validation
	if len(EmployeeNumber) > 0 && len(EmployeeNumber) < 3 {
		return nil, errors.New("employee number must be at least 3 characters long")
	}

	if len(EmployeeNumber) > 50 {
		return nil, errors.New("employee number cannot exceed 50 characters")
	}

	// Check for invalid characters in employee number
	if EmployeeNumber != "" {
		for _, char := range EmployeeNumber {
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				return nil, errors.New("employee number can only contain alphanumeric characters")
			}
		}
	}

	// Name length validation
	if len(CommonName) > 100 {
		return nil, errors.New("common name cannot exceed 100 characters")
	}

	if len(FirstName) > 50 {
		return nil, errors.New("first name cannot exceed 50 characters")
	}

	if len(LastName) > 50 {
		return nil, errors.New("last name cannot exceed 50 characters")
	}

	employee := &Employee{
		CommonName:     CommonName,
		FirstName:      FirstName,
		LastName:       LastName,
		EmployeeNumber: EmployeeNumber,
		Active:         true,
	}
	return employee, nil
}
