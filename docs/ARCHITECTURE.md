# ระบบ POS ร้านกาแฟ + ร้านอาหาร (นั่งทาน/ซื้อกลับ) — เอกสารออกแบบระบบ

## 1. ขอบเขต Phase 1 (MVP ปัจจุบัน)

เฟสนี้เป็นซอฟต์แวร์ล้วน ยังไม่ผูกกับฮาร์ดแวร์ (เครื่องพิมพ์ใบเสร็จ / ลิ้นชักเงิน / เครื่องรูดบัตร-สแกน QR) โฟกัสที่การรับออเดอร์และปิดบิลให้ถูกต้องก่อน แล้วค่อยต่อฮาร์ดแวร์ทีหลังใน Phase 2

รองรับ:
- ออเดอร์ 2 แบบ: นั่งทาน (เลือกโต๊ะ) และ ซื้อกลับบ้าน (ไม่มีโต๊ะ) — ออเดอร์นั่งทานติ๊กแยกได้เป็นรายชิ้นว่า
  รายการไหนอยากสั่งกลับบ้านด้วย (อยู่บิล/จ่ายเงินครั้งเดียวกัน แค่แท็กแยกไว้ให้ครัวแพ็ค)
- popup ตรวจสอบรายการก่อนส่งออเดอร์จริงทุกครั้ง (กันสั่งผิด/ลืมเช็ค) และ popup โชว์เลขออเดอร์ใหญ่ๆ
  ทันทีหลังส่งสำเร็จสำหรับออเดอร์ซื้อกลับ (ไว้บอกลูกค้าว่าต้องรอเรียกเลขไหน)
- จัดการเมนู/หมวดหมู่ (กาแฟ เครื่องดื่ม อาหาร ของหวาน ฯลฯ)
- เพิ่ม/แก้ไข/ลบรายการในบิลก่อนปิดบิล
- ปิดบิลแบบเงินสด (บันทึกยอดรับ-เงินทอน) — ยังไม่เชื่อมเครื่องรูดบัตรจริง
- ดูออเดอร์ที่ยังไม่เสร็จ (สำหรับหน้าร้าน) พร้อมสั่งเพิ่ม/ยกเลิกบางรายการในบิลเดิมได้
- หน้าจอครัวแยกต่างหาก ติ๊กรายการที่ทำเสร็จได้ทีละรายการ อิสระจากรายการอื่นในบิลเดียวกัน
- รายงานยอดขายรายวันเบื้องต้น
- ระบบ login พนักงาน (admin / staff)

ไม่รวมใน Phase 1: การพิมพ์ใบเสร็จจริง, การเปิดลิ้นชักเงินอัตโนมัติ, การตัดบัตร/QR ผ่านเกตเวย์จริง, ระบบสต๊อกวัตถุดิบเชิงลึก, ระบบสมาชิก/แต้ม, เดลิเวอรี

## 2. Tech Stack

| ส่วน | เทคโนโลยี | เหตุผล |
|---|---|---|
| Backend | Go + Gin (HTTP router) + GORM (ORM) | เขียน API เร็ว, มี ecosystem เยอะ, เหมาะเรียนรู้ backend ครั้งแรก |
| Database | SQLite (ไฟล์เดียว, ไม่ต้องตั้ง server) ผ่าน driver แบบ pure-Go (ไม่ต้องพึ่ง C compiler) | เริ่มง่าย ย้ายไป PostgreSQL/MySQL ทีหลังได้เมื่อมีหลายเครื่อง/หลายสาขา |
| Auth | JWT (JSON Web Token) | ง่าย เหมาะกับ SPA เรียก REST API |
| Frontend | Angular (มาตรฐานเดียวกับที่ถนัดอยู่แล้ว) | ใช้งานผ่านเบราว์เซอร์/แท็บเล็ตได้ทันที ไม่ต้องติดตั้งแอป |
| UI | Tailwind CSS (utility classes) + custom SCSS ในบางหน้าที่ทำไว้ก่อน | ปรับหน้าตาได้เร็วโดยไม่ต้องเขียน CSS เยอะ ใช้ปนกับ SCSS เดิมได้ |

