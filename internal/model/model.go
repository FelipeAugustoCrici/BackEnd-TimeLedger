package model

import "time"

// TaskStatus representa o estado de um lançamento.
type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

// TaskEntry é a entidade principal — um apontamento de horas.
type TaskEntry struct {
	ID                string     `json:"id"`
	Date              string     `json:"date"`               // "YYYY-MM-DD"
	TaskCode          string     `json:"task_code"`
	Description       string     `json:"description"`
	TimeSpentMinutes  int        `json:"time_spent_minutes"`
	HourlyRate        float64    `json:"hourly_rate"`
	TotalAmount       float64    `json:"total_amount"`
	Status            TaskStatus `json:"status"`
	Category          *string    `json:"category,omitempty"`
	Project           *string    `json:"project,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
	StartTime         *string    `json:"start_time,omitempty"` // "HH:MM"
	EndTime           *string    `json:"end_time,omitempty"`   // "HH:MM"
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateTaskEntryInput é o payload de criação.
type CreateTaskEntryInput struct {
	UserID           string     `json:"-"` // preenchido pelo handler via JWT
	Date             string     `json:"date"               binding:"required"`
	TaskCode         string     `json:"task_code"          binding:"required"`
	Description      string     `json:"description"        binding:"required"`
	TimeSpentMinutes int        `json:"time_spent_minutes" binding:"required,min=1"`
	HourlyRate       float64    `json:"hourly_rate"        binding:"required,gt=0"`
	Status           TaskStatus `json:"status"             binding:"required"`
	Category         *string    `json:"category"`
	Project          *string    `json:"project"`
	Notes            *string    `json:"notes"`
	StartTime        *string    `json:"start_time"`
	EndTime          *string    `json:"end_time"`
}

// UpdateTaskEntryInput é o payload de atualização (todos opcionais).
type UpdateTaskEntryInput struct {
	Date             *string     `json:"date"`
	TaskCode         *string     `json:"task_code"`
	Description      *string     `json:"description"`
	TimeSpentMinutes *int        `json:"time_spent_minutes"`
	HourlyRate       *float64    `json:"hourly_rate"`
	Status           *TaskStatus `json:"status"`
	Category         *string     `json:"category"`
	Project          *string     `json:"project"`
	Notes            *string     `json:"notes"`
	StartTime        *string     `json:"start_time"`
	EndTime          *string     `json:"end_time"`
}

// EntryFilters são os filtros de listagem.
type EntryFilters struct {
	UserID    string
	StartDate string
	EndDate   string
	Status    string
	Category  string
	Project   string
	Search    string
	Page      int
	PerPage   int
	SortField string
	SortDir   string
}

// UserSettings configurações do usuário (singleton).
type UserSettings struct {
	ID                  string    `json:"id"`
	HourlyRate          float64   `json:"hourly_rate"`
	DailyHoursGoal      float64   `json:"daily_hours_goal"`
	MonthlyGoal         float64   `json:"monthly_goal"`
	DefaultCategoryName *string   `json:"default_category_name,omitempty"`
	CategoryCodes       *string   `json:"category_codes,omitempty"` // JSON string
	UpdatedAt           time.Time `json:"updated_at"`
}

// UpdateSettingsInput payload para atualizar configurações.
type UpdateSettingsInput struct {
	HourlyRate          float64 `json:"hourly_rate"      binding:"required,gt=0"`
	DailyHoursGoal      float64 `json:"daily_hours_goal" binding:"required,gt=0"`
	MonthlyGoal         float64 `json:"monthly_goal"`
	DefaultCategoryName *string `json:"default_category_name"`
	CategoryCodes       *string `json:"category_codes"` // JSON string
}

// Category representa uma categoria de lançamento com cor.
type Category struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`    // hex ex: "#3b82f6"
	Billable bool   `json:"billable"` // false = não entra no cálculo de horas
}

// CreateCategoryInput payload de criação.
type CreateCategoryInput struct {
	Name     string `json:"name"     binding:"required"`
	Color    string `json:"color"    binding:"required"`
	Billable *bool  `json:"billable"` // default true se omitido
}

// DashboardSummary resumo calculado para o dashboard.
type DashboardSummary struct {
	TotalHours     float64 `json:"total_hours"`
	TotalAmount    float64 `json:"total_amount"`
	TotalTasks     int     `json:"total_tasks"`
	AvgHoursPerDay float64 `json:"avg_hours_per_day"`
}

// PaginatedResponse resposta paginada genérica.
type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
}
