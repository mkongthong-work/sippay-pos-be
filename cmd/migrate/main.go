// คำสั่งแยกต่างหากสำหรับ migrate ตาราง + seed ข้อมูลเริ่มต้นครั้งเดียว — ต้องรันครั้งนี้ก่อน deploy ขึ้น
// Vercel ครั้งแรกเสมอ (และทุกครั้งที่แก้ struct ใน models/ เพิ่มฟิลด์ใหม่) เพราะตัว serverless function เอง
// ไม่ migrate อัตโนมัติทุก cold start แล้ว (ช้าเกินไป เสี่ยง timeout — ดู bootstrap/bootstrap.go)
//
// วิธีใช้ (รันบนเครื่องที่มี Go และเข้าอินเทอร์เน็ตได้ — ต่อ Supabase ตรงๆ ไม่ผ่าน Vercel):
//
//	cd backend
//	DATABASE_URL="postgresql://postgres.xxxx:รหัสผ่านจริง@aws-0-xxxx.pooler.supabase.com:6543/postgres" \
//	  go run ./cmd/migrate
//
// (เอา connection string จาก Supabase → Project Settings → Database → Connection string → Transaction
// pooler เหมือนตอนตั้งเป็น env DATABASE_URL บน Vercel) ใช้เวลาสัก 20-40 วินาที รอจนขึ้น "migrate สำเร็จ"
package main

import (
	"log"
	"os"

	"pos-backend/db"
	"pos-backend/seed"
)

func main() {
	if os.Getenv("DATABASE_URL") == "" {
		log.Fatal("ต้องตั้ง env DATABASE_URL ก่อนรันคำสั่งนี้ (connection string ของ Supabase Postgres)")
	}

	db.Connect("")
	log.Println("กำลัง migrate ตาราง...")
	db.Migrate()
	log.Println("migrate ตารางสำเร็จ กำลัง seed ข้อมูลเริ่มต้น (ถ้ายังไม่เคย seed)...")
	seed.Run()
	log.Println("เสร็จแล้ว! เข้าระบบครั้งแรกด้วย admin / admin1234 แล้วเปลี่ยนรหัสผ่านทันที")
}
