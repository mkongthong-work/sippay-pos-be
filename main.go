// main.go คือ entrypoint สำหรับรันเป็นเซิร์ฟเวอร์ปกติ (บนเครื่อง/VM ของตัวเอง) — ถ้า deploy บน Vercel
// จะไม่ได้ใช้ไฟล์นี้เลย ใช้ api/index.go (serverless function) แทน ซึ่งเรียก bootstrap.Init() +
// router.New() ชุดเดียวกันนี้ ไม่ได้เขียน logic ซ้ำ
package main

import (
	"log"
	"os"

	"pos-backend/bootstrap"
	"pos-backend/router"
)

func main() {
	bootstrap.Init()

	r := router.New()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("POS backend listening on :%s", port)
	r.Run(":" + port)
}
