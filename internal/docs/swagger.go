package docs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>MnyMngmnt-Go — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css" />
  <style>
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/swagger/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        persistAuthorization: true,
      });
    };
  </script>
</body>
</html>`

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "MnyMngmnt-Go — Personal Finance & Forex/Investment Tracker API",
    "description": "Backend REST API untuk manajemen keuangan multi-user, banyak kantong (pockets), kategori pemasukan/pengeluaran kustom, serta pelacakan Valas (Blu Valas) & Investasi dengan perhitungan Weighted Average Cost dan Margin (Unrealized & Realized P&L).\n\n### Cara Pakai Cepat di Swagger:\n1. Jalankan **POST /api/v1/auth/register** atau **POST /api/v1/auth/login** lalu salin nilai ` + "`token`" + `.\n2. Klik tombol **Authorize** (gembok hijau di kanan atas), tempelkan token tersebut (tanpa kata Bearer), lalu klik **Authorize**.",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local Development Server"
    }
  ],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT",
        "description": "Masukkan token JWT dari endpoint /api/v1/auth/login atau /api/v1/auth/register"
      }
    }
  },
  "tags": [
    { "name": "Auth", "description": "Registrasi, Login, dan Profil User" },
    { "name": "Dashboard", "description": "Ringkasan Total Kekayaan Bersih (IDR + Valas/Investasi) & Arus Kas Bulanan" },
    { "name": "Pockets", "description": "Manajemen Kantong / Dompet Multi-Mata-Uang (Blu Utama, BCA, Cash, Blu Valas)" },
    { "name": "Categories", "description": "Kategori Pemasukan & Pengeluaran yang bisa ditambah sendiri oleh User" },
    { "name": "Transactions", "description": "Pencatatan Pemasukan, Pengeluaran, dan Transfer antar Kantong" },
    { "name": "Portfolio (Valas & Investasi)", "description": "Tracking Pembelian/Penjualan Valas & Investasi, Weighted Average Rate, Pemotongan Saldo Kantong Otomatis, dan Perhitungan Margin" }
  ],
  "paths": {
    "/api/v1/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Registrasi user baru (otomatis membuat kategori dasar)",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "email", "password"],
                "properties": {
                  "name": { "type": "string", "example": "Budi" },
                  "email": { "type": "string", "example": "budi@example.com" },
                  "password": { "type": "string", "example": "password123" },
                  "base_currency": { "type": "string", "example": "IDR" }
                }
              }
            }
          }
        },
        "responses": {
          "201": { "description": "User berhasil dibuat beserta token JWT" }
        }
      }
    },
    "/api/v1/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Login user & dapatkan token JWT",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password"],
                "properties": {
                  "email": { "type": "string", "example": "budi@example.com" },
                  "password": { "type": "string", "example": "password123" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Login berhasil" }
        }
      }
    },
    "/api/v1/users/me": {
      "get": {
        "tags": ["Auth"],
        "summary": "Lihat profil user yang sedang login",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": { "description": "Data user" }
        }
      }
    },
    "/api/v1/dashboard": {
      "get": {
        "tags": ["Dashboard"],
        "summary": "Ringkasan kekayaan bersih (Net Worth), arus kas bulanan, dan portofolio valas/investasi",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          {
            "name": "month",
            "in": "query",
            "description": "Filter bulan arus kas dalam format YYYY-MM (opsional, default bulan berjalan)",
            "schema": { "type": "string", "example": "2026-11" }
          }
        ],
        "responses": {
          "200": { "description": "Ringkasan dashboard lengkap" }
        }
      }
    },
    "/api/v1/pockets": {
      "get": {
        "tags": ["Pockets"],
        "summary": "List semua kantong milik user",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": { "description": "Daftar kantong" }
        }
      },
      "post": {
        "tags": ["Pockets"],
        "summary": "Buat kantong baru (contoh: Blu Utama, BCA, Dompet Tunai)",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "type"],
                "properties": {
                  "name": { "type": "string", "example": "Blu Utama" },
                  "type": { "type": "string", "enum": ["cash", "bank", "ewallet", "forex", "investment"], "example": "bank" },
                  "currency": { "type": "string", "example": "IDR" },
                  "initial_balance": { "type": "number", "example": 10000000 }
                }
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Kantong berhasil dibuat" }
        }
      }
    },
    "/api/v1/pockets/{id}": {
      "get": {
        "tags": ["Pockets"],
        "summary": "Detail kantong berdasarkan ID",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": { "200": { "description": "Detail kantong" } }
      },
      "put": {
        "tags": ["Pockets"],
        "summary": "Update nama/tipe kantong",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "type"],
                "properties": {
                  "name": { "type": "string", "example": "Blu Utama BCA Digital" },
                  "type": { "type": "string", "enum": ["cash", "bank", "ewallet", "forex", "investment"], "example": "bank" },
                  "currency": { "type": "string", "example": "IDR" }
                }
              }
            }
          }
        },
        "responses": { "200": { "description": "Kantong berhasil diupdate" } }
      },
      "delete": {
        "tags": ["Pockets"],
        "summary": "Hapus kantong",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": { "200": { "description": "Kantong berhasil dihapus" } }
      }
    },
    "/api/v1/categories": {
      "get": {
        "tags": ["Categories"],
        "summary": "List semua kategori pemasukan & pengeluaran user",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          {
            "name": "type",
            "in": "query",
            "description": "Filter berdasarkan tipe kategori",
            "schema": { "type": "string", "enum": ["income", "expense"] }
          }
        ],
        "responses": { "200": { "description": "Daftar kategori" } }
      },
      "post": {
        "tags": ["Categories"],
        "summary": "Tambah kategori kustom baru",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name", "type"],
                "properties": {
                  "name": { "type": "string", "example": "Langganan Streaming" },
                  "type": { "type": "string", "enum": ["income", "expense"], "example": "expense" },
                  "icon": { "type": "string", "example": "tv" },
                  "color": { "type": "string", "example": "#8B5CF6" }
                }
              }
            }
          }
        },
        "responses": { "201": { "description": "Kategori berhasil ditambahkan" } }
      }
    },
    "/api/v1/categories/{id}": {
      "put": {
        "tags": ["Categories"],
        "summary": "Update kategori milik user",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["name"],
                "properties": {
                  "name": { "type": "string", "example": "Hiburan & Streaming" },
                  "icon": { "type": "string", "example": "film" },
                  "color": { "type": "string", "example": "#8B5CF6" }
                }
              }
            }
          }
        },
        "responses": { "200": { "description": "Kategori berhasil diupdate" } }
      },
      "delete": {
        "tags": ["Categories"],
        "summary": "Hapus kategori",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": { "200": { "description": "Kategori berhasil dihapus" } }
      }
    },
    "/api/v1/transactions": {
      "get": {
        "tags": ["Transactions"],
        "summary": "List transaksi beserta ringkasan total pemasukan & pengeluaran",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "pocket_id", "in": "query", "schema": { "type": "string", "format": "uuid" } },
          { "name": "category_id", "in": "query", "schema": { "type": "string", "format": "uuid" } },
          { "name": "type", "in": "query", "schema": { "type": "string", "enum": ["income", "expense", "transfer"] } },
          { "name": "start_date", "in": "query", "description": "YYYY-MM-DD", "schema": { "type": "string" } },
          { "name": "end_date", "in": "query", "description": "YYYY-MM-DD", "schema": { "type": "string" } }
        ],
        "responses": { "200": { "description": "Daftar transaksi dan rekap cashflow" } }
      },
      "post": {
        "tags": ["Transactions"],
        "summary": "Catat pemasukan, pengeluaran, atau transfer antar kantong (otomatis potong/tambah saldo)",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["pocket_id", "type", "amount"],
                "properties": {
                  "pocket_id": { "type": "string", "format": "uuid", "description": "ID kantong sumber" },
                  "related_pocket_id": { "type": "string", "format": "uuid", "nullable": true, "description": "Wajib diisi jika type = transfer" },
                  "category_id": { "type": "string", "format": "uuid", "nullable": true },
                  "type": { "type": "string", "enum": ["income", "expense", "transfer"], "example": "expense" },
                  "amount": { "type": "number", "example": 75000 },
                  "description": { "type": "string", "example": "Makan siang" },
                  "transaction_date": { "type": "string", "format": "date-time", "nullable": true }
                }
              }
            }
          }
        },
        "responses": { "201": { "description": "Transaksi berhasil dicatat dan saldo kantong diperbarui" } }
      }
    },
    "/api/v1/transactions/{id}": {
      "delete": {
        "tags": ["Transactions"],
        "summary": "Hapus/batalkan transaksi & kembalikan saldo kantong seperti semula",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": { "200": { "description": "Transaksi dihapus dan saldo dikembalikan" } }
      }
    },
    "/api/v1/portfolio/holdings": {
      "get": {
        "tags": ["Portfolio (Valas & Investasi)"],
        "summary": "Lihat seluruh aset Valas & Investasi beserta total modal IDR, rata-rata kurs beli, dan margin (untung/rugi)",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          {
            "name": "asset_type",
            "in": "query",
            "schema": { "type": "string", "enum": ["forex", "gold", "stock", "mutual_fund", "crypto", "other"] }
          }
        ],
        "responses": { "200": { "description": "Ringkasan seluruh portofolio dan margin" } }
      },
      "post": {
        "tags": ["Portfolio (Valas & Investasi)"],
        "summary": "Buat aset tracking Valas atau Investasi baru (misal: Blu Valas USD)",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["asset_type", "symbol", "name"],
                "properties": {
                  "create_pocket": { "type": "boolean", "example": true, "description": "Jika true, otomatis membuatkan kantong khusus mata uang/aset ini" },
                  "pocket_id": { "type": "string", "format": "uuid", "nullable": true, "description": "Opsional jika ingin menghubungkan ke kantong yang sudah ada" },
                  "asset_type": { "type": "string", "enum": ["forex", "gold", "stock", "mutual_fund", "crypto", "other"], "example": "forex" },
                  "symbol": { "type": "string", "example": "USD" },
                  "name": { "type": "string", "example": "Blu Valas USD" },
                  "current_rate_idr": { "type": "number", "example": 15500 }
                }
              }
            }
          }
        },
        "responses": { "201": { "description": "Aset holding berhasil dibuat" } }
      }
    },
    "/api/v1/portfolio/holdings/{id}": {
      "get": {
        "tags": ["Portfolio (Valas & Investasi)"],
        "summary": "Detail aset Valas/Investasi, simulasi margin pada kurs tertentu, dan riwayat pembelian/penjualan bulanan",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } },
          { "name": "current_rate", "in": "query", "description": "Simulasi kurs/harga saat ini (IDR) untuk menghitung margin secara langsung", "schema": { "type": "number", "example": 16200 } }
        ],
        "responses": { "200": { "description": "Detail holding beserta margin dan histori transaksi" } }
      }
    },
    "/api/v1/portfolio/holdings/{id}/buy": {
      "post": {
        "tags": ["Portfolio (Valas & Investasi)"],
        "summary": "Catat pembelian Valas/Investasi bulanan (otomatis potong saldo kantong IDR sumber & hitung ulang Average Cost)",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["units", "rate_idr"],
                "properties": {
                  "source_pocket_id": { "type": "string", "format": "uuid", "nullable": true, "description": "ID kantong IDR (misal Blu Utama) yang saldonya akan dipotong otomatis" },
                  "category_id": { "type": "string", "format": "uuid", "nullable": true },
                  "units": { "type": "number", "example": 100, "description": "Jumlah unit valas/aset yang dibeli (misal 100 USD)" },
                  "rate_idr": { "type": "number", "example": 15850, "description": "Kurs beli per 1 unit dalam Rupiah" },
                  "fee_idr": { "type": "number", "example": 0 },
                  "notes": { "type": "string", "example": "Beli USD rutin bulan November" },
                  "transaction_date": { "type": "string", "format": "date-time", "nullable": true }
                }
              }
            }
          }
        },
        "responses": { "201": { "description": "Pembelian tercatat, saldo kantong IDR terpotong, dan kurs rata-rata diperbarui" } }
      }
    },
    "/api/v1/portfolio/holdings/{id}/sell": {
      "post": {
        "tags": ["Portfolio (Valas & Investasi)"],
        "summary": "Catat penjualan/konversi Valas ke IDR (otomatis hitung Realized Profit/Loss & tambah saldo kantong IDR tujuan)",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["units", "rate_idr"],
                "properties": {
                  "target_pocket_id": { "type": "string", "format": "uuid", "nullable": true, "description": "ID kantong IDR tujuan pencairan dana" },
                  "category_id": { "type": "string", "format": "uuid", "nullable": true },
                  "units": { "type": "number", "example": 50, "description": "Jumlah unit valas/aset yang dijual" },
                  "rate_idr": { "type": "number", "example": 16250, "description": "Kurs jual per 1 unit dalam Rupiah" },
                  "fee_idr": { "type": "number", "example": 0 },
                  "notes": { "type": "string", "example": "Take profit sebagian USD" },
                  "transaction_date": { "type": "string", "format": "date-time", "nullable": true }
                }
              }
            }
          }
        },
        "responses": { "201": { "description": "Penjualan tercatat, realized profit dihitung, dan saldo kantong IDR bertambah" } }
      }
    },
    "/api/v1/portfolio/holdings/{id}/rate": {
      "patch": {
        "tags": ["Portfolio (Valas & Investasi)"],
        "summary": "Update kurs/harga pasar terbaru secara manual (misal sesuai kurs di aplikasi Blu hari ini) untuk melihat margin terbaru",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["current_rate_idr"],
                "properties": {
                  "current_rate_idr": { "type": "number", "example": 16200 }
                }
              }
            }
          }
        },
        "responses": { "200": { "description": "Kurs pasar diperbarui dan margin dihitung ulang" } }
      }
    }
  }
}`

func RegisterSwaggerRoutes(r *gin.Engine) {
	r.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})
	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIHTML))
	})
	r.GET("/swagger/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(openAPISpec))
	})
}