หมายเหตุ: เมื่อพร้อมเปิดใช้จริงหลายเครื่อง (POS หน้าร้าน + จอครัว + จอผู้จัดการ พร้อมกัน) แนะนำย้าย SQLite → PostgreSQL เพราะ SQLite เขียนพร้อมกันจากหลาย process ได้จำกัด

## 3. โครงสร้างข้อมูล (Database Schema)

```
users            id, username, password_hash, name, role(admin|staff), created_at

categories       id, name, sort_order, station (ครัว/บาร์/อื่นๆ หรือพิมพ์เอง — เก็บไว้เผื่ออนาคต
                 แยกจอทำงานตามสถานี ยังไม่มีหน้าไหนกรองจริงตอนนี้)

category_option_templates        id, category_id(FK), name, is_required, sort_order
                                  ค่าเริ่มต้นของกลุ่มตัวเลือกที่ตั้งไว้ระดับหมวดหมู่ เช่น หมวด "กาแฟ"
                                  ตั้ง "ความหวาน" ไว้ครั้งเดียว แล้วนำไป "ใช้" กับเมนูกาแฟหลายอย่างได้

category_option_template_choices id, template_id(FK), name, price_delta, sort_order

menu_items       id, category_id(FK), name, price, is_available, created_at

menu_option_groups   id, menu_item_id(FK), name, is_required, sort_order
                     เช่น "ความหวาน" ของเมนูกาแฟ ตั้งค่าตอนสร้าง/แก้ไขเมนูในหน้าจัดการเมนู

menu_option_choices  id, option_group_id(FK), name, price_delta, sort_order
                     เช่น "หวานน้อย" price_delta=0, "ไซส์ใหญ่" price_delta=+10

dining_tables    id, name, zone, status(available|occupied), capacity (จำนวนคนที่นั่งได้)

orders           id, order_type(dine_in|takeaway), table_id(FK, null ได้), guest_count
                 (จำนวนคนที่มา ไม่บังคับ ตั้งตอนสร้างที่ POS หรือแก้ทีหลังที่หน้าออเดอร์ก็ได้),
                 status(open|preparing|served|paid|cancelled),
                 subtotal, discount_type(none|amount|percent), discount_value, discount_amount,
                 total_amount, created_by(FK users), created_at, updated_at

order_items      id, order_id(FK), menu_item_id(FK), quantity, unit_price,
                 note, status(pending|preparing|served), is_takeaway
                 is_takeaway ใช้กรณีออเดอร์นั่งทาน แต่บางรายการอยากสั่งกลับบ้านด้วย (อยู่บิลเดียวกัน)

order_item_options  id, order_item_id(FK), option_group_id, group_name,
                    choice_id, choice_name, price_delta
                    เก็บเป็น snapshot ตอนสั่ง (ไม่อ้างอิงกลับไปที่ menu_option_choices ตรงๆ)
                    เพื่อไม่ให้ออเดอร์เก่าเปลี่ยนไปตามการแก้เมนูภายหลัง

payments         id, order_id(FK), method(cash), subtotal, discount_amount, amount,
                 received_amount, change_amount, paid_by(FK users), paid_at
```

ความสัมพันธ์หลัก: `orders 1—N order_items`, `order_items 1—N order_item_options`, `menu_items 1—N menu_option_groups 1—N menu_option_choices`, `categories 1—N category_option_templates 1—N category_option_template_choices`, `orders 1—1 payments` (เมื่อปิดบิลแล้ว), `dining_tables 1—N orders` (โต๊ะหนึ่งมีได้หลายออเดอร์ตามเวลา)

