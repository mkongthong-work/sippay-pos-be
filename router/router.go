// Package router ประกอบ route ทั้งหมดของ SipPay เป็น *gin.Engine เดียว — แยกออกมาจาก main.go เดิม
// เพื่อให้เรียกใช้ได้ทั้งจาก main.go (รันเป็นเซิร์ฟเวอร์ค้างปกติ ด้วย router.Run) และจาก api/index.go
// (Vercel serverless function ที่แค่ ServeHTTP ทีละ request ไม่เรียก Run) โดยไม่ต้องเขียน route ซ้ำสองที่
package router

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"pos-backend/handlers"
	"pos-backend/middleware"
)

// New สร้าง gin.Engine พร้อม route ทั้งหมด — ไม่ยุ่งกับการต่อฐานข้อมูล/seed (ดู bootstrap.Init แยกต่างหาก)
func New() *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	// เสิร์ฟไฟล์อัปโหลดจากดิสก์ในเครื่องตรงๆ — ใช้เฉพาะตอนไม่ได้ตั้งค่า Supabase Storage เท่านั้น (ถ้าใช้
	// Supabase Storage ไฟล์จะมี URL เต็มของตัวเองอยู่แล้ว ไม่ผ่าน route นี้ ดู utils/storage.go)
	if os.Getenv("SUPABASE_STORAGE_BUCKET") == "" {
		r.Static("/uploads/slips", "./uploads/slips")
		r.Static("/uploads/menu-items", "./uploads/menu-items")
	}

	api := r.Group("/api")
	{
		api.POST("/auth/login", handlers.Login)
		// เข้าระบบด้วย PIN 6 หลัก (สำหรับเครื่อง POS ที่พนักงานหลายคนใช้ร่วมกัน) — สาธารณะเหมือน /auth/login
		// เพราะยังไม่มี token ตอนเรียก 2 เส้นนี้
		api.GET("/auth/pin-users", handlers.ListPinLoginUsers)
		api.POST("/auth/pin-login", handlers.PinLogin)

		authed := api.Group("/")
		authed.Use(middleware.AuthRequired())
		{
			authed.GET("auth/me", handlers.Me)

			authed.GET("categories", handlers.ListCategories)
			authed.GET("categories/:id/option-templates", handlers.ListCategoryOptionTemplates)
			authed.GET("option-templates", handlers.ListAllCategoryOptionTemplates)
			authed.GET("option-templates/archived", handlers.ListArchivedCategoryOptionTemplates)
			authed.GET("menu-items", handlers.ListMenuItems)
			authed.GET("tables", handlers.ListTables)
			authed.GET("zones", handlers.ListZones)
			authed.GET("shop-settings", handlers.GetShopSettings)

			authed.POST("orders", handlers.CreateOrder)
			authed.GET("orders", handlers.ListOrders)
			authed.GET("orders/:id", handlers.GetOrder)
			authed.POST("orders/:id/items", handlers.AddOrderItem)
			authed.PUT("orders/:id/items/:itemId", handlers.UpdateOrderItem)
			authed.DELETE("orders/:id/items/:itemId", handlers.DeleteOrderItem)
			authed.PUT("orders/:id/status", handlers.UpdateOrderStatus)
			authed.PUT("orders/:id/discount", handlers.UpdateOrderDiscount)
			authed.PUT("orders/:id/guests", handlers.UpdateOrderGuestCount)
			authed.PUT("orders/:id/table", handlers.ChangeOrderTable)
			authed.POST("orders/:id/pay", handlers.PayOrder)
			authed.PUT("orders/:id/payment", handlers.UpdatePaymentSlip)
			authed.GET("orders/:id/invoice", handlers.GetOrderInvoicePDF)

			authed.GET("reservations", handlers.ListReservations)
			authed.POST("reservations", handlers.CreateReservation)
			authed.PUT("reservations/:id/cancel", handlers.CancelReservation)
			authed.PUT("reservations/:id/no-show", handlers.MarkNoShow)

			authed.GET("reports/daily", handlers.DailyReport)
			authed.GET("reports/range", handlers.SalesRange)

			authed.GET("members", handlers.ListMembers)
			authed.GET("members/by-phone/:phone", handlers.GetMemberByPhone)
			authed.GET("members/:id", handlers.GetMember)
			authed.POST("members", handlers.CreateMember)
			authed.PUT("members/:id", handlers.UpdateMember)
			authed.GET("members/:id/history", handlers.GetMemberHistory)
			authed.GET("loyalty-settings", handlers.GetLoyaltySettings)

			admin := authed.Group("/")
			admin.Use(middleware.AdminOnly())
			{
				admin.POST("categories", handlers.CreateCategory)
				admin.PUT("categories/:id", handlers.UpdateCategory)
				admin.DELETE("categories/:id", handlers.DeleteCategory)
				admin.PUT("categories/:id/archive", handlers.ArchiveCategory)
				admin.PUT("categories/:id/restore", handlers.RestoreCategory)

				admin.POST("menu-items", handlers.CreateMenuItem)
				admin.PUT("menu-items/:id", handlers.UpdateMenuItem)
				admin.DELETE("menu-items/:id", handlers.DeleteMenuItem)
				admin.PUT("menu-items/:id/archive", handlers.ArchiveMenuItem)
				admin.PUT("menu-items/:id/restore", handlers.RestoreMenuItem)
				admin.PUT("menu-items/:id/image", handlers.UploadMenuItemImage)
				admin.DELETE("menu-items/:id/image", handlers.DeleteMenuItemImage)

				admin.POST("menu-items/:id/option-groups", handlers.CreateOptionGroup)
				admin.PUT("option-groups/:id", handlers.UpdateOptionGroup)
				admin.DELETE("option-groups/:id", handlers.DeleteOptionGroup)
				admin.POST("option-groups/:id/choices", handlers.AddOptionChoice)
				admin.PUT("choices/:id", handlers.UpdateOptionChoice)
				admin.DELETE("choices/:id", handlers.DeleteOptionChoice)

				admin.POST("option-templates", handlers.CreateCategoryOptionTemplate)
				admin.PUT("option-templates/:id", handlers.UpdateCategoryOptionTemplate)
				admin.DELETE("option-templates/:id", handlers.DeleteCategoryOptionTemplate)
				admin.PUT("option-templates/:id/archive", handlers.ArchiveCategoryOptionTemplate)
				admin.PUT("option-templates/:id/restore", handlers.RestoreCategoryOptionTemplate)
				admin.POST("option-templates/:id/choices", handlers.AddCategoryOptionTemplateChoice)
				admin.PUT("template-choices/:id", handlers.UpdateCategoryOptionTemplateChoice)
				admin.DELETE("template-choices/:id", handlers.DeleteCategoryOptionTemplateChoice)
				admin.POST("menu-items/:id/option-groups/from-template/:templateId", handlers.ApplyOptionTemplateToMenuItem)

				admin.POST("tables", handlers.CreateTable)
				admin.PUT("tables/:id", handlers.UpdateTable)

				admin.POST("zones", handlers.CreateZone)
				admin.PUT("zones/:id", handlers.UpdateZone)

				admin.GET("users", handlers.ListUsers)
				admin.POST("users", handlers.CreateUser)
				admin.PUT("users/:id", handlers.UpdateUser)

				admin.PUT("shop-settings", handlers.UpdateShopSettings)

				admin.POST("members/:id/adjust-points", handlers.AdjustPoints)
				admin.PUT("loyalty-settings", handlers.UpdateLoyaltySettings)
			}
		}
	}

	return r
}
