package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiokyky/MnyMngmnt-Go/internal/config"
	"github.com/kiokyky/MnyMngmnt-Go/internal/dto"
	"github.com/kiokyky/MnyMngmnt-Go/internal/handler"
	"github.com/kiokyky/MnyMngmnt-Go/internal/models"
	"github.com/shopspring/decimal"
)

func TestFullWorkflowIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.LoadConfig()

	db, err := config.ConnectDB(cfg)
	if err != nil {
		t.Skipf("Skipping DB integration test (PostgreSQL not reachable in current environment): %v", err)
	}

	router := handler.SetupRouter(db, cfg)

	// Gunakan email unik per test run & hapus bersih semua data user setelah test selesai
	testEmail := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	var testUserID uuid.UUID

	defer func() {
		if testUserID != uuid.Nil {
			db.Where("user_id = ?", testUserID).Delete(&models.AssetTransaction{})
			db.Where("user_id = ?", testUserID).Delete(&models.AssetHolding{})
			db.Where("user_id = ?", testUserID).Delete(&models.Transaction{})
			db.Where("user_id = ?", testUserID).Delete(&models.Pocket{})
			db.Where("user_id = ?", testUserID).Delete(&models.Category{})
			db.Where("id = ?", testUserID).Delete(&models.User{})
		}
	}()

	// 1. Cek Health & Swagger endpoint
	{
		w := performRequest(router, http.MethodGet, "/health", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("expected /health 200, got %d", w.Code)
		}

		wSwagger := performRequest(router, http.MethodGet, "/swagger/openapi.json", "", nil)
		if wSwagger.Code != http.StatusOK {
			t.Fatalf("expected /swagger/openapi.json 200, got %d", wSwagger.Code)
		}
	}

	// 2. Register User baru
	var token string
	{
		body := map[string]interface{}{
			"name":          "Integration Tester",
			"email":         testEmail,
			"password":      "password123",
			"base_currency": "IDR",
		}
		w := performRequest(router, http.MethodPost, "/api/v1/auth/register", "", body)
		if w.Code != http.StatusCreated {
			t.Fatalf("register failed (%d): %s", w.Code, w.Body.String())
		}

		var authRes dto.AuthResponse
		if err := json.Unmarshal(w.Body.Bytes(), &authRes); err != nil {
			t.Fatalf("failed to parse register response: %v", err)
		}
		token = authRes.Token
		testUserID = authRes.User.ID
	}

	// 3. Buat Kantong IDR "Blu Utama" dengan saldo awal Rp 10.000.000
	var bluUtamaID uuid.UUID
	{
		body := map[string]interface{}{
			"name":            "Blu Utama",
			"type":            "bank",
			"currency":        "IDR",
			"initial_balance": 10000000,
		}
		w := performRequest(router, http.MethodPost, "/api/v1/pockets", token, body)
		if w.Code != http.StatusCreated {
			t.Fatalf("create pocket failed (%d): %s", w.Code, w.Body.String())
		}

		var res struct {
			Data models.Pocket `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		bluUtamaID = res.Data.ID
	}

	// 4. Uji Transaksi Pengeluaran & Proteksi Saldo Tidak Cukup
	{
		// Pengeluaran melebihi saldo (Rp 50.000.000 > Rp 10.000.000) harus ditolak
		overdraftBody := map[string]interface{}{
			"pocket_id":   bluUtamaID,
			"type":        "expense",
			"amount":      50000000,
			"description": "Belanja melebihi saldo",
		}
		wFail := performRequest(router, http.MethodPost, "/api/v1/transactions", token, overdraftBody)
		if wFail.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 when balance insufficient, got %d", wFail.Code)
		}

		// Pengeluaran normal Rp 250.000 lalu hapus/batalkan transaksi
		expenseBody := map[string]interface{}{
			"pocket_id":   bluUtamaID,
			"type":        "expense",
			"amount":      250000,
			"description": "Makan malam",
		}
		wExp := performRequest(router, http.MethodPost, "/api/v1/transactions", token, expenseBody)
		if wExp.Code != http.StatusCreated {
			t.Fatalf("create expense failed (%d): %s", wExp.Code, wExp.Body.String())
		}
		var txRes struct {
			Data models.Transaction `json:"data"`
		}
		_ = json.Unmarshal(wExp.Body.Bytes(), &txRes)

		// Hapus transaksi -> saldo kembali ke Rp 10.000.000
		wDel := performRequest(router, http.MethodDelete, "/api/v1/transactions/"+txRes.Data.ID.String(), token, nil)
		if wDel.Code != http.StatusOK {
			t.Fatalf("delete transaction failed (%d): %s", wDel.Code, wDel.Body.String())
		}
	}

	// 5. Buat Aset Valas "Blu Valas USD" & Beli Valas 2 Bulan Berturut-turut (Otomatis Potong Saldo Blu Utama)
	var usdHoldingID uuid.UUID
	{
		createHoldingBody := map[string]interface{}{
			"create_pocket":    true,
			"asset_type":       "forex",
			"symbol":           "USD",
			"name":             "Blu Valas USD",
			"current_rate_idr": 15500,
		}
		wHold := performRequest(router, http.MethodPost, "/api/v1/portfolio/holdings", token, createHoldingBody)
		if wHold.Code != http.StatusCreated {
			t.Fatalf("create holding failed (%d): %s", wHold.Code, wHold.Body.String())
		}
		var holdRes struct {
			Data dto.HoldingMarginView `json:"data"`
		}
		_ = json.Unmarshal(wHold.Body.Bytes(), &holdRes)
		usdHoldingID = holdRes.Data.Holding.ID

		// Bulan 1: Beli $100 @ Rp 15.500 (Rp 1.550.000)
		buy1 := map[string]interface{}{
			"source_pocket_id": bluUtamaID,
			"units":            100,
			"rate_idr":         15500,
			"notes":            "Beli Valas Bulan 1",
		}
		wBuy1 := performRequest(router, http.MethodPost, fmt.Sprintf("/api/v1/portfolio/holdings/%s/buy", usdHoldingID), token, buy1)
		if wBuy1.Code != http.StatusCreated {
			t.Fatalf("buy 1 failed (%d): %s", wBuy1.Code, wBuy1.Body.String())
		}

		// Bulan 2: Beli $100 @ Rp 16.000 (Rp 1.600.000)
		buy2 := map[string]interface{}{
			"source_pocket_id": bluUtamaID,
			"units":            100,
			"rate_idr":         16000,
			"notes":            "Beli Valas Bulan 2",
		}
		wBuy2 := performRequest(router, http.MethodPost, fmt.Sprintf("/api/v1/portfolio/holdings/%s/buy", usdHoldingID), token, buy2)
		if wBuy2.Code != http.StatusCreated {
			t.Fatalf("buy 2 failed (%d): %s", wBuy2.Code, wBuy2.Body.String())
		}
	}

	// 6. Update Kurs Blu ke Rp 16.200 & Verifikasi Margin (+Rp 90.000 / +2.86%)
	{
		rateBody := map[string]interface{}{
			"current_rate_idr": 16200,
		}
		wRate := performRequest(router, http.MethodPatch, fmt.Sprintf("/api/v1/portfolio/holdings/%s/rate", usdHoldingID), token, rateBody)
		if wRate.Code != http.StatusOK {
			t.Fatalf("update rate failed (%d): %s", wRate.Code, wRate.Body.String())
		}

		var rateRes struct {
			Summary dto.HoldingMarginView `json:"summary"`
		}
		_ = json.Unmarshal(wRate.Body.Bytes(), &rateRes)

		if !rateRes.Summary.Holding.TotalCostIDR.Equal(decimal.NewFromInt(3150000)) {
			t.Fatalf("expected total cost 3150000, got %s", rateRes.Summary.Holding.TotalCostIDR)
		}
		if !rateRes.Summary.Holding.AverageBuyRateIDR.Equal(decimal.NewFromInt(15750)) {
			t.Fatalf("expected avg buy rate 15750, got %s", rateRes.Summary.Holding.AverageBuyRateIDR)
		}
		if !rateRes.Summary.UnrealizedPnLIDR.Equal(decimal.NewFromInt(90000)) {
			t.Fatalf("expected unrealized PnL 90000, got %s", rateRes.Summary.UnrealizedPnLIDR)
		}
	}

	// 7. Jual $50 USD @ Rp 16.200 cairkan ke Blu Utama & Cek Dashboard Net Worth
	{
		sellBody := map[string]interface{}{
			"target_pocket_id": bluUtamaID,
			"units":            50,
			"rate_idr":         16200,
			"notes":            "Take profit 50 USD",
		}
		wSell := performRequest(router, http.MethodPost, fmt.Sprintf("/api/v1/portfolio/holdings/%s/sell", usdHoldingID), token, sellBody)
		if wSell.Code != http.StatusCreated {
			t.Fatalf("sell failed (%d): %s", wSell.Code, wSell.Body.String())
		}

		// Cek Dashboard:
		// Saldo Blu Utama = 10.000.000 - 3.150.000 + 810.000 = Rp 7.660.000
		// Nilai Pasar Sisa $150 USD @ 16.200 = Rp 2.430.000
		// Total Net Worth = Rp 10.090.000
		wDash := performRequest(router, http.MethodGet, "/api/v1/dashboard", token, nil)
		if wDash.Code != http.StatusOK {
			t.Fatalf("dashboard failed (%d): %s", wDash.Code, wDash.Body.String())
		}
		var dash dto.DashboardSummaryResponse
		_ = json.Unmarshal(wDash.Body.Bytes(), &dash)

		if !dash.TotalIDRPocketsBalance.Equal(decimal.NewFromInt(7660000)) {
			t.Fatalf("expected IDR pockets balance 7660000, got %s", dash.TotalIDRPocketsBalance)
		}
		if !dash.PortfolioMarketValue.Equal(decimal.NewFromInt(2430000)) {
			t.Fatalf("expected portfolio market value 2430000, got %s", dash.PortfolioMarketValue)
		}
		if !dash.TotalNetWorthIDR.Equal(decimal.NewFromInt(10090000)) {
			t.Fatalf("expected total net worth 10090000, got %s", dash.TotalNetWorthIDR)
		}
	}
}

func performRequest(r http.Handler, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

