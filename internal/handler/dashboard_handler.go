package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kiokyky/MnyMngmnt-Go/internal/middleware"
	"github.com/kiokyky/MnyMngmnt-Go/internal/service"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	monthStr := c.Query("month") // opsional: YYYY-MM

	summary, err := h.dashboardService.GetDashboardSummary(userID, monthStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

