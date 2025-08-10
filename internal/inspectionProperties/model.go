package inspectionproperties

import (
	"qc_api/internal/db"

	"github.com/google/uuid"
)

type InspectionProperty struct {
	db.BaseModel
	InspectionID         uuid.UUID `json:"inspection_id"`
	Address              string    `json:"address"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	CustomerEngagement   bool      `json:"customer_engagement"`
	SignsPlacedCorrectly bool      `json:"signs_placed_correctly"`
	ApplicationCoverage  int8      `json:"application_coverage"`
}