หมายเหตุ: การ "ใช้" (apply) ค่าเริ่มต้นของหมวดหมู่กับเมนูใดเมนูหนึ่ง จะ copy ข้อมูลจาก `category_option_templates` ไปสร้างเป็น `menu_option_groups`/`menu_option_choices` จริงของเมนูนั้น ไม่ได้ผูก reference ตรงๆ เพื่อไม่ให้แก้ค่าเริ่มต้นภายหลังกระทบเมนูที่เคย apply ไปแล้ว

## 4. Order Flow (State Machine)

```
สร้างออเดอร์ใหม่ → open
  → staff กดเริ่มทำ → preparing
  → เสิร์ฟแล้ว → served
  → ปิดบิล/รับเงิน → paid
  (ทุกสถานะยกเลิกได้ → cancelled)
```

หน้า "จัดการเมนู" และหน้า "หน้าขาย (POS)" เป็นคนละหน้ากัน: พนักงานเปิดหน้าขายเพื่อรับออเดอร์ทั้งวัน ส่วนหน้าจัดการเมนูเปิดเฉพาะตอนแก้ราคา/เพิ่มเมนูใหม่

## 5. API Endpoints (REST, prefix `/api`)

```
POST   /auth/login                       เข้าสู่ระบบ, คืน JWT
GET    /auth/me                          ข้อมูลผู้ใช้ปัจจุบัน

GET    /categories                       ดึงหมวดหมู่ทั้งหมด
POST   /categories                       เพิ่มหมวดหมู่ (admin)
PUT    /categories/:id
DELETE /categories/:id

GET    /menu-items?category_id=          ดึงเมนู (กรองตามหมวดได้) พร้อม option_groups.choices
POST   /menu-items                       เพิ่มเมนู (admin)
PUT    /menu-items/:id
DELETE /menu-items/:id

POST   /menu-items/:id/option-groups     สร้างกลุ่มตัวเลือกใหม่เฉพาะเมนูนี้ + ตัวเลือกย่อยในครั้งเดียว (admin)
DELETE /option-groups/:id                ลบกลุ่มตัวเลือก (admin)
POST   /option-groups/:id/choices        เพิ่มตัวเลือกย่อยเข้ากลุ่มที่มีอยู่ (admin)
DELETE /choices/:id                      ลบตัวเลือกย่อย (admin)

GET    /categories/:id/option-templates                       ดึงค่าเริ่มต้นตัวเลือกของหมวดหมู่
POST   /categories/:id/option-templates                       สร้างค่าเริ่มต้นใหม่ของหมวดหมู่ (admin)
DELETE /option-templates/:id                                   ลบค่าเริ่มต้น (admin)
POST   /menu-items/:id/option-groups/from-template/:templateId นำค่าเริ่มต้นของหมวดหมู่มาใช้กับเมนูนี้
                                                                (copy เป็นกลุ่มตัวเลือกจริงของเมนู, admin)

GET    /tables                           ดึงรายชื่อโต๊ะ + สถานะ
POST   /tables                           เพิ่มโต๊ะ (admin)
PUT    /tables/:id

POST   /orders                           สร้างออเดอร์ใหม่ (dine_in/takeaway + รายการเริ่มต้น
                                          แต่ละรายการระบุ option_choice_ids[], note, is_takeaway ได้)
GET    /orders?status=open,preparing     ออเดอร์ที่ยังไม่ปิด (หน้าจอครัว/หน้าร้าน)
                                          preload ถึง Items.MenuItem.Category ด้วย เพื่อให้จอครัวรู้ station
                                          (ครัว/บาร์/อื่นๆ) ของแต่ละรายการ ไปจัดกลุ่มแสดงในออเดอร์เดียวกันได้
                                          จอครัวยังเรียงออเดอร์ใหม่สุดไว้บนสุดเสมอ (ตาม created_at) และโชว์
                                          เวลาส่ง + นับนาทีที่ผ่านมาแบบ live (ไม่ต้อง refresh)
GET    /orders/:id                       รายละเอียดออเดอร์
POST   /orders/:id/items                 เพิ่มรายการเข้าออเดอร์เดิม (ใช้ตอน "สั่งเพิ่ม" ที่โต๊ะเดิม)
PUT    /orders/:id/items/:itemId         แก้จำนวน/สถานะ/is_takeaway ของรายการ (status: pending|preparing|served)
                                          หน้าจอครัวมีปุ่มต่อรายการ "เริ่มทำ" (pending->preparing), "เสร็จ"
                                          (preparing->served), "ย้อนกลับ" (ย้อน 1 ขั้นเผื่อกดพลาด)
                                          ทุกครั้งที่สถานะรายชิ้นเปลี่ยน backend จะเรียก
                                          autoSyncOrderStatusFromItems ปรับสถานะบิลโดยรวมให้อัตโนมัติ:
                                          ครบทุกชิ้น -> served, เริ่มทำอย่างน้อย 1 ชิ้นแต่ยังไม่ครบ -> preparing,
                                          ถอยกลับ pending หมดทุกชิ้น -> open ไม่ต้องกดปุ่มเปลี่ยนสถานะบิลเอง
DELETE /orders/:id/items/:itemId         ลบรายการออกจากบิล (ใช้ตอน "ยกเลิก" บางรายการ)
PUT    /orders/:id/status                เปลี่ยนสถานะออเดอร์ (หน้าออเดอร์เหลือปุ่ม manual แค่ "กำลังเริ่มทำ"
                                          คือ open -> preparing เท่านั้น ส่วน preparing -> served เป็น auto)
PUT    /orders/:id/discount              ตั้ง/แก้ไขส่วนลด (none|amount|percent) คำนวณยอดสุทธิใหม่ทันที
PUT    /orders/:id/guests                แก้ไขจำนวนคนที่มา (ตั้งตอนสร้างที่ POS ได้ หรือแก้ทีหลังที่หน้าออเดอร์)
POST   /orders/:id/pay                   ปิดบิล บันทึกยอดรับ-เงินทอน (ยอดที่ต้องจ่ายคือยอดสุทธิหลังหักส่วนลด)

GET    /reports/daily?date=YYYY-MM-DD    ยอดขายรวม, จำนวนบิล, เมนูขายดีของวันนั้น
```

