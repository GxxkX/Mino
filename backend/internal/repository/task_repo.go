package repository

import (
	"database/sql"
	"fmt"

	"github.com/mino/backend/internal/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) List(userID string, limit, offset int) ([]*model.Task, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM mino_tasks WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(
		`SELECT id, user_id, conversation_id, title, description, status, priority, due_date, completed_at, created_at, updated_at
		 FROM mino_tasks WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*model.Task
	for rows.Next() {
		t := &model.Task{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.ConversationID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepository) FindByID(id, userID string) (*model.Task, error) {
	t := &model.Task{}
	err := r.db.QueryRow(
		`SELECT id, user_id, conversation_id, title, description, status, priority, due_date, completed_at, created_at, updated_at
		 FROM mino_tasks WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&t.ID, &t.UserID, &t.ConversationID, &t.Title, &t.Description,
		&t.Status, &t.Priority, &t.DueDate, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (r *TaskRepository) Create(t *model.Task) error {
	return r.db.QueryRow(
		`INSERT INTO mino_tasks (user_id, conversation_id, title, description, status, priority, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, created_at, updated_at`,
		t.UserID, t.ConversationID, t.Title, t.Description, t.Status, t.Priority, t.DueDate,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TaskRepository) Update(t *model.Task) error {
	res, err := r.db.Exec(
		`UPDATE mino_tasks SET title=$1, description=$2, status=$3, priority=$4, due_date=$5, completed_at=$6, updated_at=NOW()
		 WHERE id=$7 AND user_id=$8`,
		t.Title, t.Description, t.Status, t.Priority, t.DueDate, t.CompletedAt, t.ID, t.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *TaskRepository) Delete(id, userID string) error {
	res, err := r.db.Exec(`DELETE FROM mino_tasks WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (r *TaskRepository) CreateBatch(tasks []*model.Task) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO mino_tasks (user_id, conversation_id, title, description, status, priority)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at, updated_at`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range tasks {
		if err := stmt.QueryRow(t.UserID, t.ConversationID, t.Title, t.Description, t.Status, t.Priority).
			Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}
