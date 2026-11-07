package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/middleware"
	"github.com/kiokyky/MnyMngmnt-Go/internal/service"
)

type PocketHandler struct {
	pocketService *service.PocketService
}

func NewPocketHandler(pocketService *service.PocketService) *PocketHandler {
	return &PocketHandler{pocketService: pocketService}
}

func (h *PocketHandler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	pockets, err := h.pocketService.ListPockets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pockets})
}

func (h *PocketHandler) Get(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	pocketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	pocket, err := h.pocketService.GetPocket(userID, pocketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pocket})
}

func (h *PocketHandler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreatePocketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pocket, err := h.pocketService.CreatePocket(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": pocket})
}

func (h *PocketHandler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	pocketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	var req dto.UpdatePocketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pocket, err := h.pocketService.UpdatePocket(userID, pocketID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": pocket})
}

func (h *PocketHandler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	pocketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format ID tidak valid"})
		return
	}

	if err := h.pocketService.DeletePocket(userID, pocketID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "kantong berhasil dihapus"})
}

