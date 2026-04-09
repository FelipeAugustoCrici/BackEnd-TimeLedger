package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"controltasks/internal/model"
	"controltasks/internal/service"
)

type EntryHandler struct {
	svc         *service.EntryService
	settingsSvc *service.SettingsService
}

func NewEntryHandler(svc *service.EntryService, settingsSvc *service.SettingsService) *EntryHandler {
	return &EntryHandler{svc: svc, settingsSvc: settingsSvc}
}

// userIDFromCtx extrai o user_id dos claims JWT armazenados no contexto Gin.
func userIDFromCtx(c *gin.Context) (string, bool) {
	raw, exists := c.Get("claims")
	if !exists {
		return "", false
	}
	claims, ok := raw.(*model.Claims)
	if !ok || claims.UserID == "" {
		return "", false
	}
	return claims.UserID, true
}

// GET /entries
func (h *EntryHandler) List(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	f := model.EntryFilters{
		UserID:    userID,
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Status:    c.Query("status"),
		Category:  c.Query("category"),
		Project:   c.Query("project"),
		Search:    c.Query("search"),
	}

	// Atalhos de período
	switch c.Query("period") {
	case "today":
		today := time.Now().Format("2006-01-02")
		f.StartDate, f.EndDate = today, today
	case "week":
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := now.AddDate(0, 0, -(weekday - 1))
		sunday := monday.AddDate(0, 0, 6)
		f.StartDate = monday.Format("2006-01-02")
		f.EndDate = sunday.Format("2006-01-02")
	case "month":
		now := time.Now()
		f.StartDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		f.EndDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1).Format("2006-01-02")
	}

	entries, err := h.svc.List(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries, "total": len(entries)})
}

// GET /entries/:id
func (h *EntryHandler) GetByID(c *gin.Context) {
	entry, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if entry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lançamento não encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// POST /entries
func (h *EntryHandler) Create(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	var in model.CreateTaskEntryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	in.UserID = userID

	entry, err := h.svc.Create(in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": entry})
}

// PUT /entries/:id
func (h *EntryHandler) Update(c *gin.Context) {
	var in model.UpdateTaskEntryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.svc.Update(c.Param("id"), in)
	if err == sql.ErrNoRows || entry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "lançamento não encontrado"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entry})
}

// DELETE /entries/:id
func (h *EntryHandler) Delete(c *gin.Context) {
	err := h.svc.Delete(c.Param("id"))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "lançamento não encontrado"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "lançamento removido"})
}

// GET /entries/meta/projects
func (h *EntryHandler) ListProjects(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}
	projects, err := h.svc.ListProjects(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// GET /entries/meta/categories
func (h *EntryHandler) ListCategories(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}
	categories, err := h.svc.ListCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// GET /dashboard
func (h *EntryHandler) Dashboard(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	now := time.Now()
	var startDate, endDate string

	// Aceita start_date/end_date direto ou atalhos de período
	if sd := c.Query("start_date"); sd != "" {
		startDate = sd
		endDate = c.Query("end_date")
	} else {
		switch c.DefaultQuery("period", "month") {
		case "today":
			today := now.Format("2006-01-02")
			startDate, endDate = today, today
		case "month":
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
			endDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).AddDate(0, 1, -1).Format("2006-01-02")
		default: // week
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			monday := now.AddDate(0, 0, -(weekday - 1))
			startDate = monday.Format("2006-01-02")
			endDate = monday.AddDate(0, 0, 6).Format("2006-01-02")
		}
	}

	summary, err := h.svc.Summary(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

// POST /entries/apply-rate
// Re-aplica o hourly_rate atual das settings em todas as entries do usuário,
// recalculando total_amount = (time_spent_minutes / 60) * hourly_rate.
func (h *EntryHandler) ApplyRate(c *gin.Context) {
	userID, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	settings, err := h.settingsSvc.Get(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.svc.ApplyRateToEntries(userID, settings.HourlyRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "hourly_rate aplicado com sucesso",
		"hourly_rate":  settings.HourlyRate,
		"entries_updated": updated,
	})
}
