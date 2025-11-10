package repository

import (
	"context"
	"database/sql"

	"github.com/joshjms/castletown/internal/models"
	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dbUrl string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) GetProblemDetails(ctx context.Context, problemID int64) (*models.Problem, error) {
	var p models.Problem
	err := r.db.QueryRowContext(ctx, "SELECT id, time_limit, memory_limit FROM problems WHERE id = $1", problemID).Scan(&p.ID, &p.TimeLimit, &p.MemoryLimit)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) GetTestcases(ctx context.Context, problemID int64) ([]models.Testcase, error) {
	var testcases []models.Testcase
	rows, err := r.db.QueryContext(ctx, "SELECT id, input, output FROM testcases WHERE problem_id = $1", problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tc models.Testcase
		if err := rows.Scan(&tc.ID, &tc.Input, &tc.Output); err != nil {
			return nil, err
		}
		testcases = append(testcases, tc)
	}
	return testcases, nil
}
