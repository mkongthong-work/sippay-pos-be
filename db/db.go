package db

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pos-backend/models"
)

var DB *gorm.DB

// Init เชื่อมต่อฐานข้อมูลและสร้างตารางอัตโนมัติตาม struct ใน models
//   - ถ้าตั้ง env DATABASE_URL ไว้ (เช่น connection string ของ Supabase Postgres) → ใช้ Postgres
//     แนะนำให้ใช้ connection string จาก "Transaction pooler" ของ Supabase (พอร์ต 6543) ไม่ใช่พอร์ตตรง
//     5432 เพราะ backend รันเป็น serverless function บน Vercel แต่ละ request อาจเป็นคนละ process กัน
//     ถ้าต่อพอร์ตตรงจะเปิด connection ค้างจนเกินโควตาของ Postgres ได้ง่ายมาก
//   - ถ้าไม่ได้ตั้ง (รันเองบนเครื่อง/VM ปกติ) → ใช้ไฟล์ SQLite ตาม path ที่รับมาเหมือนเดิม
func Init(path string) {
	var database *gorm.DB
	var err error

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		// PreferSimpleProtocol: true — ปิดการใช้ prepared statement ของ pgx driver จำเป็นมากตอนต่อผ่าน
		// Supabase Transaction pooler (pgbouncer โหมด transaction) เพราะ pooler แชร์ connection ข้าม
		// request กัน ถ้าปล่อยให้ driver แคช prepared statement ไว้ที่ session จะชนกันจน error
		// "prepared statement ... already exists" (SQLSTATE 42P05) แบบที่เจอ — ปิดแล้วทุกคำสั่งจะส่งเป็น
		// simple query แทน ช้าลงนิดหน่อยแต่ทำงานถูกต้องกับ pooler แบบนี้
		database, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		}), &gorm.Config{})
	} else {
		database, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	}
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = database.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.CategoryOptionTemplate{},
		&models.CategoryOptionTemplateChoice{},
		&models.MenuItem{},
		&models.MenuOptionGroup{},
		&models.MenuOptionChoice{},
		&models.DiningTable{},
		&models.Zone{},
		&models.Reservation{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderItemOption{},
		&models.Payment{},
		&models.InvoiceCounter{},
		&models.ShopSettings{},
		&models.Member{},
		&models.MemberPointHistory{},
		&models.LoyaltySettings{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	DB = database

	backfillZonesFromTables()
	ensureShopSettings()
	ensureLoyaltySettings()
}

// ensureShopSettings สร้างแถวข้อมูลร้านค้า (id=1) ให้ครั้งแรกถ้ายังไม่มี ใส่ค่า placeholder ไว้ก่อน
// ให้ admin มาแก้ที่หน้า "ตั้งค่าร้านค้า" ทีหลัง (รันซ้ำได้ ไม่ error ถ้ามีแถวอยู่แล้วจะข้ามไปเฉยๆ)
func ensureShopSettings() {
	var count int64
	DB.Model(&models.ShopSettings{}).Count(&count)
	if count == 0 {
		DB.Create(&models.ShopSettings{
			Name:    "ชื่อร้านของคุณ",
			Address: "ที่อยู่ร้าน (แก้ไขได้ที่หน้าตั้งค่าร้านค้า)",
			Phone:   "",
			TaxID:   "",
		})
	}
}

// ensureLoyaltySettings สร้างแถวค่าตั้งค่าระบบสะสมแต้ม (id=1) ให้ครั้งแรกถ้ายังไม่มี ใส่ค่าเริ่มต้นที่พอใช้งาน
// ได้จริงไว้ก่อน ให้ admin มาแก้ที่หน้า "สมาชิก > ตั้งค่า" ทีหลัง (รันซ้ำได้ ไม่ error ถ้ามีแถวอยู่แล้วจะข้ามไป)
func ensureLoyaltySettings() {
	var count int64
	DB.Model(&models.LoyaltySettings{}).Count(&count)
	if count == 0 {
		DB.Create(&models.LoyaltySettings{
			IsEnabled: true,
			Accumulation: models.PointAccumulationRule{
				SpendPerPoint:    25,
				PointsExpiryDays: 365,
				MinSpendToEarn:   0,
			},
			Redemption: models.RedemptionRule{
				PointsPerBaht:     10,
				MinPointsToRedeem: 50,
				MaxDiscountRatio:  0.2,
			},
			TierRules: models.TierRules{
				{Tier: "bronze", Label: "Bronze", MinTotalSpent: 0, PointsMultiplier: 1},
				{Tier: "silver", Label: "Silver", MinTotalSpent: 5000, PointsMultiplier: 1.25},
				{Tier: "gold", Label: "Gold", MinTotalSpent: 10000, PointsMultiplier: 1.5},
				{Tier: "platinum", Label: "Platinum", MinTotalSpent: 30000, PointsMultiplier: 2},
			},
		})
	}
}

// backfillZonesFromTables สร้างแถวใน Zone ให้ครบตามชื่อโซนที่โต๊ะเก่าเคยตั้งไว้ (เป็น string) มาก่อนที่จะมี
// ตาราง Zone แยกต่างหาก เพื่อไม่ให้โซนที่มีอยู่แล้วหายไปจากหน้าจัดการโซน หลังอัปเดตโค้ด (รันซ้ำได้ ไม่ error)
func backfillZonesFromTables() {
	var zoneNames []string
	DB.Model(&models.DiningTable{}).
		Where("zone IS NOT NULL AND zone != ''").
		Distinct().
		Pluck("zone", &zoneNames)

	for _, name := range zoneNames {
		var existing models.Zone
		if DB.Where("name = ?", name).First(&existing).Error != nil {
			DB.Create(&models.Zone{Name: name, IsActive: true})
		}
	}
}
