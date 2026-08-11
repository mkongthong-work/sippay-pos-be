// Package assets ฝังไฟล์ font ที่ต้องใช้ตอนรันไทม์ไว้ในตัวไบนารีเลย (go:embed) เพื่อไม่ต้องพกไฟล์ .ttf
// แยกไปกับ binary ตอน deploy ตอนนี้ใช้สำหรับสร้าง PDF ใบเสร็จ/ใบกำกับภาษีที่มีข้อความภาษาไทย
// (ดู handlers/invoice.go) — ฟอนต์ Sarabun สัญญาอนุญาต SIL Open Font License 1.1 (ดู fonts/OFL.txt)
package assets

import _ "embed"

//go:embed fonts/Sarabun-Regular.ttf
var SarabunRegularTTF []byte

//go:embed fonts/Sarabun-Bold.ttf
var SarabunBoldTTF []byte
