package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"controltasks/internal/crypto"
	"controltasks/internal/model"
)

type EntryRepository struct {
	db *sql.DB
}

func NewEntryRepository(db *sql.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

// ─── helpers de criptografia ──────────────────────────────────────────────────

func encryptRate(v float64) (string, error) {
	return crypto.EncryptFloat64(v)
}

func decryptRate(s string) (float64, error) {
	return crypto.DecryptFloat64(s)
}

// scanEntry lê uma linha do banco e descriptografa os campos financeiros.
func scanEntry(row interface {
	Scan(...any) error
}) (*model.TaskEntry, error) {
	var e model.TaskEntry
	var hourlyRateEnc, totalAmountEnc string

	if err := row.Scan(
		&e.ID, &e.Date, &e.TaskCode, &e.Description,
		&e.TimeSpentMinutes, &hourlyRateEnc, &totalAmountEnc,
		&e.Status, &e.Category, &e.Project, &e.Notes,
		&e.StartTime, &e.EndTime,
		&e.CreatedAt, &e.UpdatedAt,
	); err != nil {
		return nil, err
	}

	var err error
	if e.HourlyRate, err = decryptRate(hourlyRateEnc); err != nil {
		return nil, fmt.Errorf("decrypt hourly_rate: %w", err)
	}
	if e.TotalAmount, err = decryptRate(totalAmountEnc); err != nil {
		return nil, fmt.Errorf("decrypt total_amount: %w", err)
	}

	return &e, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (r *EntryRepository) List(f model.EntryFilters) ([]model.TaskEntry, error) {
	query := `
		SELECT id, date::text, task_code, description, time_spent_minutes,
		       hourly_rate, total_amount, status, category, project, notes,
		       start_time, end_time, created_at, updated_at
		FROM task_entries WHERE user_id = $1`

	args := []any{f.UserID}
	i := 2

	if f.StartDate != "" {
		query += fmt.Sprintf(" AND date >= $%d", i); args = append(args, f.StartDate); i++
	}
	if f.EndDate != "" {
		query += fmt.Sprintf(" AND date <= $%d", i); args = append(args, f.EndDate); i++
	}
	if f.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", i); args = append(args, f.Status); i++
	}
	if f.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", i); args = append(args, f.Category); i++
	}
	if f.Project != "" {
		query += fmt.Sprintf(" AND project = $%d", i); args = append(args, f.Project); i++
	}
	if f.Search != "" {
		query += fmt.Sprintf(" AND (task_code ILIKE $%d OR description ILIKE $%d)", i, i+1)
		like := "%" + strings.TrimSpace(f.Search) + "%"
		args = append(args, like, like)
		i += 2
	}
	query += " ORDER BY date DESC, created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.TaskEntry
	for rows.Next() {
		e, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, *e)
	}
	return entries, rows.Err()
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func (r *EntryRepository) GetByID(id string) (*model.TaskEntry, error) {
	row := r.db.QueryRow(`
		SELECT id, date::text, task_code, description, time_spent_minutes,
		       hourly_rate, total_amount, status, category, project, notes,
		       start_time, end_time, created_at, updated_at
		FROM task_entries WHERE id = $1`, id)

	e, err := scanEntry(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return e, err
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (r *EntryRepository) Create(in model.CreateTaskEntryInput) (*model.TaskEntry, error) {
	totalAmount := (float64(in.TimeSpentMinutes) / 60.0) * in.HourlyRate

	hourlyEnc, err := encryptRate(in.HourlyRate)
	if err != nil {
		return nil, fmt.Errorf("encrypt hourly_rate: %w", err)
	}
	totalEnc, err := encryptRate(totalAmount)
	if err != nil {
		return nil, fmt.Errorf("encrypt total_amount: %w", err)
	}

	row := r.db.QueryRow(`
		INSERT INTO task_entries
		  (user_id, date, task_code, description, time_spent_minutes, hourly_rate,
		   total_amount, status, category, project, notes, start_time, end_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id, date::text, task_code, description, time_spent_minutes,
		          hourly_rate, total_amount, status, category, project, notes,
		          start_time, end_time, created_at, updated_at`,
		in.UserID, in.Date, in.TaskCode, in.Description, in.TimeSpentMinutes,
		hourlyEnc, totalEnc, in.Status,
		in.Category, in.Project, in.Notes, in.StartTime, in.EndTime,
	)
	return scanEntry(row)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (r *EntryRepository) Update(id string, in model.UpdateTaskEntryInput) (*model.TaskEntry, error) {
	current, err := r.GetByID(id)
	if err != nil || current == nil {
		return nil, err
	}

	if in.Date != nil            { current.Date = *in.Date }
	if in.TaskCode != nil        { current.TaskCode = *in.TaskCode }
	if in.Description != nil     { current.Description = *in.Description }
	if in.Status != nil          { current.Status = *in.Status }
	if in.Category != nil        { current.Category = in.Category }
	if in.Project != nil         { current.Project = in.Project }
	if in.Notes != nil           { current.Notes = in.Notes }
	if in.TimeSpentMinutes != nil { current.TimeSpentMinutes = *in.TimeSpentMinutes }
	if in.HourlyRate != nil      { current.HourlyRate = *in.HourlyRate }
	if in.StartTime != nil       { current.StartTime = in.StartTime }
	if in.EndTime != nil         { current.EndTime = in.EndTime }

	current.TotalAmount = (float64(current.TimeSpentMinutes) / 60.0) * current.HourlyRate

	hourlyEnc, err := encryptRate(current.HourlyRate)
	if err != nil {
		return nil, fmt.Errorf("encrypt hourly_rate: %w", err)
	}
	totalEnc, err := encryptRate(current.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("encrypt total_amount: %w", err)
	}

	row := r.db.QueryRow(`
		UPDATE task_entries SET
		  date=$1, task_code=$2, description=$3, time_spent_minutes=$4,
		  hourly_rate=$5, total_amount=$6, status=$7, category=$8,
		  project=$9, notes=$10, start_time=$11, end_time=$12, updated_at=NOW()
		WHERE id=$13
		RETURNING id, date::text, task_code, description, time_spent_minutes,
		          hourly_rate, total_amount, status, category, project, notes,
		          start_time, end_time, created_at, updated_at`,
		current.Date, current.TaskCode, current.Description,
		current.TimeSpentMinutes, hourlyEnc, totalEnc,
		current.Status, current.Category, current.Project, current.Notes,
		current.StartTime, current.EndTime, id,
	)
	return scanEntry(row)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (r *EntryRepository) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM task_entries WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ─── Meta ─────────────────────────────────────────────────────────────────────

func (r *EntryRepository) ListProjects(userID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT project FROM task_entries WHERE project IS NOT NULL AND user_id = $1 ORDER BY project`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *EntryRepository) ListCategories(userID string) ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT category FROM task_entries WHERE category IS NOT NULL AND user_id = $1 ORDER BY category`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// ─── Summary ──────────────────────────────────────────────────────────────────
// Como os valores estão criptografados, o SUM() no SQL não é possível.
// Buscamos os registros e somamos em Go após descriptografar.

func (r *EntryRepository) Summary(userID, startDate, endDate string) (*model.DashboardSummary, error) {
	var billableMinutes int
	var totalAmount float64
	var totalCount int
	uniqueDays := map[string]struct{}{}

	// Busca todos os lançamentos do período com flag billable
	rows, err := r.db.Query(`
		SELECT te.date::text, te.time_spent_minutes, te.total_amount,
		       COALESCE(c.billable, true) as billable
		FROM task_entries te
		LEFT JOIN categories c ON c.name = te.category
		WHERE te.user_id = $1
		  AND te.date BETWEEN $2 AND $3`,
		userID, startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var date, totalAmountEnc string
		var minutes int
		var billable bool
		if err := rows.Scan(&date, &minutes, &totalAmountEnc, &billable); err != nil {
			return nil, err
		}
		totalCount++
		uniqueDays[date] = struct{}{}

		if !billable {
			continue
		}

		amount, err := decryptRate(totalAmountEnc)
		if err != nil {
			return nil, fmt.Errorf("decrypt total_amount in summary: %w", err)
		}
		billableMinutes += minutes
		totalAmount += amount
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	s := &model.DashboardSummary{
		TotalHours:  float64(billableMinutes) / 60.0,
		TotalAmount: totalAmount,
		TotalTasks:  totalCount,
	}
	if len(uniqueDays) > 0 {
		s.AvgHoursPerDay = s.TotalHours / float64(len(uniqueDays))
	}
	return s, nil
}

// ApplyRateToEntries re-aplica um hourly_rate em todas as entries do usuário,
// recalculando total_amount = (time_spent_minutes / 60) * hourlyRate.
func (r *EntryRepository) ApplyRateToEntries(userID string, hourlyRate float64) (int, error) {
	rows, err := r.db.Query(
		`SELECT id, time_spent_minutes FROM task_entries WHERE user_id = $1`, userID,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type row struct {
		id      string
		minutes int
	}
	var records []row
	for rows.Next() {
		var rec row
		if err := rows.Scan(&rec.id, &rec.minutes); err != nil {
			return 0, err
		}
		records = append(records, rec)
	}

	updated := 0
	for _, rec := range records {
		totalAmount := (float64(rec.minutes) / 60.0) * hourlyRate
		hourlyEnc, err := encryptRate(hourlyRate)
		if err != nil {
			return updated, fmt.Errorf("encrypt hourly_rate: %w", err)
		}
		totalEnc, err := encryptRate(totalAmount)
		if err != nil {
			return updated, fmt.Errorf("encrypt total_amount: %w", err)
		}
		_, err = r.db.Exec(
			`UPDATE task_entries SET hourly_rate=$1, total_amount=$2, updated_at=NOW() WHERE id=$3`,
			hourlyEnc, totalEnc, rec.id,
		)
		if err != nil {
			return updated, fmt.Errorf("update entry %s: %w", rec.id, err)
		}
		updated++
	}
	return updated, nil
}
