# MnyMngmnt-Go — Money Management & Forex/Investment Tracker API

Backend REST API menggunakan **Go (Gin + GORM + PostgreSQL)** untuk pencatatan keuangan pribadi multi-user, manajemen banyak kantong (*pockets*), kategori kustom, serta pelacakan portofolio **Valas (Blu Valas, dll.) & Investasi** lengkap dengan perhitungan **Weighted Average Cost**, **pemotongan saldo kantong otomatis**, dan **Margin (Unrealized & Realized P&L)**.

---

## Cara Menjalankan

### 1. Siapkan Database PostgreSQL Lokal (Homebrew)
Pastikan service PostgreSQL Anda berjalan dan buat database `mny_mngmnt`:
```bash
brew services start postgresql@16
createdb mny_mngmnt
cp .env.example .env
```
*(Sesuaikan `DB_USER` dan `DB_PASSWORD` di `.env` dengan user PostgreSQL lokal Anda, misalnya `DB_USER=talangmas` dan `DB_PASSWORD=` jika tanpa password).*

### 2. Jalankan Server API
```bash
go run ./cmd/api
```
Server akan berjalan di `http://localhost:8080` dan otomatis melakukan migrasi tabel PostgreSQL (`AutoMigrate`).

---

## Contoh Alur Penggunaan (Studi Kasus Valas Blu)

### 1. Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Budi",
    "email": "budi@example.com",
    "password": "password123",
    "base_currency": "IDR"
  }'
```
*(Saat registrasi, sistem otomatis membuatkan kategori dasar pemasukan & pengeluaran yang bisa Anda tambah/ubah sendiri lewat `/api/v1/categories`).*

### 2. Buat Kantong IDR ("Blu Utama")
```bash
curl -X POST http://localhost:8080/api/v1/pockets \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Blu Utama",
    "type": "bank",
    "currency": "IDR",
    "initial_balance": 10000000
  }'
```

### 3. Buat Aset Tracking Valas ("Blu Valas USD")
Set `"create_pocket": true` agar sistem otomatis membuatkan kantong valas USD yang terhubung:
```bash
curl -X POST http://localhost:8080/api/v1/portfolio/holdings \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "create_pocket": true,
    "asset_type": "forex",
    "symbol": "USD",
    "name": "Blu Valas USD",
    "current_rate_idr": 15500
  }'
```

### 4. Beli Valas Bulanan (Otomatis Potong Saldo Kantong IDR)
- **Bulan 1**: Beli **$100 USD** di kurs **Rp 15.500** menggunakan saldo dari `source_pocket_id` ("Blu Utama"):
```bash
curl -X POST http://localhost:8080/api/v1/portfolio/holdings/<HOLDING_ID>/buy \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "source_pocket_id": "<BLU_UTAMA_POCKET_ID>",
    "units": 100,
    "rate_idr": 15500,
    "fee_idr": 0,
    "notes": "Tabungan Valas Bulan 1"
  }'
```
- **Bulan 2**: Beli lagi **$100 USD** di kurs **Rp 16.000**:
```bash
curl -X POST http://localhost:8080/api/v1/portfolio/holdings/<HOLDING_ID>/buy \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "source_pocket_id": "<BLU_UTAMA_POCKET_ID>",
    "units": 100,
    "rate_idr": 16000,
    "fee_idr": 0,
    "notes": "Tabungan Valas Bulan 2"
  }'
```
Hasilnya:
- Saldo kantong **Blu Utama** otomatis berkurang **Rp 3.150.000** (dan tercatat di riwayat transaksi pengeluaran).
- Saldo kantong **Blu Valas USD** otomatis bertambah menjadi **$200 USD**.
- `total_cost_idr` tercatat **Rp 3.150.000** dengan `average_buy_rate_idr` = **Rp 15.750 / USD**.

### 5. Cek Margin (Untung / Rugi) Berdasarkan Kurs Blu Saat Ini
Anda bisa memperbarui kurs pasar terbaru di Blu:
```bash
curl -X PATCH http://localhost:8080/api/v1/portfolio/holdings/<HOLDING_ID>/rate \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "current_rate_idr": 16200
  }'
```
Atau langsung cek detail & simulasi kurs lewat query parameter `?current_rate=16200`:
```bash
curl "http://localhost:8080/api/v1/portfolio/holdings/<HOLDING_ID>?current_rate=16200" \
  -H "Authorization: Bearer <TOKEN>"
```
Response akan menampilkan:
- `total_units`: `200` (USD)
- `total_cost_idr`: `3150000` (Modal Rupiah yang sudah masuk)
- `average_buy_rate_idr`: `15750` (Kurs rata-rata beli Anda)
- `market_value_idr`: `3240000` (Nilai sekarang jika dicairkan di kurs 16.200)
- `unrealized_pnl_idr`: `90000` (**Untung +Rp 90.000**)
- `unrealized_pnl_percent`: `2.86` (**Margin +2.86%**)

### 6. Cek Ringkasan Dashboard & Total Kekayaan Bersih (`GET /api/v1/dashboard`)
```bash
curl "http://localhost:8080/api/v1/dashboard?month=2026-11" \
  -H "Authorization: Bearer <TOKEN>"
```
Mengembalikan gabungan:
- `total_idr_pockets_balance`: Total saldo seluruh kantong Rupiah
- `portfolio_market_value_idr` & `portfolio_unrealized_pnl_idr`: Total nilai pasar & margin seluruh Valas/Investasi
- `total_net_worth_idr`: Total kekayaan bersih (Saldo Kantong IDR + Nilai Pasar Portofolio Valas/Investasi)
- `monthly_income_idr`, `monthly_expense_idr`, & `expense_by_category`: Ringkasan arus kas bulanan dan rincian pengeluaran per kategori


