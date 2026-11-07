package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PortfolioService struct {
	db *gorm.DB
}

func NewPortfolioService(db *gorm.DB) *PortfolioService {
	return &PortfolioService{db: db}
}

func CalculateMarginView(holding models.AssetHolding, overrideRateIDR *decimal.Decimal) dto.HoldingMarginView {
	currentRate := holding.LastKnownRateIDR
	if overrideRateIDR != nil && overrideRateIDR.GreaterThan(decimal.Zero) {
		currentRate = *overrideRateIDR
	}
	if currentRate.LessThanOrEqual(decimal.Zero) {
		currentRate = holding.AverageBuyRateIDR
	}

	marketValueIDR := holding.TotalUnits.Mul(currentRate).Round(4)
	unrealizedPnLIDR := marketValueIDR.Sub(holding.TotalCostIDR).Round(4)

	unrealizedPct := decimal.Zero
	if holding.TotalCostIDR.GreaterThan(decimal.Zero) {
		unrealizedPct = unrealizedPnLIDR.Div(holding.TotalCostIDR).Mul(decimal.NewFromInt(100)).Round(2)
	}

	totalReturnIDR := unrealizedPnLIDR.Add(holding.RealizedPnLIDR).Round(4)

	return dto.HoldingMarginView{
		Holding:              holding,
		CurrentRateIDR:       currentRate,
		MarketValueIDR:       marketValueIDR,
		UnrealizedPnLIDR:     unrealizedPnLIDR,
		UnrealizedPnLPercent: unrealizedPct,
		TotalReturnIDR:       totalReturnIDR,
	}
}

func (s *PortfolioService) CreateHolding(userID uuid.UUID, req dto.CreateHoldingRequest) (*dto.HoldingMarginView, error) {
	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	name := strings.TrimSpace(req.Name)

	var createdHolding models.AssetHolding

	err := s.db.Transaction(func(tx *gorm.DB) error {
		pocketID := req.PocketID

		if pocketID != nil {
			var pocket models.Pocket
			if err := tx.Where("id = ? AND user_id = ?", *pocketID, userID).First(&pocket).Error; err != nil {
				return errors.New("kantong valas/investasi yang dipilih tidak ditemukan")
			}
		} else if req.CreatePocket {
			pocketType := models.PocketTypeInvestment
			if req.AssetType == models.AssetTypeForex {
				pocketType = models.PocketTypeForex
			}
			newPocket := models.Pocket{
				UserID:   userID,
				Name:     name,
				Type:     pocketType,
				Currency: symbol,
				Balance:  decimal.Zero,
			}
			if err := tx.Create(&newPocket).Error; err != nil {
				return err
			}
			pocketID = &newPocket.ID
		}

		createdHolding = models.AssetHolding{
			UserID:            userID,
			PocketID:          pocketID,
			AssetType:         req.AssetType,
			Symbol:            symbol,
			Name:              name,
			TotalUnits:        decimal.Zero,
			TotalCostIDR:      decimal.Zero,
			AverageBuyRateIDR: decimal.Zero,
			LastKnownRateIDR:  req.CurrentRateIDR,
			RealizedPnLIDR:    decimal.Zero,
		}

		return tx.Create(&createdHolding).Error
	})
	if err != nil {
		return nil, err
	}

	s.db.Preload("Pocket").First(&createdHolding, "id = ?", createdHolding.ID)
	view := CalculateMarginView(createdHolding, nil)
	return &view, nil
}

func (s *PortfolioService) ListHoldings(userID uuid.UUID, assetType string) (*dto.PortfolioOverviewResponse, error) {
	query := s.db.Where("user_id = ?", userID).Preload("Pocket")
	if assetType != "" {
		query = query.Where("asset_type = ?", assetType)
	}

	var holdings []models.AssetHolding
	if err := query.Order("created_at ASC").Find(&holdings).Error; err != nil {
		return nil, err
	}

	views := make([]dto.HoldingMarginView, 0, len(holdings))
	totalInvested := decimal.Zero
	totalMarketValue := decimal.Zero
	totalUnrealized := decimal.Zero
	totalRealized := decimal.Zero

	for _, h := range holdings {
		v := CalculateMarginView(h, nil)
		views = append(views, v)

		totalInvested = totalInvested.Add(h.TotalCostIDR)
		totalMarketValue = totalMarketValue.Add(v.MarketValueIDR)
		totalUnrealized = totalUnrealized.Add(v.UnrealizedPnLIDR)
		totalRealized = totalRealized.Add(h.RealizedPnLIDR)
	}

	totalUnrealizedPct := decimal.Zero
	if totalInvested.GreaterThan(decimal.Zero) {
		totalUnrealizedPct = totalUnrealized.Div(totalInvested).Mul(decimal.NewFromInt(100)).Round(2)
	}

	return &dto.PortfolioOverviewResponse{
		TotalInvestedIDR:       totalInvested.Round(4),
		TotalMarketValueIDR:    totalMarketValue.Round(4),
		TotalUnrealizedPnLIDR:  totalUnrealized.Round(4),
		TotalUnrealizedPercent: totalUnrealizedPct,
		TotalRealizedPnLIDR:    totalRealized.Round(4),
		Holdings:               views,
	}, nil
}

