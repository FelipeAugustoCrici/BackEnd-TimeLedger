package repository

import (
	"database/sql"
	"controltasks/internal/model"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) List() ([]model.Category, error) {
	rows, err := r.db.Query(`SELECT id, name, color, billable FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.Billable); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *CategoryRepository) Create(in model.CreateCategoryInput) (*model.Category, error) {
	billable := true
	if in.Billable != nil {
		billable = *in.Billable
	}
	var c model.Category
	err := r.db.QueryRow(
		`INSERT INTO categories (name, color, billable) VALUES ($1, $2, $3)
		 ON CONFLICT (name) DO UPDATE SET color = EXCLUDED.color, billable = EXCLUDED.billable
		 RETURNING id, name, color, billable`,
		in.Name, in.Color, billable,
	).Scan(&c.ID, &c.Name, &c.Color, &c.Billable)
	return &c, err
}

func (r *CategoryRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM categories WHERE id = $1`, id)
	return err
}
