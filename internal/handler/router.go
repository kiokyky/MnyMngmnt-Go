package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kiokyky/MnyMngmnt-Go/internal/config"
	"github.com/kiokyky/MnyMngmnt-Go/internal/docs"
	"github.com/kiokyky/MnyMngmnt-Go/internal/middleware"
	"github.com/kiokyky/MnyMngmnt-Go/internal/service"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	docs.RegisterSwaggerRoutes(r)

	authService := service.NewAuthService(db, cfg.JWTSecret, cfg.JWTExpiryHours)
	pocketService := service.NewPocketService(db)
	categoryService := service.NewCategoryService(db)
	txService := service.NewTransactionService(db)
	portfolioService := service.NewPortfolioService(db)
	dashboardService := service.NewDashboardService(db, pocketService, portfolioService)

	authHandler := NewAuthHandler(authService)
	pocketHandler := NewPocketHandler(pocketService)
	categoryHandler := NewCategoryHandler(categoryService)
	txHandler := NewTransactionHandler(txService)
	portfolioHandler := NewPortfolioHandler(portfolioService)
	dashboardHandler := NewDashboardHandler(dashboardService)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/users/me", authHandler.Me)
			protected.GET("/dashboard", dashboardHandler.Summary)

			pockets := protected.Group("/pockets")
			{
				pockets.GET("", pocketHandler.List)
				pockets.POST("", pocketHandler.Create)
				pockets.GET("/:id", pocketHandler.Get)
				pockets.PUT("/:id", pocketHandler.Update)
				pockets.DELETE("/:id", pocketHandler.Delete)
			}

			categories := protected.Group("/categories")
			{
				categories.GET("", categoryHandler.List)
				categories.POST("", categoryHandler.Create)
				categories.PUT("/:id", categoryHandler.Update)
				categories.DELETE("/:id", categoryHandler.Delete)
			}

			transactions := protected.Group("/transactions")
			{
				transactions.GET("", txHandler.List)
				transactions.POST("", txHandler.Create)
				transactions.DELETE("/:id", txHandler.Delete)
			}

			portfolio := protected.Group("/portfolio")
			{
				portfolio.GET("/holdings", portfolioHandler.ListHoldings)
				portfolio.POST("/holdings", portfolioHandler.CreateHolding)
				portfolio.GET("/holdings/:id", portfolioHandler.GetHoldingDetail)
				portfolio.GET("/holdings/:id/history", portfolioHandler.GetHoldingDetail)
				portfolio.POST("/holdings/:id/buy", portfolioHandler.Buy)
				portfolio.POST("/holdings/:id/sell", portfolioHandler.Sell)
				portfolio.PATCH("/holdings/:id/rate", portfolioHandler.UpdateRate)
			}
		}
	}

	return r
}