func (s *PortfolioService) GetHoldingDetail(userID, holdingID uuid.UUID, overrideRate *decimal.Decimal) (*dto.HoldingMarginView, []models.AssetTransaction, error) {
	var holding models.AssetHolding
	if err := s.db.Preload("Pocket").Where("id = ? AND user_id = ?", holdingID, userID).First(&holding).Error; err != nil {
		return nil, nil, errors.New("aset/valas tidak ditemukan")
	}

	var history []models.AssetTransaction
	if err := s.db.Preload("SourcePocket").
		Where("holding_id = ? AND user_id = ?", holdingID, userID).
		Order("transaction_date DESC, created_at DESC").
		Find(&history).Error; err != nil {
		return nil, nil, err
	}

	view := CalculateMarginView(holding, overrideRate)
	return &view, history, nil
}

// BuyAsset records a monthly/periodic purchase of Forex or Investment asset,
// updates weighted average buy rate & total IDR cost basis,
// and automatically deducts IDR balance from SourcePocketID if provided.
func (s *PortfolioService) BuyAsset(userID, holdingID uuid.UUID, req dto.BuyAssetRequest) (*dto.HoldingMarginView, *models.AssetTransaction, error) {
	if req.Units.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("jumlah unit/valas harus lebih dari 0")
	}
	if req.RateIDR.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("kurs/harga beli (rate_idr) harus lebih dari 0")
	}
	if req.FeeIDR.LessThan(decimal.Zero) {
		return nil, nil, errors.New("biaya admin (fee_idr) tidak boleh negatif")
	}

	txDate := time.Now()
	if req.TransactionDate != nil && !req.TransactionDate.IsZero() {
		txDate = *req.TransactionDate
	}

	totalCostIDR := req.Units.Mul(req.RateIDR).Add(req.FeeIDR).Round(4)

	var updatedHolding models.AssetHolding
	var assetTx models.AssetTransaction

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", holdingID, userID).First(&updatedHolding).Error; err != nil {
			return errors.New("aset/valas tidak ditemukan")
		}

		// 1. Jika source_pocket_id diisi, potong saldo kantong IDR sumber & catat transaksi pengeluaran IDR
		if req.SourcePocketID != nil {
			var sourcePocket models.Pocket
			if err := tx.Where("id = ? AND user_id = ?", *req.SourcePocketID, userID).First(&sourcePocket).Error; err != nil {
				return errors.New("kantong sumber dana IDR tidak ditemukan")
			}
			if sourcePocket.Balance.LessThan(totalCostIDR) {
				return fmt.Errorf("saldo kantong %s tidak mencukupi (dibutuhkan: %s, tersedia: %s)",
					sourcePocket.Name, totalCostIDR.StringFixed(2), sourcePocket.Balance.StringFixed(2))
			}

			sourcePocket.Balance = sourcePocket.Balance.Sub(totalCostIDR)
			if err := tx.Save(&sourcePocket).Error; err != nil {
				return err
			}

			desc := strings.TrimSpace(req.Notes)
			if desc == "" {
				desc = fmt.Sprintf("Beli %s %s @ Rp %s", req.Units.String(), updatedHolding.Symbol, req.RateIDR.StringFixed(2))
			}

			cashflowTx := models.Transaction{
				UserID:          userID,
				PocketID:        sourcePocket.ID,
				RelatedPocketID: updatedHolding.PocketID,
				CategoryID:      req.CategoryID,
				Type:            models.TransactionTypeExpense,
				Amount:          totalCostIDR,
				Description:     desc,
				TransactionDate: txDate,
			}
			if err := tx.Create(&cashflowTx).Error; err != nil {
				return err
			}
		}

		// 2. Jika holding terhubung ke kantong Valas/Investasi, tambahkan saldo unitnya
		if updatedHolding.PocketID != nil {
			var assetPocket models.Pocket
			if err := tx.Where("id = ? AND user_id = ?", *updatedHolding.PocketID, userID).First(&assetPocket).Error; err == nil {
				assetPocket.Balance = assetPocket.Balance.Add(req.Units)
				if err := tx.Save(&assetPocket).Error; err != nil {
					return err
				}
			}
		}

		// 3. Hitung ulang TotalUnits, TotalCostIDR, dan Weighted Average Buy Rate
		newTotalUnits := updatedHolding.TotalUnits.Add(req.Units)
		newTotalCostIDR := updatedHolding.TotalCostIDR.Add(totalCostIDR)
		newAvgRate := newTotalCostIDR.Div(newTotalUnits).Round(4)

		updatedHolding.TotalUnits = newTotalUnits
		updatedHolding.TotalCostIDR = newTotalCostIDR
		updatedHolding.AverageBuyRateIDR = newAvgRate
		updatedHolding.LastKnownRateIDR = req.RateIDR

		if err := tx.Save(&updatedHolding).Error; err != nil {
			return err
		}

		// 4. Simpan riwayat transaksi lot pembelian
		assetTx = models.AssetTransaction{
			HoldingID:       updatedHolding.ID,
			UserID:          userID,
			SourcePocketID:  req.SourcePocketID,
			Type:            models.AssetTxBuy,
			Units:           req.Units,
			RateIDR:         req.RateIDR,
			FeeIDR:          req.FeeIDR,
			TotalIDR:        totalCostIDR,
			RealizedPnLIDR:  decimal.Zero,
			Notes:           strings.TrimSpace(req.Notes),
			TransactionDate: txDate,
		}
		return tx.Create(&assetTx).Error
	})
	if err != nil {
		return nil, nil, err
	}

	s.db.Preload("Pocket").First(&updatedHolding, "id = ?", updatedHolding.ID)
	view := CalculateMarginView(updatedHolding, nil)
	return &view, &assetTx, nil
}

