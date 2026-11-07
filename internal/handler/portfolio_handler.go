package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/middleware"
	"github.com/kiokyky/MnyMngmnt-Go/internal/service"
	"github.com/shopspring/decimal"
)

type PortfolioHandler struct {
	portfolioService *service.PortfolioService
}

func NewPortfolioHandler(portfolioService *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{portfolioService: portfolioService}
}

func (h *PortfolioHandler) ListHoldings(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	assetType := c.Query("asset_type")

	overview, err := h.portfolioService.ListHoldings(userID, assetType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *PortfolioHandler) CreateHolding(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateHoldingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, err := h.portfolioService.CreateHolding(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": view})
}

func (h *PortfolioHandler) GetHoldingDetail(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	var overrideRate *decimal.Decimal
	if rateStr := c.Query("current_rate"); rateStr != "" {
		if parsed, err := decimal.NewFromString(rateStr); err == nil && parsed.GreaterThan(decimal.Zero) {
			overrideRate = &parsed
		}
	}

	view, history, err := h.portfolioService.GetHoldingDetail(userID, holdingID, overrideRate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": view,
		"history": history,
	})
}

func (h *PortfolioHandler) Buy(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	var req dto.BuyAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, tx, err := h.portfolioService.BuyAsset(userID, holdingID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "pembelian valas/aset berhasil dicatat",
		"summary":     view,
		"transaction": tx,
	})
}

func (h *PortfolioHandler) Sell(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	var req dto.SellAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, tx, err := h.portfolioService.SellAsset(userID, holdingID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "penjualan valas/aset berhasil dicatat",
		"summary":     view,
		"transaction": tx,
	})
}

func (h *PortfolioHandler) UpdateRate(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	holdingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	var req dto.UpdateHoldingRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, err := h.portfolioService.UpdateMarketRate(userID, holdingID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "kurs/harga pasar berhasil diperbarui",
		"summary": view,
	})
}

