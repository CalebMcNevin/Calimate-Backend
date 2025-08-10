package employees

import (
	"errors"

	"gorm.io/gorm"
)

type EmployeeService struct {
	DB *gorm.DB
}

func NewEmployeeService(db *gorm.DB) EmployeeService {
	return EmployeeService{db}
}

func (s *EmployeeService) CreateEmployee(employee *Employee) error {
	return s.DB.Create(employee).Error
}

func (s *EmployeeService) GetEmployees() ([]Employee, error) {
	var employees []Employee
	err := s.DB.Find(&employees).Error
	return employees, err
}

func (s *EmployeeService) GetEmployeeByID(id string) (Employee, error) {
	var employee Employee
	err := s.DB.First(&employee, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Employee{}, err
	}
	return employee, err
}
