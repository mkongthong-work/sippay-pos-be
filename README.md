# POS Backend (Go + Gin + GORM + SQLite)

## ติดตั้งและรัน

ต้องมี Go >= 1.22 ติดตั้งในเครื่อง ([ดาวน์โหลด](https://go.dev/dl/))

```bash
cd backend
cp .env.example .env      # แก้ JWT_SECRET เป็นค่าของตัวเอง
go mod tidy                # ดึง dependency ทั้งหมดครั้งแรก (ต้องต่อเน็ต)
go run .
```

เซิร์ฟเวอร์จะรันที่ `http://localhost:8080` และสร้างไฟล์ฐานข้อมูล `pos.db` ให้อัตโนมัติ พร้อมข้อมูลตัวอย่าง (เมนู, โต๊ะ, ผู้ใช้แอดมิน)

**ผู้ใช้เริ่มต้น:** username `admin` / password `admin1234` (ควรเปลี่ยนก่อนใช้งานจริง)

## ทดสอบเรียก API

```bash
# login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin1234"}'

# เอา token ที่ได้มาใช้เรียก endpoint อื่น
curl http://localhost:8080/api/menu-items \
  -H "Authorization: Bearer <token>"
```

## โครงสร้างโค้ด

```
main.go            จุดเริ่มต้น, ตั้งค่า route ทั้งหมด
db/                 เชื่อมต่อฐานข้อมูล + auto-migrate
models/             struct ของตาราง (User, Category, MenuItem, DiningTable, Order, OrderItem, Payment, Member)
handlers/           logic ของแต่ละ endpoint แยกตามหมวด (auth, menu, table, order, report, member)
middleware/         ตรวจสอบ JWT token และสิทธิ์ admin
utils/              สร้าง/ตรวจสอบ JWT
seed/               ใส่ข้อมูลตัวอย่างตอนรันครั้งแรก
```

## หมายเหตุสำคัญ

- ใช้ SQLite driver แบบ pure-Go (`glebarez/sqlite`) จึงไม่ต้องติดตั้ง C compiler (gcc) เพื่อ build
- เมื่อพร้อมเปิดหลายเครื่องพร้อมกันจริงจัง (POS + จอครัว + จอผู้จัดการ) แนะนำย้ายไป PostgreSQL — โครงสร้างโค้ดออกแบบผ่าน GORM ไว้แล้ว เปลี่ยนแค่ driver ใน `db/db.go`
- Phase ถัดไป (เชื่อมเครื่องพิมพ์ใบเสร็จ/ลิ้นชักเงิน/เครื่องรูดบัตร) จะเพิ่ม endpoint และ integration ใหม่โดยไม่ต้องรื้อโครงสร้างเดิม
