package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/handlers"
	"pos-backend/middleware"
	"pos-backend/seed"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "pos.db"
	}
	db.Init(dbPath)
	seed.Run()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	api := router.Group("/api")
	{
		api.POST("/auth/login", handlers.Login)

		authed := api.Group("/")
		authed.Use(middleware.AuthRequired())
		{
			authed.GET("auth/me", handlers.Me)

			authed.GET("categories", handlers.ListCategories)
			authed.GET("categories/:id/option-templates", handlers.ListCategoryOptionTemplates)
			authed.GET("menu-items", handlers.ListMenuItems)
			authed.GET("tables", handlers.ListTables)

			authed.POST("orders", handlers.CreateOrder)
			authed.GET("orders", handlers.ListOrders)
			authed.GET("orders/:id", handlers.GetOrder)
			authed.POST("orders/:id/items", handlers.AddOrderItem)
			authed.PUT("orders/:id/items/:itemId", handlers.UpdateOrderItem)
			authed.DELETE("orders/:id/items/:itemId", handlers.DeleteOrderItem)
			authed.PUT("orders/:id/status", handlers.UpdateOrderStatus)
			authed.PUT("orders/:id/discount", handlers.UpdateOrderDiscount)
			authed.PUT("orders/:id/guests", handlers.UpdateOrderGuestCount)
			authed.POST("orders/:id/pay", handlers.PayOrder)

			authed.GET("reports/daily", handlers.DailyReport)

			admin := authed.Group("/")
			admin.Use(middleware.AdminOnly())
			{
				admin.POST("categories", handlers.CreateCategory)
				admin.PUT("categories/:id", handlers.UpdateCategory)
				admin.DELETE("categories/:id", handlers.DeleteCategory)

				admin.POST("menu-items", handlers.CreateMenuItem)
				admin.PUT("menu-items/:id", handlers.UpdateMenuItem)
				admin.DELETE("menu-items/:id", handlers.DeleteMenuItem)

				admin.POST("menu-items/:id/option-groups", handlers.CreateOptionGroup)
				admin.PUT("option-groups/:id", handlers.UpdateOptionGroup)
				admin.DELETE("option-groups/:id", handlers.DeleteOptionGroup)
				admin.POST("option-groups/:id/choices", handlers.AddOptionChoice)
				admin.DELETE("choices/:id", handlers.DeleteOptionChoice)

				admin.POST("categories/:id/option-templates", handlers.CreateCategoryOptionTemplate)
				admin.DELETE("option-templates/:id", handlers.DeleteCategoryOptionTemplate)
				admin.POST("menu-items/:id/option-groups/from-template/:templateId", handlers.ApplyOptionTemplateToMenuItem)

				admin.POST("tables", handlers.CreateTable)
				admin.PUT("tables/:id", handlers.UpdateTable)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("POS backend listening on :%s", port)
	router.Run(":" + port)
}
