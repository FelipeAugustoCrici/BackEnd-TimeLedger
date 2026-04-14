package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"controltasks/internal/model"
	"controltasks/internal/service"
)

type ReportHandler struct {
	entrySvc *service.EntryService
}

func NewReportHandler(entrySvc *service.EntryService) *ReportHandler {
	return &ReportHandler{entrySvc: entrySvc}
}

// userIDFromCtx extrai o user_id dos claims JWT armazenados no contexto Gin.
func (h *ReportHandler) userIDFromCtx(c *gin.Context) (string, bool) {
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

// buildFilters constrói os filtros baseado nos parâmetros da query
func (h *ReportHandler) buildFilters(c *gin.Context, userID string) model.EntryFilters {
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

	return f
}

// GET /reports/entries - Lista todas as entries para cálculos de relatório (sem paginação)
func (h *ReportHandler) GetEntries(c *gin.Context) {
	userID, ok := h.userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	f := h.buildFilters(c, userID)

	entries, err := h.entrySvc.List(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": entries, "total": len(entries)})
}

// GET /reports/entries-paginated - Lista entries paginadas para tabela de detalhes
func (h *ReportHandler) GetEntriesPaginated(c *gin.Context) {
	userID, ok := h.userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	f := h.buildFilters(c, userID)

	// Parâmetros de paginação
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			f.Page = p
		}
	}
	if perPage := c.Query("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 1000 {
			f.PerPage = pp
		}
	}
	if f.Page == 0 {
		f.Page = 1
	}
	if f.PerPage == 0 {
		f.PerPage = 10
	}

	// Parâmetros de ordenação
	f.SortField = c.DefaultQuery("sort_field", "date")
	f.SortDir = c.DefaultQuery("sort_dir", "desc")

	entries, total, err := h.entrySvc.ListPaginated(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := model.PaginatedResponse{
		Data:    entries,
		Total:   total,
		Page:    f.Page,
		PerPage: f.PerPage,
	}
	c.JSON(http.StatusOK, response)
}