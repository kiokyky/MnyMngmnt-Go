package service

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/middleware"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db             *gorm.DB
	jwtSecret      string
	jwtExpiryHours int
}

func NewAuthService(db *gorm.DB, jwtSecret string, jwtExpiryHours int) *AuthService {
	return &AuthService{
		db:             db,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var existing models.User
	if err := s.db.Where("email = ?", email).First(&existing).Error; err == nil {
		return nil, errors.New("email sudah terdaftar")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	baseCurrency := strings.ToUpper(strings.TrimSpace(req.BaseCurrency))
	if baseCurrency == "" {
		baseCurrency = "IDR"
	}

	user := models.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		PasswordHash: string(hashedBytes),
		BaseCurrency: baseCurrency,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		defaultCategories := []models.Category{
			{UserID: user.ID, Name: "Gaji", Type: models.CategoryTypeIncome, Icon: "wallet", Color: "#10B981"},
			{UserID: user.ID, Name: "Bonus & THR", Type: models.CategoryTypeIncome, Icon: "gift", Color: "#34D399"},
			{UserID: user.ID, Name: "Hasil Investasi / Valas", Type: models.CategoryTypeIncome, Icon: "trending-up", Color: "#059669"},
			{UserID: user.ID, Name: "Makan & Minum", Type: models.CategoryTypeExpense, Icon: "utensils", Color: "#EF4444"},
			{UserID: user.ID, Name: "Transportasi", Type: models.CategoryTypeExpense, Icon: "car", Color: "#F97316"},
			{UserID: user.ID, Name: "Tagihan & Utilitas", Type: models.CategoryTypeExpense, Icon: "zap", Color: "#EAB308"},
			{UserID: user.ID, Name: "Belanja Kebutuhan", Type: models.CategoryTypeExpense, Icon: "shopping-bag", Color: "#EC4899"},
			{UserID: user.ID, Name: "Beli Valas & Investasi", Type: models.CategoryTypeExpense, Icon: "dollar-sign", Color: "#6366F1"},
		}

		return tx.Create(&defaultCategories).Error
	})
	if err != nil {
		return nil, err
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiryHours)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, s.jwtSecret, s.jwtExpiryHours)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

