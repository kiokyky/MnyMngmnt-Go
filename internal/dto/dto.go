package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"github.com/shopspring/decimal"
)

// --- Auth DTOs ---

type RegisterRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=120"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	BaseCurrency string `json:"base_currency"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// --- Pocket DTOs ---

type CreatePocketRequest struct {
	Name           string            `json:"name" binding:"required,max=120"`
	Type           models.PocketType `json:"type" binding:"required,oneof=cash bank ewallet forex investment"`
	Currency       string            `json:"currency"`
	InitialBalance decimal.Decimal   `json:"initial_balance"`
}

type UpdatePocketRequest struct {
	Name     string            `json:"name" binding:"required,max=120"`
	Type     models.PocketType `json:"type" binding:"required,oneof=cash bank ewallet forex investment"`
	Currency string            `json:"currency"`
}

// --- Category DTOs ---

type CreateCategoryRequest struct {
	Name  string              `json:"name" binding:"required,max=100"`
	Type  models.CategoryType `json:"type" binding:"required,oneof=income expense"`
	Icon  string              `json:"icon"`
	Color string              `json:"color"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name" binding:"required,max=100"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// --- Transaction DTOs ---

type CreateTransactionRequest struct {
	PocketID        uuid.UUID              `json:"pocket_id" binding:"required"`
	RelatedPocketID *uuid.UUID             `json:"related_pocket_id"`
	CategoryID      *uuid.UUID             `json:"category_id"`
	Type            models.TransactionType `json:"type" binding:"required,oneof=income expense transfer"`
	Amount          decimal.Decimal        `json:"amount" binding:"required"`
	Description     string                 `json:"description"`
	TransactionDate *time.Time             `json:"transaction_date"`
}

type TransactionFilter struct {
	PocketID   string `form:"pocket_id"`
	CategoryID string `form:"category_id"`
	Type       string `form:"type"`
	StartDate  string `form:"start_date"` // YYYY-MM-DD
	EndDate    string `form:"end_date"`   // YYYY-MM-DD
}

type TransactionSummaryResponse struct {
	TotalIncome   decimal.Decimal      `json:"total_income"`
	TotalExpense  decimal.Decimal      `json:"total_expense"`
	NetCashflow   decimal.Decimal      `json:"net_cashflow"`
	Transactions  []models.Transaction `json:"transactions"`
}

// --- Portfolio (Valas & Investment) DTOs ---

type CreateHoldingRequest struct {
	PocketID       *uuid.UUID       `json:"pocket_id"`
	CreatePocket   bool             `json:"create_pocket"`
	AssetType      models.AssetType `json:"asset_type" binding:"required,oneof=forex gold stock mutual_fund crypto other"`
	Symbol         string           `json:"symbol" binding:"required,max=30"`
	Name           string           `json:"name" binding:"required,max=120"`
	CurrentRateIDR decimal.Decimal  `json:"current_rate_idr"`
}

type BuyAssetRequest struct {
	SourcePocketID  *uuid.UUID      `json:"source_pocket_id"`
	CategoryID      *uuid.UUID      `json:"category_id"`
	Units           decimal.Decimal `json:"units" binding:"required"`
	RateIDR         decimal.Decimal `json:"rate_idr" binding:"required"`
	FeeIDR          decimal.Decimal `json:"fee_idr"`
	Notes           string          `json:"notes"`
	TransactionDate *time.Time      `json:"transaction_date"`
}

type SellAssetRequest struct {
	TargetPocketID  *uuid.UUID      `json:"target_pocket_id"`
	CategoryID      *uuid.UUID      `json:"category_id"`
	Units           decimal.Decimal `json:"units" binding:"required"`
	RateIDR         decimal.Decimal `json:"rate_idr" binding:"required"`
	FeeIDR          decimal.Decimal `json:"fee_idr"`
	Notes           string          `json:"notes"`
	TransactionDate *time.Time      `json:"transaction_date"`
}

type UpdateHoldingRateRequest struct {
	CurrentRateIDR decimal.Decimal `json:"current_rate_idr" binding:"required"`
}

type HoldingMarginView struct {
	Holding                models.AssetHolding `json:"holding"`
	CurrentRateIDR         decimal.Decimal     `json:"current_rate_idr"`
	MarketValueIDR         decimal.Decimal     `json:"market_value_idr"`
	UnrealizedPnLIDR       decimal.Decimal     `json:"unrealized_pnl_idr"`
	UnrealizedPnLPercent   decimal.Decimal     `json:"unrealized_pnl_percent"`
	TotalReturnIDR         decimal.Decimal     `json:"total_return_idr"` // Unrealized + Realized
}

type PortfolioOverviewResponse struct {
	TotalInvestedIDR       decimal.Decimal     `json:"total_invested_idr"`
	TotalMarketValueIDR    decimal.Decimal     `json:"total_market_value_idr"`
	TotalUnrealizedPnLIDR  decimal.Decimal     `json:"total_unrealized_pnl_idr"`
	TotalUnrealizedPercent decimal.Decimal     `json:"total_unrealized_percent"`
	TotalRealizedPnLIDR    decimal.Decimal     `json:"total_realized_pnl_idr"`
	Holdings               []HoldingMarginView `json:"holdings"`
}

// --- Dashboard & Net Worth DTOs ---

type CategoryBreakdownItem struct {
	CategoryID   uuid.UUID       `json:"category_id"`
	CategoryName string          `json:"category_name"`
	Type         string          `json:"type"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
}

type DashboardSummaryResponse struct {
	TotalIDRPocketsBalance decimal.Decimal             `json:"total_idr_pockets_balance"`
	PortfolioMarketValue   decimal.Decimal             `json:"portfolio_market_value_idr"`
	PortfolioInvestedCost  decimal.Decimal             `json:"portfolio_invested_cost_idr"`
	PortfolioUnrealizedPnL decimal.Decimal             `json:"portfolio_unrealized_pnl_idr"`
	PortfolioUnrealizedPct decimal.Decimal             `json:"portfolio_unrealized_percent"`
	TotalNetWorthIDR       decimal.Decimal             `json:"total_net_worth_idr"`
	MonthlyIncomeIDR       decimal.Decimal             `json:"monthly_income_idr"`
	MonthlyExpenseIDR      decimal.Decimal             `json:"monthly_expense_idr"`
	MonthlyNetCashflowIDR  decimal.Decimal             `json:"monthly_net_cashflow_idr"`
	ExpenseByCategory      []CategoryBreakdownItem     `json:"expense_by_category"`
	Pockets                []models.Pocket             `json:"pockets"`
	Portfolio              *PortfolioOverviewResponse  `json:"portfolio"`
}


