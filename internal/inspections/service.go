package inspections

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InspectionService struct {
	DB *gorm.DB
}

func NewInspectionService(db *gorm.DB) *InspectionService {
	return &InspectionService{
		DB: db,
	}
}

func (s *InspectionService) CreateInspection(inspection *Inspection) (*Inspection, error) {
	result := s.DB.Create(&inspection)
	if result.Error != nil {
		return nil, result.Error
	}
	return inspection, nil
}

func (s *InspectionService) GetInspections() ([]Inspection, error) {
	var inspections []Inspection
	result := s.DB.Find(&inspections)
	if result.Error != nil {
		return nil, result.Error
	}
	return inspections, nil
}

func (s *InspectionService) GetInspectionByID(inspection_id uuid.UUID) (*Inspection, error) {
	var inspection Inspection
	result := s.DB.Where("id = ?", inspection_id).First(&inspection)
	if result.Error != nil {
		return nil, result.Error
	}
	return &inspection, nil
}

func (s *InspectionService) GetInspectionsByEmployee(employee_id uuid.UUID) ([]Inspection, error) {
	var inspections []Inspection
	result := s.DB.Where(&Inspection{EmployeeID: employee_id}).Find(&inspections)
	if result.Error != nil {
		return nil, result.Error
	}
	return inspections, nil
}
func (s *InspectionService) UpdateInspection(id uuid.UUID, patch InspectionPatch) (*Inspection, error) {
	result := s.DB.Model(&Inspection{}).Where("id = ?", id).Updates(patch)
	if result.Error != nil {
		return nil, result.Error
	}
	var inspection Inspection
	result = s.DB.Where("id = ?", id).First(&inspection)
	if result.Error != nil {
		return nil, result.Error
	}
	return &inspection, nil
}