## 6. Roadmap หลังจาก MVP

**Phase 2 — ต่อฮาร์ดแวร์:** เชื่อมเครื่องพิมพ์ใบเสร็จ (ผ่าน ESC/POS ผ่าน USB/LAN หรือ cloud print), เปิดลิ้นชักเงินอัตโนมัติตอนปิดบิลเงินสด (สั่งงานผ่านเครื่องพิมพ์ที่รองรับ cash-drawer kick), เชื่อม QR พร้อมเพย์ผ่านผู้ให้บริการ (เช่น เจ้าธนาคาร/ผู้ให้บริการ payment gateway) และเครื่องรูดบัตร EDC

**Phase 3 — ขยายฟีเจอร์:** ระบบสต๊อกวัตถุดิบ (ตัดสต๊อกอัตโนมัติตามสูตร), ระบบสมาชิก/สะสมแต้ม, พิมพ์ใบสั่งครัวแยกโซน (บาร์กาแฟ vs ครัวอาหาร), รองรับหลายสาขา, เชื่อม LINE/เดลิเวอรี, dashboard รายงานเชิงลึก

## 7. โครงสร้างโปรเจกต์ที่สร้างให้

```
pos-project/
  docs/ARCHITECTURE.md      เอกสารนี้
  backend/                  Go + Gin + GORM + SQLite (pure-Go driver)
  frontend/                 Angular + Tailwind CSS
```

ดูวิธีรันแต่ละฝั่งใน `backend/README.md` และ `frontend/README.md`
