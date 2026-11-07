package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"gorm.io/gorm"
)

type PocketService struct {
	db *gorm.DB
}

func NewPocketService(db *gorm.DB) *PocketService {
	return &PocketService{db: db}
}

func (s *PocketService) ListPockets(userID uuid.UUID) ([]models.Pocket, error) {
	var pockets []models.Pocket
	if err := s.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&pockets).Error; err != nil {
		return nil, err
	}
	return pockets, nil
}

func (s *PocketService) GetPocket(userID, pocketID uuid.UUID) (*models.Pocket, error) {
	var pocket models.Pocket
	if err := s.db.Where("id = ? AND user_id = ?", pocketID, userID).First(&pocket).Error; err != nil {
		return nil, errors.New("kantong tidak ditemukan")
	}
	return &pocket, nil
}

func (s *PocketService) CreatePocket(userID uuid.UUID, req dto.CreatePocketRequest) (*models.Pocket, error) {
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "IDR"
	}

	pocket := models.Pocket{
		UserID:   userID,
		Name:     strings.TrimSpace(req.Name),
		Type:     req.Type,
		Currency: currency,
		Balance:  req.InitialBalance,
	}

	if err := s.db.Create(&pocket).Error; err != nil {
		return nil, err
	}
	return &pocket, nil
}

func (s *PocketService) UpdatePocket(userID, pocketID uuid.UUID, req dto.UpdatePocketRequest) (*models.Pocket, error) {
	pocket, err := s.GetPocket(userID, pocketID)
	if err != nil {
		return nil, err
	}

	pocket.Name = strings.TrimSpace(req.Name)
	pocket.Type = req.Type
	if req.Currency != "" {
		pocket.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	}

	if err := s.db.Save(pocket).Error; err != nil {
		return nil, err
	}
	return pocket, nil
}

func (s *PocketService) DeletePocket(userID, pocketID uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", pocketID, userID).Delete(&models.Pocket{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("kantong tidak ditemukan")
	}
	return nil
}

