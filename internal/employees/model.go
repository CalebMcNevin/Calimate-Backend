package employees

import (
	"qc_api/internal/db"
	"qc_api/internal/inspections"
)

type Employee struct {
	db.BaseModel
	CommonName     string                   `gorm:"not null" json:"common_name"`
	FirstName      string                   `gorm:"not null" json:"first_name"`
	LastName       string                   `gorm:"not null" json:"last_name"`
	EmployeeNumber string                   `gorm:"" json:"employee_number"`
	Active         bool                     `gorm:"default:true" json:"active"`
	Inspections    []inspections.Inspection `gorm:"foreignKey:EmployeeID"`
}

func Models() []interface{} {
	return []interface{}{
		&Employee{},
	}
}

func NewEmployee(CommonName, FirstName, LastName, EmployeeNumber string) (*Employee, error) {
	employee := &Employee{
		CommonName:     CommonName,
		FirstName:      FirstName,
		LastName:       LastName,
		EmployeeNumber: EmployeeNumber,
		Active:         true,
	}
	return employee, nil
}
