package employees

import (
	"errors"
	"qc_api/internal/utils"

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

func (s *EmployeeService) GetEmployees(filter EmployeeFilter) ([]Employee, error) {
	var employees []Employee
	query := utils.ApplyFilter(s.DB.Model(&Employee{}), filter)
	return employees, query.Preload("Inspections").Find(&employees).Error
}

func (s *EmployeeService) GetEmployeeByID(id string) (Employee, error) {
	var employee Employee
	err := s.DB.First(&employee, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Employee{}, err
	}
	return employee, err
}

func (s *EmployeeService) UpdateEmployee(id string, patch EmployeePatch) (*Employee, error) {
	result := s.DB.Model(&Employee{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var employee Employee
	result = s.DB.Where("id = ?", id).First(&employee)
	if result.Error != nil {
		return nil, result.Error
	}
	return &employee, nil
}
