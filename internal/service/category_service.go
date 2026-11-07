package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) ListCategories(userID uuid.UUID, categoryType string) ([]models.Category, error) {
	query := s.db.Where("user_id = ?", userID)
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	var categories []models.Category
	if err := query.Order("type ASC, name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *CategoryService) CreateCategory(userID uuid.UUID, req dto.CreateCategoryRequest) (*models.Category, error) {
	cat := models.Category{
		UserID: userID,
		Name:   strings.TrimSpace(req.Name),
		Type:   req.Type,
		Icon:   strings.TrimSpace(req.Icon),
		Color:  strings.TrimSpace(req.Color),
	}

	if err := s.db.Create(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *CategoryService) UpdateCategory(userID, categoryID uuid.UUID, req dto.UpdateCategoryRequest) (*models.Category, error) {
	var cat models.Category
	if err := s.db.Where("id = ? AND user_id = ?", categoryID, userID).First(&cat).Error; err != nil {
		return nil, errors.New("kategori tidak ditemukan")
	}

	cat.Name = strings.TrimSpace(req.Name)
	cat.Icon = strings.TrimSpace(req.Icon)
	cat.Color = strings.TrimSpace(req.Color)

	if err := s.db.Save(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *CategoryService) DeleteCategory(userID, categoryID uuid.UUID) error {
	res := s.db.Where("id = ? AND user_id = ?", categoryID, userID).Delete(&models.Category{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("kategori tidak ditemukan")
	}
	return nil
}

