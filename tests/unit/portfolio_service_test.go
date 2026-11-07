package unit_test

import (
	"testing"

	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"github.com/kiokyky/MnyMngmnt-Go/internal/service"
	"github.com/shopspring/decimal"
)

func TestCalculateMarginView_MonthlyBluForexScenario(t *testing.T) {
	// Skenario 1:
	// Bulan 1: Beli $100 USD di kurs Rp 15.500 -> Modal Rp 1.550.000
	// Bulan 2: Beli $100 USD di kurs Rp 16.000 -> Modal Rp 1.600.000
	// Total Unit = $200 USD, Total Modal = Rp 3.150.000, Average Rate = Rp 15.750
	holding := models.AssetHolding{
		AssetType:         models.AssetTypeForex,
		Symbol:            "USD",
		Name:              "Blu Valas USD",
		TotalUnits:        decimal.NewFromInt(200),
		TotalCostIDR:      decimal.NewFromInt(3150000),
		AverageBuyRateIDR: decimal.NewFromInt(15750),
		LastKnownRateIDR:  decimal.NewFromInt(16000),
		RealizedPnLIDR:    decimal.Zero,
	}

	// Cek margin saat kurs Blu naik ke Rp 16.200 (Untung / Profit)
	currentBluRate := decimal.NewFromInt(16200)
	view := service.CalculateMarginView(holding, &currentBluRate)

	expectedMarketValue := decimal.NewFromInt(3240000) // 200 * 16.200
	expectedUnrealizedPnL := decimal.NewFromInt(90000) // 3.240.000 - 3.150.000
	expectedPct := decimal.RequireFromString("2.86")   // 90.000 / 3.150.000 * 100 = 2.86%

	if !view.MarketValueIDR.Equal(expectedMarketValue) {
		t.Fatalf("expected MarketValueIDR %s, got %s", expectedMarketValue, view.MarketValueIDR)
	}
	if !view.UnrealizedPnLIDR.Equal(expectedUnrealizedPnL) {
		t.Fatalf("expected UnrealizedPnLIDR %s, got %s", expectedUnrealizedPnL, view.UnrealizedPnLIDR)
	}
	if !view.UnrealizedPnLPercent.Equal(expectedPct) {
		t.Fatalf("expected UnrealizedPnLPercent %s, got %s", expectedPct, view.UnrealizedPnLPercent)
	}
}

func TestCalculateMarginView_UnrealizedLossScenario(t *testing.T) {
	// Skenario 2: Kurs turun di bawah rata-rata harga beli (Rugi / Floating Loss)
	holding := models.AssetHolding{
		AssetType:         models.AssetTypeForex,
		Symbol:            "SGD",
		Name:              "Blu Valas SGD",
		TotalUnits:        decimal.NewFromInt(500),
		TotalCostIDR:      decimal.NewFromInt(6000000), // Average beli: Rp 12.000 / SGD
		AverageBuyRateIDR: decimal.NewFromInt(12000),
		LastKnownRateIDR:  decimal.NewFromInt(11700), // Kurs turun ke Rp 11.700
		RealizedPnLIDR:    decimal.Zero,
	}

	view := service.CalculateMarginView(holding, nil)

	expectedMarketValue := decimal.NewFromInt(5850000)   // 500 * 11.700
	expectedUnrealizedPnL := decimal.NewFromInt(-150000) // -Rp 150.000
	expectedPct := decimal.RequireFromString("-2.5")     // -2.50%

	if !view.MarketValueIDR.Equal(expectedMarketValue) {
		t.Fatalf("expected MarketValueIDR %s, got %s", expectedMarketValue, view.MarketValueIDR)
	}
	if !view.UnrealizedPnLIDR.Equal(expectedUnrealizedPnL) {
		t.Fatalf("expected UnrealizedPnLIDR %s, got %s", expectedUnrealizedPnL, view.UnrealizedPnLIDR)
	}
	if !view.UnrealizedPnLPercent.Equal(expectedPct) {
		t.Fatalf("expected UnrealizedPnLPercent %s, got %s", expectedPct, view.UnrealizedPnLPercent)
	}
}

func TestCalculateMarginView_ZeroHolding(t *testing.T) {
	// Skenario 3: Holding baru atau sudah dijual habis (0 unit) tidak boleh division by zero
	holding := models.AssetHolding{
		AssetType:         models.AssetTypeGold,
		Symbol:            "XAU",
		Name:              "Tabungan Emas",
		TotalUnits:        decimal.Zero,
		TotalCostIDR:      decimal.Zero,
		AverageBuyRateIDR: decimal.Zero,
		LastKnownRateIDR:  decimal.NewFromInt(1350000),
		RealizedPnLIDR:    decimal.NewFromInt(250000),
	}

	view := service.CalculateMarginView(holding, nil)

	if !view.MarketValueIDR.IsZero() {
		t.Fatalf("expected 0 market value, got %s", view.MarketValueIDR)
	}
	if !view.UnrealizedPnLIDR.IsZero() {
		t.Fatalf("expected 0 unrealized PnL, got %s", view.UnrealizedPnLIDR)
	}
	if !view.UnrealizedPnLPercent.IsZero() {
		t.Fatalf("expected 0%% unrealized PnL percent, got %s", view.UnrealizedPnLPercent)
	}
	if !view.TotalReturnIDR.Equal(decimal.NewFromInt(250000)) {
		t.Fatalf("expected TotalReturnIDR 250000 from realized PnL, got %s", view.TotalReturnIDR)
	}
}