// SellAsset records a sale/conversion of Forex or Investment back to IDR,
// calculates Realized P&L against the weighted average buy rate,
// reduces TotalCostIDR proportionally, and credits IDR to TargetPocketID if provided.
func (s *PortfolioService) SellAsset(userID, holdingID uuid.UUID, req dto.SellAssetRequest) (*dto.HoldingMarginView, *models.AssetTransaction, error) {
	if req.Units.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("jumlah unit/valas yang dijual harus lebih dari 0")
	}
	if req.RateIDR.LessThanOrEqual(decimal.Zero) {
		return nil, nil, errors.New("kurs/harga jual (rate_idr) harus lebih dari 0")
	}
	if req.FeeIDR.LessThan(decimal.Zero) {
		return nil, nil, errors.New("biaya admin (fee_idr) tidak boleh negatif")
	}

	txDate := time.Now()
	if req.TransactionDate != nil && !req.TransactionDate.IsZero() {
		txDate = *req.TransactionDate
	}

	netProceedsIDR := req.Units.Mul(req.RateIDR).Sub(req.FeeIDR).Round(4)
	if netProceedsIDR.LessThan(decimal.Zero) {
		return nil, nil, errors.New("total hasil penjualan setelah biaya admin tidak boleh negatif")
	}

	var updatedHolding models.AssetHolding
	var assetTx models.AssetTransaction

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND user_id = ?", holdingID, userID).First(&updatedHolding).Error; err != nil {
			return errors.New("aset/valas tidak ditemukan")
		}

		if updatedHolding.TotalUnits.LessThan(req.Units) {
			return fmt.Errorf("saldo unit %s tidak mencukupi (tersedia: %s, ingin dijual: %s)",
				updatedHolding.Symbol, updatedHolding.TotalUnits.String(), req.Units.String())
		}

		// Hitung modal (cost basis) dari unit yang dijual menggunakan AverageBuyRateIDR
		var costBasisSoldIDR decimal.Decimal
		remainingUnits := updatedHolding.TotalUnits.Sub(req.Units)
		if remainingUnits.IsZero() {
			costBasisSoldIDR = updatedHolding.TotalCostIDR
		} else {
			costBasisSoldIDR = req.Units.Mul(updatedHolding.AverageBuyRateIDR).Round(4)
		}

		realizedPnL := netProceedsIDR.Sub(costBasisSoldIDR).Round(4)

		// 1. Jika target_pocket_id diisi, tambahkan hasil penjualan IDR ke kantong tersebut
		if req.TargetPocketID != nil {
			var targetPocket models.Pocket
			if err := tx.Where("id = ? AND user_id = ?", *req.TargetPocketID, userID).First(&targetPocket).Error; err != nil {
				return errors.New("kantong tujuan pencairan IDR tidak ditemukan")
			}

			targetPocket.Balance = targetPocket.Balance.Add(netProceedsIDR)
			if err := tx.Save(&targetPocket).Error; err != nil {
				return err
			}

			desc := strings.TrimSpace(req.Notes)
			if desc == "" {
				desc = fmt.Sprintf("Jual %s %s @ Rp %s (P&L: Rp %s)",
					req.Units.String(), updatedHolding.Symbol, req.RateIDR.StringFixed(2), realizedPnL.StringFixed(2))
			}

			cashflowTx := models.Transaction{
				UserID:          userID,
				PocketID:        targetPocket.ID,
				RelatedPocketID: updatedHolding.PocketID,
				CategoryID:      req.CategoryID,
				Type:            models.TransactionTypeIncome,
				Amount:          netProceedsIDR,
				Description:     desc,
				TransactionDate: txDate,
			}
			if err := tx.Create(&cashflowTx).Error; err != nil {
				return err
			}
		}

		// 2. Jika holding terhubung ke kantong Valas/Investasi, kurangi saldo unitnya
		if updatedHolding.PocketID != nil {
			var assetPocket models.Pocket
			if err := tx.Where("id = ? AND user_id = ?", *updatedHolding.PocketID, userID).First(&assetPocket).Error; err == nil {
				assetPocket.Balance = assetPocket.Balance.Sub(req.Units)
				if err := tx.Save(&assetPocket).Error; err != nil {
					return err
				}
			}
		}

		// 3. Update holding
		updatedHolding.TotalUnits = remainingUnits
		updatedHolding.TotalCostIDR = updatedHolding.TotalCostIDR.Sub(costBasisSoldIDR)
		if remainingUnits.IsZero() {
			updatedHolding.TotalCostIDR = decimal.Zero
			updatedHolding.AverageBuyRateIDR = decimal.Zero
		}
		updatedHolding.RealizedPnLIDR = updatedHolding.RealizedPnLIDR.Add(realizedPnL)
		updatedHolding.LastKnownRateIDR = req.RateIDR

		if err := tx.Save(&updatedHolding).Error; err != nil {
			return err
		}

		// 4. Catat riwayat transaksi penjualan
		assetTx = models.AssetTransaction{
			HoldingID:       updatedHolding.ID,
			UserID:          userID,
			SourcePocketID:  req.TargetPocketID,
			Type:            models.AssetTxSell,
			Units:           req.Units,
			RateIDR:         req.RateIDR,
			FeeIDR:          req.FeeIDR,
			TotalIDR:        netProceedsIDR,
			RealizedPnLIDR:  realizedPnL,
			Notes:           strings.TrimSpace(req.Notes),
			TransactionDate: txDate,
		}
		return tx.Create(&assetTx).Error
	})
	if err != nil {
		return nil, nil, err
	}

	s.db.Preload("Pocket").First(&updatedHolding, "id = ?", updatedHolding.ID)
	view := CalculateMarginView(updatedHolding, nil)
	return &view, &assetTx, nil
}

func (s *PortfolioService) UpdateMarketRate(userID, holdingID uuid.UUID, req dto.UpdateHoldingRateRequest) (*dto.HoldingMarginView, error) {
	if req.CurrentRateIDR.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("kurs/harga saat ini harus lebih dari 0")
	}

	var holding models.AssetHolding
	if err := s.db.Preload("Pocket").Where("id = ? AND user_id = ?", holdingID, userID).First(&holding).Error; err != nil {
		return nil, errors.New("aset/valas tidak ditemukan")
	}

	holding.LastKnownRateIDR = req.CurrentRateIDR
	if err := s.db.Save(&holding).Error; err != nil {
		return nil, err
	}

	view := CalculateMarginView(holding, nil)
	return &view, nil
}

