package service

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DashboardService struct {
	db               *gorm.DB
	pocketService    *PocketService
	portfolioService *PortfolioService
}

func NewDashboardService(db *gorm.DB, pocketService *PocketService, portfolioService *PortfolioService) *DashboardService {
	return &DashboardService{
		db:               db,
		pocketService:    pocketService,
		portfolioService: portfolioService,
	}
}

func (s *DashboardService) GetDashboardSummary(userID uuid.UUID, monthStr string) (*dto.DashboardSummaryResponse, error) {
	pockets, err := s.pocketService.ListPockets(userID)
	if err != nil {
		return nil, err
	}

	portfolio, err := s.portfolioService.ListHoldings(userID, "")
	if err != nil {
		return nil, err
	}

	// Hitung total saldo kantong IDR (kantong non-valas / mata uang IDR)
	totalIDRPockets := decimal.Zero
	for _, p := range pockets {
		if strings.EqualFold(p.Currency, "IDR") && p.Type != models.PocketTypeForex && p.Type != models.PocketTypeInvestment {
			totalIDRPockets = totalIDRPockets.Add(p.Balance)
		}
	}

	// Tentukan rentang waktu bulan berjalan (atau sesuai parameter ?month=YYYY-MM)
	now := time.Now()
	if monthStr != "" {
		if parsed, err := time.Parse("2006-01", monthStr); err == nil {
			now = parsed
		}
	}
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	var monthlyTxs []models.Transaction
	if err := s.db.Preload("Category").
		Where("user_id = ? AND transaction_date >= ? AND transaction_date <= ?", userID, startOfMonth, endOfMonth).
		Find(&monthlyTxs).Error; err != nil {
		return nil, err
	}

	monthlyIncome := decimal.Zero
	monthlyExpense := decimal.Zero
	categoryTotals := make(map[uuid.UUID]*dto.CategoryBreakdownItem)

	for _, tx := range monthlyTxs {
		switch tx.Type {
		case models.TransactionTypeIncome:
			monthlyIncome = monthlyIncome.Add(tx.Amount)
		case models.TransactionTypeExpense:
			monthlyExpense = monthlyExpense.Add(tx.Amount)
			if tx.CategoryID != nil && tx.Category != nil {
				item, exists := categoryTotals[*tx.CategoryID]
				if !exists {
					item = &dto.CategoryBreakdownItem{
						CategoryID:   *tx.CategoryID,
						CategoryName: tx.Category.Name,
						Type:         string(tx.Category.Type),
						TotalAmount:  decimal.Zero,
					}
					categoryTotals[*tx.CategoryID] = item
				}
				item.TotalAmount = item.TotalAmount.Add(tx.Amount)
			}
		}
	}

	expenseBreakdown := make([]dto.CategoryBreakdownItem, 0, len(categoryTotals))
	for _, item := range categoryTotals {
		expenseBreakdown = append(expenseBreakdown, *item)
	}

	totalNetWorthIDR := totalIDRPockets.Add(portfolio.TotalMarketValueIDR).Round(4)

	return &dto.DashboardSummaryResponse{
		TotalIDRPocketsBalance: totalIDRPockets.Round(4),
		PortfolioMarketValue:   portfolio.TotalMarketValueIDR,
		PortfolioInvestedCost:  portfolio.TotalInvestedIDR,
		PortfolioUnrealizedPnL: portfolio.TotalUnrealizedPnLIDR,
		PortfolioUnrealizedPct: portfolio.TotalUnrealizedPercent,
		TotalNetWorthIDR:       totalNetWorthIDR,
		MonthlyIncomeIDR:       monthlyIncome.Round(4),
		MonthlyExpenseIDR:      monthlyExpense.Round(4),
		MonthlyNetCashflowIDR:  monthlyIncome.Sub(monthlyExpense).Round(4),
		ExpenseByCategory:      expenseBreakdown,
		Pockets:                pockets,
		Portfolio:              portfolio,
	}, nil
}

