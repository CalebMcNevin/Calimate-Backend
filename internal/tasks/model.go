package tasks

import (
	"qc_api/internal/db"
)

type Task struct {
	db.BaseModel
	Description  string `json:"description"`
	Notes        string `json:"notes"`
	InspectionID string `json:"inspection_id"`
}
