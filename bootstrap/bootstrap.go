// Package bootstrap รวม logic การเตรียมระบบตอนเริ่มทำงาน (ต่อฐานข้อมูล + seed ข้อมูลเริ่มต้น + เตรียม
// โฟลเดอร์อัปโหลด) ไว้ที่เดียว ให้ทั้ง main.go (รันเป็นเซิร์ฟเวอร์ปกติ) และ api/index.go (Vercel serverless
// function) เรียกใช้ร่วมกันได้ — สำคัญมากสำหรับฝั่ง Vercel เพราะต้องเรียกแค่ครั้งเดียวตอน cold start
// ไม่ใช่ทุก request (ดู sync.Once ใน api/index.go)
package bootstrap

import (
	"log"
	"os"

	"pos-backend/db"
	"pos-backend/seed"
)

// Init ต่อฐานข้อมูล (SQLite หรือ Postgres แล้วแต่ env — ดู db.Init) + seed ข้อมูลเริ่มต้น + เตรียมโฟลเดอร์
// อัปโหลดไฟล์ในเครื่อง (ข้ามขั้นตอนนี้ถ้าตั้งค่าใช้ Supabase Storage ไว้ เพราะไม่จำเป็นต้องมีโฟลเดอร์ในเครื่อง)
func Init() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "pos.db"
	}
	db.Init(dbPath)
	seed.Run()

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
