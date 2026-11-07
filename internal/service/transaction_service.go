package service

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) ListTransactions(userID uuid.UUID, filter dto.TransactionFilter) (*dto.TransactionSummaryResponse, error) {
	query := s.db.Where("user_id = ?", userID).
		Preload("Pocket").
		Preload("RelatedPocket").
		Preload("Category")

	if filter.PocketID != "" {
		query = query.Where("pocket_id = ? OR related_pocket_id = ?", filter.PocketID, filter.PocketID)
	}
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.StartDate != "" {
		if t, err := time.Parse("2006-01-02", filter.StartDate); err == nil {
			query = query.Where("transaction_date >= ?", t)
		}
	}
	if filter.EndDate != "" {
		if t, err := time.Parse("2006-01-02", filter.EndDate); err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("transaction_date <= ?", endOfDay)
		}
	}

	var transactions []models.Transaction
	if err := query.Order("transaction_date DESC, created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}

	totalIncome := decimal.Zero
	totalExpense := decimal.Zero
	for _, tx := range transactions {
		switch tx.Type {
		case models.TransactionTypeIncome:
			totalIncome = totalIncome.Add(tx.Amount)
		case models.TransactionTypeExpense:
			totalExpense = totalExpense.Add(tx.Amount)
		}
	}

	return &dto.TransactionSummaryResponse{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
		NetCashflow:  totalIncome.Sub(totalExpense),
		Transactions: transactions,
	}, nil
}

func (s *TransactionService) CreateTransaction(userID uuid.UUID, req dto.CreateTransactionRequest) (*models.Transaction, error) {
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("nominal transaksi harus lebih dari 0")
	}

	txDate := time.Now()
	if req.TransactionDate != nil && !req.TransactionDate.IsZero() {
		txDate = *req.TransactionDate
	}

	var createdTx models.Transaction

	err := s.db.Transaction(func(tx *gorm.DB) error {
		var pocket models.Pocket
		if err := tx.Where("id = ? AND user_id = ?", req.PocketID, userID).First(&pocket).Error; err != nil {
			return errors.New("kantong sumber tidak ditemukan")
		}

		if req.CategoryID != nil {
			var cat models.Category
			if err := tx.Where("id = ? AND user_id = ?", *req.CategoryID, userID).First(&cat).Error; err != nil {
				return errors.New("kategori tidak ditemukan")
			}
		}

		switch req.Type {
		case models.TransactionTypeIncome:
			pocket.Balance = pocket.Balance.Add(req.Amount)
			if err := tx.Save(&pocket).Error; err != nil {
				return err
			}

		case models.TransactionTypeExpense:
			if pocket.Balance.LessThan(req.Amount) {
				return errors.New("saldo kantong tidak mencukupi")
			}
			pocket.Balance = pocket.Balance.Sub(req.Amount)
			if err := tx.Save(&pocket).Error; err != nil {
				return err
			}

		case models.TransactionTypeTransfer:
			if req.RelatedPocketID == nil || *req.RelatedPocketID == req.PocketID {
				return errors.New("kantong tujuan transfer (related_pocket_id) wajib diisi dan berbeda dari kantong asal")
			}
			if pocket.Balance.LessThan(req.Amount) {
				return errors.New("saldo kantong asal tidak mencukupi untuk transfer")
			}

			var targetPocket models.Pocket
			if err := tx.Where("id = ? AND user_id = ?", *req.RelatedPocketID, userID).First(&targetPocket).Error; err != nil {
				return errors.New("kantong tujuan transfer tidak ditemukan")
			}

			pocket.Balance = pocket.Balance.Sub(req.Amount)
			targetPocket.Balance = targetPocket.Balance.Add(req.Amount)

			if err := tx.Save(&pocket).Error; err != nil {
				return err
			}
			if err := tx.Save(&targetPocket).Error; err != nil {
				return err
			}

		default:
			return errors.New("tipe transaksi tidak valid")
		}

		createdTx = models.Transaction{
			UserID:          userID,
			PocketID:        req.PocketID,
			RelatedPocketID: req.RelatedPocketID,
			CategoryID:      req.CategoryID,
			Type:            req.Type,
			Amount:          req.Amount,
			Description:     strings.TrimSpace(req.Description),
			TransactionDate: txDate,
		}

		return tx.Create(&createdTx).Error
	})
	if err != nil {
		return nil, err
	}

	s.db.Preload("Pocket").Preload("RelatedPocket").Preload("Category").First(&createdTx, "id = ?", createdTx.ID)
	return &createdTx, nil
}

func (s *TransactionService) DeleteTransaction(userID, txID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var record models.Transaction
		if err := tx.Where("id = ? AND user_id = ?", txID, userID).First(&record).Error; err != nil {
			return errors.New("transaksi tidak ditemukan")
		}

		var pocket models.Pocket
		if err := tx.Where("id = ? AND user_id = ?", record.PocketID, userID).First(&pocket).Error; err != nil {
			return errors.New("kantong terkait tidak ditemukan")
		}

		switch record.Type {
		case models.TransactionTypeIncome:
			pocket.Balance = pocket.Balance.Sub(record.Amount)
			if err := tx.Save(&pocket).Error; err != nil {
				return err
			}

		case models.TransactionTypeExpense:
			pocket.Balance = pocket.Balance.Add(record.Amount)
			if err := tx.Save(&pocket).Error; err != nil {
				return err
			}

		case models.TransactionTypeTransfer:
			pocket.Balance = pocket.Balance.Add(record.Amount)
			if err := tx.Save(&pocket).Error; err != nil {
				return err
			}
			if record.RelatedPocketID != nil {
				var targetPocket models.Pocket
				if err := tx.Where("id = ? AND user_id = ?", *record.RelatedPocketID, userID).First(&targetPocket).Error; err == nil {
					targetPocket.Balance = targetPocket.Balance.Sub(record.Amount)
					if err := tx.Save(&targetPocket).Error; err != nil {
						return err
					}
				}
			}
		}

		return tx.Delete(&record).Error
	})
}

