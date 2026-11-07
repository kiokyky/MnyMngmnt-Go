package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PocketType string

const (
	PocketTypeCash       PocketType = "cash"
	PocketTypeBank       PocketType = "bank"
	PocketTypeEWallet    PocketType = "ewallet"
	PocketTypeForex      PocketType = "forex"
	PocketTypeInvestment PocketType = "investment"
)

type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

type AssetType string

const (
	AssetTypeForex      AssetType = "forex"
	AssetTypeGold       AssetType = "gold"
	AssetTypeStock      AssetType = "stock"
	AssetTypeMutualFund AssetType = "mutual_fund"
	AssetTypeCrypto     AssetType = "crypto"
	AssetTypeOther      AssetType = "other"
)

type AssetTxType string

const (
	AssetTxBuy  AssetTxType = "BUY"
	AssetTxSell AssetTxType = "SELL"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"size:120;not null" json:"name"`
	Email        string    `gorm:"size:180;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	BaseCurrency string    `gorm:"size:10;not null;default:'IDR'" json:"base_currency"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	if u.BaseCurrency == "" {
		u.BaseCurrency = "IDR"
	}
	return nil
}

type Pocket struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
	Name      string          `gorm:"size:120;not null" json:"name"`
	Type      PocketType      `gorm:"size:30;not null;default:'bank'" json:"type"`
	Currency  string          `gorm:"size:10;not null;default:'IDR'" json:"currency"`
	Balance   decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (p *Pocket) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Currency == "" {
		p.Currency = "IDR"
	}
	if p.Type == "" {
		p.Type = PocketTypeBank
	}
	return nil
}

type Category struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID    `gorm:"type:uuid;index;not null" json:"user_id"`
	Name      string       `gorm:"size:100;not null" json:"name"`
	Type      CategoryType `gorm:"size:20;index;not null" json:"type"`
	Icon      string       `gorm:"size:60" json:"icon"`
	Color     string       `gorm:"size:20" json:"color"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type Transaction struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
	PocketID        uuid.UUID       `gorm:"type:uuid;index;not null" json:"pocket_id"`
	Pocket          *Pocket         `gorm:"foreignKey:PocketID" json:"pocket,omitempty"`
	RelatedPocketID *uuid.UUID      `gorm:"type:uuid;index" json:"related_pocket_id,omitempty"`
	RelatedPocket   *Pocket         `gorm:"foreignKey:RelatedPocketID" json:"related_pocket,omitempty"`
	CategoryID      *uuid.UUID      `gorm:"type:uuid;index" json:"category_id,omitempty"`
	Category        *Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Type            TransactionType `gorm:"size:20;index;not null" json:"type"`
	Amount          decimal.Decimal `gorm:"type:numeric(20,4);not null" json:"amount"`
	Description     string          `gorm:"size:255" json:"description"`
	TransactionDate time.Time       `gorm:"index;not null" json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.TransactionDate.IsZero() {
		t.TransactionDate = time.Now()
	}
	return nil
}

// AssetHolding stores aggregated cost-basis & unit tracking for a Forex currency or Investment asset.
type AssetHolding struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
	PocketID          *uuid.UUID      `gorm:"type:uuid;index" json:"pocket_id,omitempty"`
	Pocket            *Pocket         `gorm:"foreignKey:PocketID" json:"pocket,omitempty"`
	AssetType         AssetType       `gorm:"size:30;index;not null;default:'forex'" json:"asset_type"`
	Symbol            string          `gorm:"size:30;index;not null" json:"symbol"`
	Name              string          `gorm:"size:120;not null" json:"name"`
	TotalUnits        decimal.Decimal `gorm:"type:numeric(20,6);not null;default:0" json:"total_units"`
	TotalCostIDR      decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"total_cost_idr"`
	AverageBuyRateIDR decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"average_buy_rate_idr"`
	LastKnownRateIDR  decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"last_known_rate_idr"`
	RealizedPnLIDR    decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"realized_pnl_idr"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (a *AssetHolding) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.AssetType == "" {
		a.AssetType = AssetTypeForex
	}
	return nil
}

// AssetTransaction records every monthly/periodic BUY or SELL lot for a holding.
type AssetTransaction struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	HoldingID       uuid.UUID       `gorm:"type:uuid;index;not null" json:"holding_id"`
	UserID          uuid.UUID       `gorm:"type:uuid;index;not null" json:"user_id"`
	SourcePocketID  *uuid.UUID      `gorm:"type:uuid;index" json:"source_pocket_id,omitempty"`
	SourcePocket    *Pocket         `gorm:"foreignKey:SourcePocketID" json:"source_pocket,omitempty"`
	Type            AssetTxType     `gorm:"size:10;not null" json:"type"`
	Units           decimal.Decimal `gorm:"type:numeric(20,6);not null" json:"units"`
	RateIDR         decimal.Decimal `gorm:"type:numeric(20,4);not null" json:"rate_idr"`
	FeeIDR          decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"fee_idr"`
	TotalIDR        decimal.Decimal `gorm:"type:numeric(20,4);not null" json:"total_idr"`
	RealizedPnLIDR  decimal.Decimal `gorm:"type:numeric(20,4);not null;default:0" json:"realized_pnl_idr"`
	Notes           string          `gorm:"size:255" json:"notes"`
	TransactionDate time.Time       `gorm:"index;not null" json:"transaction_date"`
	CreatedAt       time.Time       `json:"created_at"`
}

func (at *AssetTransaction) BeforeCreate(tx *gorm.DB) error {
	if at.ID == uuid.Nil {
		at.ID = uuid.New()
	}
	if at.TransactionDate.IsZero() {
		at.TransactionDate = time.Now()
	}
	return nil
}

