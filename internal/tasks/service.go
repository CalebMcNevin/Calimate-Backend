package tasks

import (
	"github.com/jmoiron/sqlx"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type TaskService struct {
	DB *sqlx.DB
}

func NewTaskService(db *sqlx.DB) *TaskService {
	return &TaskService{
		DB: db,
	}
}

func (s *TaskService) CreateTask(task Task) error {
	query := `INSERT INTO tasks (id, description, notes, inpection_id) VALUES (?,?,?,?);`
	_, err := s.DB.Exec(query, task.ID, task.Description, task.Notes, task.InspectionID)
	if err != nil {
		log.Errorf("Error inserting new task: %v", err)
		return err
	}
	return nil
}

func (s *TaskService) GetTaskByID(taskID uuid.UUID) {
	query := `SELECT * FROM tasks WHERE taskID = ?;`
	s.DB.QueryRowx(query, taskID)
}

func (s *TaskService) GetTasks() {}

func (s *TaskService) UpdateTask() {}

func (s *TaskService) DeleteTask() {}
