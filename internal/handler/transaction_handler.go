package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/middleware"
	"github.com/kiokyky/MnyMngmnt-Go/internal/service"
)

type TransactionHandler struct {
	txService *service.TransactionService
}

func NewTransactionHandler(txService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{txService: txService}
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var filter dto.TransactionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	summary, err := h.txService.ListTransactions(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.txService.CreateTransaction(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": tx})
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	if err := h.txService.DeleteTransaction(userID, txID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "transaksi berhasil dihapus dan saldo kantong dikembalikan"})
}

