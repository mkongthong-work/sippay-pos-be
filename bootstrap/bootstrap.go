// Package bootstrap รวม logic การเตรียมระบบตอนเริ่มทำงาน (ต่อฐานข้อมูล + migrate + seed ข้อมูลเริ่มต้น +
// เตรียมโฟลเดอร์อัปโหลด) ไว้ที่เดียว ให้ทั้ง main.go (รันเป็นเซิร์ฟเวอร์ปกติ) และ api/index.go (Vercel
// serverless function) เรียกใช้ร่วมกันได้ — สำคัญมากสำหรับฝั่ง Vercel เพราะต้องเรียกแค่ครั้งเดียวตอน
// cold start ไม่ใช่ทุก request (ดู sync.Once ใน api/index.go)
package bootstrap

import (
	"log"
	"os"

	"pos-backend/db"
	"pos-backend/seed"
)

// Init ต่อฐานข้อมูล (SQLite หรือ Postgres แล้วแต่ env — ดู db.Connect) เสมอ แต่ "migrate ตาราง + seed
// ข้อมูลเริ่มต้น" จะรันเฉพาะตอนไม่ได้อยู่บน Vercel เท่านั้น (เช็คจาก env VERCEL ที่ Vercel ตั้งให้อัตโนมัติทุก
// deployment) เพราะ AutoMigrate ผ่าน Supabase pooler ใช้เวลานาน (แต่ละ query มี latency ~200ms คูณด้วย
// จำนวนตาราง) เกิน 30 วินาทีที่ Vercel ให้ function boot ก่อนตัดการเชื่อมต่อ — ต้อง migrate ครั้งเดียวแยก
// ต่างหากด้วย `go run ./cmd/migrate` (ดู backend/cmd/migrate/main.go และ DEPLOY_VERCEL.md) ก่อน deploy
// จริงครั้งแรก หรือทุกครั้งที่แก้ struct ใน models/
//
// ถ้าจำเป็นต้องบังคับ migrate ตอนรันบน Vercel จริงๆ (ไม่แนะนำ เสี่ยง timeout) ตั้ง env
// SIPPAY_FORCE_MIGRATE=1 ได้
func Init() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "pos.db"
	}
	db.Connect(dbPath)

	runningOnVercel := os.Getenv("VERCEL") == "1"
	if !runningOnVercel || os.Getenv("SIPPAY_FORCE_MIGRATE") == "1" {
		db.Migrate()
		seed.Run()
	}

	if os.Getenv("SUPABASE_STORAGE_BUCKET") == "" {
		// โฟลเดอร์เก็บรูปสลิปโอนเงิน/รูปเมนูที่อัปโหลดไว้ (ดู handlers/payment.go, handlers/menu.go)
		if err := os.MkdirAll("./uploads/slips", 0755); err != nil {
			log.Fatalf("failed to create uploads directory: %v", err)
		}
		if err := os.MkdirAll("./uploads/menu-items", 0755); err != nil {
			log.Fatalf("failed to create uploads directory: %v", err)
		}
	}
}
