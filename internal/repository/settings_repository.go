package repository

import (
	"database/sql"
	"fmt"

	"controltasks/internal/crypto"
	"controltasks/internal/model"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retorna as configurações do usuário, criando um registro padrão se não existir.
func (r *SettingsRepository) Get(userID string) (*model.UserSettings, error) {
	var s model.UserSettings
	var hourlyEnc, goalEnc, monthlyEnc string

	err := r.db.QueryRow(`
		SELECT id, hourly_rate, daily_hours_goal, monthly_goal, updated_at
		FROM user_settings WHERE user_id = $1`, userID,
	).Scan(&s.ID, &hourlyEnc, &goalEnc, &monthlyEnc, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		// Cria settings padrão para o novo usuário
		defaultHourly, _ := crypto.EncryptFloat64(0)
		defaultGoal, _ := crypto.EncryptFloat64(8)
		defaultMonthly, _ := crypto.EncryptFloat64(0)
		err = r.db.QueryRow(`
			INSERT INTO user_settings (user_id, hourly_rate, daily_hours_goal, monthly_goal)
			VALUES ($1, $2, $3, $4)
			RETURNING id, hourly_rate, daily_hours_goal, monthly_goal, updated_at`,
			userID, defaultHourly, defaultGoal, defaultMonthly,
		).Scan(&s.ID, &hourlyEnc, &goalEnc, &monthlyEnc, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	if s.HourlyRate, err = crypto.DecryptFloat64(hourlyEnc); err != nil {
		return nil, fmt.Errorf("decrypt hourly_rate: %w", err)
	}
	if s.DailyHoursGoal, err = crypto.DecryptFloat64(goalEnc); err != nil {
		return nil, fmt.Errorf("decrypt daily_hours_goal: %w", err)
	}
	if monthlyEnc != "" {
		if s.MonthlyGoal, err = crypto.DecryptFloat64(monthlyEnc); err != nil {
			return nil, fmt.Errorf("decrypt monthly_goal: %w", err)
		}
	}

	return &s, nil
}

// Update salva as configurações do usuário criptografando os campos financeiros.
func (r *SettingsRepository) Update(userID string, in model.UpdateSettingsInput) (*model.UserSettings, error) {
	hourlyEnc, err := crypto.EncryptFloat64(in.HourlyRate)
	if err != nil {
		return nil, fmt.Errorf("encrypt hourly_rate: %w", err)
	}
	goalEnc, err := crypto.EncryptFloat64(in.DailyHoursGoal)
	if err != nil {
		return nil, fmt.Errorf("encrypt daily_hours_goal: %w", err)
	}
	monthlyEnc, err := crypto.EncryptFloat64(in.MonthlyGoal)
	if err != nil {
		return nil, fmt.Errorf("encrypt monthly_goal: %w", err)
	}

	var s model.UserSettings
	var hourlyEncOut, goalEncOut, monthlyEncOut string

	err = r.db.QueryRow(`
		INSERT INTO user_settings (user_id, hourly_rate, daily_hours_goal, monthly_goal)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		  SET hourly_rate=$2, daily_hours_goal=$3, monthly_goal=$4, updated_at=NOW()
		RETURNING id, hourly_rate, daily_hours_goal, monthly_goal, updated_at`,
		userID, hourlyEnc, goalEnc, monthlyEnc,
	).Scan(&s.ID, &hourlyEncOut, &goalEncOut, &monthlyEncOut, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if s.HourlyRate, err = crypto.DecryptFloat64(hourlyEncOut); err != nil {
		return nil, fmt.Errorf("decrypt hourly_rate: %w", err)
	}
	if s.DailyHoursGoal, err = crypto.DecryptFloat64(goalEncOut); err != nil {
		return nil, fmt.Errorf("decrypt daily_hours_goal: %w", err)
	}
	if monthlyEncOut != "" {
		if s.MonthlyGoal, err = crypto.DecryptFloat64(monthlyEncOut); err != nil {
			return nil, fmt.Errorf("decrypt monthly_goal: %w", err)
		}
	}

	return &s, nil
}
