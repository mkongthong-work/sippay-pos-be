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
- ปิดบิลแบบเงินสด (บันทึกยอดรับ-เงินทอน) หรือโอนเงิน (ถือว่าโอนมาพอดียอด ไม่มีเงินทอน กรอกเลขอ้างอิงการโอน
  และ/หรือแนบรูปสลิปได้ — ไม่บังคับตอนปิดบิล จะมาแนบ/แก้ไขทีหลังที่หน้ารายงานก็ได้) — ยังไม่เชื่อมเครื่องรูดบัตรจริง
- ดูออเดอร์ที่ยังไม่เสร็จ (สำหรับหน้าร้าน) พร้อมสั่งเพิ่ม/ยกเลิกบางรายการในบิลเดิมได้
- หน้าจอครัวแยกต่างหาก ติ๊กรายการที่ทำเสร็จได้ทีละรายการ อิสระจากรายการอื่นในบิลเดียวกัน
- รายงานยอดขาย: แท็บรายวัน (สรุป + กราฟเมนูขายดี + กราฟยอดขายตามช่วงเวลา + ตารางประวัติบิล ใช้ Chart.js)
  และแท็บแนวโน้ม (กราฟเส้นยอดขายรายวัน เลือกช่วง 7 วัน/30 วัน/เดือนนี้ได้)
- ระบบ login พนักงาน (admin / staff)
- หน้าจัดการพนักงาน (เฉพาะ admin) — เพิ่มบัญชีใหม่, เปลี่ยน role, รีเซ็ตรหัสผ่าน, เปิด/ปิดใช้งานบัญชี
- จองโต๊ะ/กันโต๊ะ — ครอบคลุม 3 สถานการณ์ที่ลูกค้าอยากได้โต๊ะ: (1) ลูกค้ามาถึงร้านแล้วรอโต๊ะสักครู่ ให้กันโต๊ะไว้
  ตอนนี้เลย (2) ลูกค้าโทร/แจ้งจองล่วงหน้า ระบุวันเวลาที่จะมา (3) ลูกค้านั่งอยู่แล้วอยากย้ายไปโต๊ะอื่น (ย้ายโต๊ะ
  จากหน้าออเดอร์ได้โดยตรง) โต๊ะที่กัน/จองไว้จะเปลี่ยนสถานะเป็น "จองไว้" (ซ่อนจากลูกค้าคนอื่นที่หน้าขาย แต่
  พนักงานยังเลือกเปิดบิลให้ลูกค้าที่จองไว้ได้ตามปกติ ระบบปิดรายการจองให้อัตโนมัติทันทีที่เปิดบิล)
- พิมพ์ใบเสร็จ — ออกแบบบิลกว้าง 80mm (ตรงกับกระดาษ thermal printer มาตรฐานที่แนะนำไว้) กดพิมพ์ได้จากหน้า
  คิดเงินทันทีหลังปิดบิลสำเร็จ หรือพิมพ์ซ้ำย้อนหลังจากหน้ารายงาน (ประวัติบิล) ก็ได้ ตอนนี้พิมพ์ผ่าน
  เบราว์เซอร์ (window.print()) ไปก่อน เตรียมโครงสร้างข้อมูลไว้ให้ต่อเครื่องพิมพ์ ESC/POS จริงในอนาคตได้ง่าย
  (ดูหัวข้อ 6 Roadmap) มีหน้า "ตั้งค่าร้านค้า" (เฉพาะ admin) ให้กรอกชื่อร้าน/ที่อยู่/เบอร์โทร/เลขผู้เสียภาษี
  โชว์บนหัวใบเสร็จ
- ระบบสมาชิก/สะสมแต้ม (หน้า "สมาชิก") — ลงทะเบียนสมาชิกด้วยชื่อ+เบอร์โทร (ไม่ซ้ำกัน), ค้นหาสมาชิกด้วยชื่อ/เบอร์,
  ดูประวัติแต้มย้อนหลัง, ปรับแต้มด้วยมือ (admin เท่านั้น), ตั้งค่ากฎการสะสม/แลกแต้มและเกณฑ์ระดับสมาชิก
  (bronze/silver/gold/platinum) ได้ที่แท็บ "ตั้งค่า" (admin เท่านั้น) — ยังไม่เชื่อมกับหน้าขาย/คิดเงินจริง
  (ยังไม่มีการสะสม/แลกแต้มอัตโนมัติตอนปิดบิล เป็นแค่ CRUD จัดการสมาชิกและแต้มด้วยมือก่อน)

ไม่รวมใน Phase 1: การพิมพ์ใบเสร็จจริง, การเปิดลิ้นชักเงินอัตโนมัติ, การตัดบัตร/QR ผ่านเกตเวย์จริง, ระบบสต๊อกวัตถุดิบเชิงลึก, การสะสม/แลกแต้มอัตโนมัติตอนปิดบิล, เดลิเวอรี

## 2. Tech Stack

| ส่วน | เทคโนโลยี | เหตุผล |
|---|---|---|
| Backend | Go + Gin (HTTP router) + GORM (ORM) | เขียน API เร็ว, มี ecosystem เยอะ, เหมาะเรียนรู้ backend ครั้งแรก |
| Database | SQLite (ไฟล์เดียว, ไม่ต้องตั้ง server) ผ่าน driver แบบ pure-Go (ไม่ต้องพึ่ง C compiler) | เริ่มง่าย ย้ายไป PostgreSQL/MySQL ทีหลังได้เมื่อมีหลายเครื่อง/หลายสาขา |
| Auth | JWT (JSON Web Token) | ง่าย เหมาะกับ SPA เรียก REST API |
| Frontend | Angular (มาตรฐานเดียวกับที่ถนัดอยู่แล้ว) | ใช้งานผ่านเบราว์เซอร์/แท็บเล็ตได้ทันที ไม่ต้องติดตั้งแอป |
| UI | Tailwind CSS (utility classes) + custom SCSS ในบางหน้าที่ทำไว้ก่อน | ปรับหน้าตาได้เร็วโดยไม่ต้องเขียน CSS เยอะ ใช้ปนกับ SCSS เดิมได้ |
| กราฟ | Chart.js (เรียกใช้ตรงผ่าน `new Chart(canvas, config)` ไม่ผ่าน wrapper library) | เบา ไม่ต้องพึ่ง Angular wrapper เพิ่ม ใช้ที่หน้ารายงานเท่านั้นตอนนี้ |
| ข้อความแจ้งเตือน | ng-zorro-antd (เฉพาะโมดูล `nz-alert`) | ใช้แสดงข้อความ error/success แทน `<p>` ที่ทำเองทุกหน้า ให้หน้าตาเป็นมาตรฐานเดียวกัน |
| พิมพ์ใบเสร็จ | `window.print()` ของเบราว์เซอร์ + CSS `@media print` (component `app-receipt`) | ยังไม่ได้ต่อเครื่องพิมพ์ ESC/POS จริง ใช้กลไกพิมพ์ของเบราว์เซอร์ไปก่อน ออกแบบเลย์เอาต์กว้าง 80mm ให้ตรงกับกระดาษ thermal printer ไว้รอต่อจริงใน Phase 2 |

หมายเหตุ: เมื่อพร้อมเปิดใช้จริงหลายเครื่อง (POS หน้าร้าน + จอครัว + จอผู้จัดการ พร้อมกัน) แนะนำย้าย SQLite → PostgreSQL เพราะ SQLite เขียนพร้อมกันจากหลาย process ได้จำกัด

## 3. โครงสร้างข้อมูล (Database Schema)

```
users            id, username, password_hash, name, role(admin|staff), is_active, created_at
                 ปิดใช้งานบัญชีได้แทนการลบทิ้งจริง (เช่น พนักงานลาออก) เพราะยังมี order/payment เก่าที่
                 อ้างอิง user id นี้อยู่ (created_by/paid_by) บัญชีที่ปิดใช้งานจะ login ไม่ได้ และถ้า token
                 เดิมยังไม่หมดอายุอยู่ก็ถูกตัดสิทธิ์ทันที (เช็คทุก request ผ่าน AuthRequired middleware
                 ไม่ใช่แค่ตอน login) — จัดการที่หน้า "จัดการพนักงาน" (เฉพาะ admin เห็นเมนูนี้)

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

dining_tables    id, name, zone, status(available|occupied|reserved), capacity (จำนวนคนที่นั่งได้)
                 reserved = กัน/จองไว้ให้ลูกค้าคนใดคนหนึ่งแล้ว (ดูตาราง reservations) ยังเลือกที่หน้าขายได้
                 (เพื่อเปิดบิลให้ลูกค้าที่จองไว้) แต่ซ่อนสถานะนี้ไว้ให้พนักงานสังเกตด้วยแท็กชื่อลูกค้า

zones            id, name, is_active
                 เปิด/ปิดใช้งานโซนได้ (เช่น ปิดซ่อม หรือมีการจองที่นั่งไว้ทั้งโซน) ผูกกับ dining_tables.zone
                 แบบชื่อ (string) ไม่ใช่ FK เพราะโต๊ะเดิมเก็บชื่อโซนเป็น string มาก่อนมีตารางนี้ — ตอน migrate
                 ครั้งแรกจะ backfill โซนจากชื่อที่โต๊ะเก่าเคยตั้งไว้ให้อัตโนมัติ โซนที่ปิดใช้งานจะไม่โชว์ในป็อปอัพ
                 เลือกโต๊ะของหน้าขาย (POS) แต่ยังแก้ไขโต๊ะที่อยู่ในโซนนั้นได้ตามปกติที่หน้าจัดการโต๊ะ
                 (หน้าจัดการโต๊ะแยกเป็น 2 แท็บ: "จัดการโต๊ะ" กับ "จัดการโซน")

shop_settings    id(=1 แถวเดียวเสมอ), name, address, phone, tax_id
                 ข้อมูลร้านค้าโชว์บนหัวใบเสร็จตอนพิมพ์บิล แก้ไขได้ที่หน้า "ตั้งค่าร้านค้า" (admin) สร้างแถว
                 default ให้อัตโนมัติตอน migrate ครั้งแรกถ้ายังไม่มี (ensureShopSettings ใน db.go)

reservations     id, table_id(FK), customer_name, customer_phone, party_size, reserved_for(เวลา, null ได้),
                 note, status(active|seated|cancelled|no_show), order_id(FK, null จนกว่าจะเปิดบิล),
                 created_by(FK users), created_at
                 reserved_for = null หมายถึงกันโต๊ะไว้ตอนนี้เลย (ลูกค้ามาถึงร้านแล้ว) มีค่า = จองล่วงหน้าไว้
                 เวลานั้น ตอนสร้างจะเช็คว่าโต๊ะว่างอยู่ก่อน แล้วเปลี่ยน dining_tables.status เป็น reserved ทันที
                 พอพนักงานเปิดบิล (CreateOrder) ที่โต๊ะนี้จริง จะปิดเป็น seated + ผูก order_id ให้อัตโนมัติ
                 ยกเลิก/บันทึกลูกค้าไม่มา (no_show) จะปล่อยโต๊ะกลับเป็น available ให้เอง (เฉพาะกรณีโต๊ะยังอยู่ใน
                 สถานะ reserved จริง กันไม่ให้ไปทับสถานะ occupied ถ้ามีคนเปิดบิลไปแล้ว)

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

payments         id, order_id(FK), method(cash|transfer), subtotal, discount_amount, amount,
                 received_amount, change_amount, transfer_ref, slip_image_path,
                 paid_by(FK users), paid_at
                 method=transfer ถือว่าลูกค้าโอนมาพอดียอด (received_amount=amount, change_amount=0 เสมอ
                 ไม่สนตัวเลขที่ frontend ส่งมา) transfer_ref/slip_image_path ไม่บังคับตอนปิดบิล แนบ/แก้ไข
                 ทีหลังได้ผ่าน PUT /orders/:id/payment (ดูหัวข้อ 5)

members          id, name, phone(uniqueIndex), points_balance, tier(bronze|silver|gold|platinum),
                 total_spent, is_active, created_at, updated_at
                 สมัครสมาชิกใหม่เริ่มที่ points_balance=0/tier=bronze/total_spent=0/is_active=true เสมอ
                 ผูกด้วยเบอร์โทร (ไม่ซ้ำกัน) ใช้ค้นหาเร็วที่หน้าขายได้ผ่าน GET /members/by-phone/:phone
                 ยังไม่มี logic คำนวณ tier อัตโนมัติจาก total_spent หรือสะสม/แลกแต้มอัตโนมัติตอนปิดบิล (Phase 3)

member_point_history  id, member_id(FK), order_id(FK, null ได้ — null คือปรับแต้มด้วยมือจากหน้าจัดการสมาชิก
                       ไม่ได้ผูกกับบิลใด), change(+/-), reason, created_at
                       บันทึกทุกครั้งที่แต้มสมาชิกเปลี่ยน (ตอนนี้มีทางเดียวคือ POST /members/:id/adjust-points)

loyalty_settings id(=1 แถวเดียวเสมอ), is_enabled, accumulation(json), redemption(json), tier_rules(json), updated_at
                 ค่าตั้งค่าระบบสะสมแต้มทั้งร้าน สร้างแถว default ให้อัตโนมัติตอน migrate ครั้งแรกถ้ายังไม่มี
                 (ensureLoyaltySettings ใน db.go) — accumulation/redemption/tier_rules เก็บเป็น JSON ในคอลัมน์
                 เดียว (struct ที่ implement sql.Scanner/driver.Valuer เอง เพราะ SQLite ไม่มี native JSON column
                 และค่าพวกนี้แก้ทั้งชุดพร้อมกันเสมอ ไม่มีที่ไหน query แยกเป็นรายฟิลด์):
                   accumulation: spend_per_point, points_expiry_days, min_spend_to_earn
                   redemption:   points_per_baht, min_points_to_redeem, max_discount_ratio(0-1)
                   tier_rules:   [{ tier, label, min_total_spent, points_multiplier }, ...]
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
POST   /auth/login                       เข้าสู่ระบบ, คืน JWT (บัญชีที่ is_active=false จะ login ไม่ได้)
GET    /auth/me                          ข้อมูลผู้ใช้ปัจจุบัน

GET    /users                            ดึงรายชื่อพนักงานทั้งหมด (admin เท่านั้น รวมบัญชีปิดใช้งานด้วย)
POST   /users                            เพิ่มพนักงานใหม่ (admin) — username, password (>=6 ตัว), name, role
PUT    /users/:id                        แก้ชื่อ/สิทธิ์(role)/รีเซ็ตรหัสผ่าน/เปิด-ปิดใช้งานบัญชี (admin)
                                          ทุกฟิลด์เลือกส่งเฉพาะที่จะแก้ได้ กันแอดมินพลาดลด role ตัวเองออกจาก
                                          admin หรือปิดใช้งานบัญชีตัวเอง (เช็คทั้ง backend และซ่อนปุ่มใน UI)

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

GET    /shop-settings                    ดึงข้อมูลร้านค้า (ชื่อ/ที่อยู่/เบอร์โทร/เลขผู้เสียภาษี) ให้ทุกคนที่
                                          login แล้วเรียกได้ (ใช้โชว์บนหัวใบเสร็จตอนพนักงานพิมพ์บิล)
PUT    /shop-settings                    แก้ไขข้อมูลร้านค้า (admin เท่านั้น)

GET    /zones                            ดึงรายชื่อโซนทั้งหมด (รวมที่ปิดใช้งานอยู่ด้วย)
POST   /zones                            เพิ่มโซนใหม่ (admin)
PUT    /zones/:id                        แก้ชื่อ และ/หรือเปิด-ปิดใช้งานโซน (admin) — ถ้าเปลี่ยนชื่อ จะอัปเดต
                                          dining_tables.zone ของโต๊ะที่ใช้ชื่อเดิมให้เป็นชื่อใหม่ตามไปด้วย

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
PUT    /orders/:id/table                 ย้ายออเดอร์นั่งทานที่เปิดอยู่ไปโต๊ะอื่น (โต๊ะใหม่ต้องว่าง) โต๊ะเดิม
                                          ปล่อยกลับเป็นว่างให้อัตโนมัติ — ใช้ตอนลูกค้านั่งอยู่แล้วอยากย้ายที่นั่ง
POST   /orders/:id/pay                   ปิดบิล (ยอดที่ต้องจ่ายคือยอดสุทธิหลังหักส่วนลด) method=cash บันทึก
                                          ยอดรับ-เงินทอนตามที่ส่งมา method=transfer ถือว่าโอนมาพอดียอดเสมอ
                                          (ไม่คำนวณเงินทอน) ส่ง transfer_ref (เลขอ้างอิงการโอน) มาด้วยได้
                                          ไม่บังคับ ไม่ส่งมาก็ปิดบิลได้ตามปกติ
PUT    /orders/:id/payment               แนบ/แก้ไขเลขอ้างอิงการโอน และ/หรือรูปสลิปของบิลที่ปิดไปแล้ว
                                          (multipart/form-data: ref ข้อความ, slip ไฟล์รูป — ส่งมาแค่ฟิลด์ไหน
                                          ก็แก้ไขเฉพาะฟิลด์นั้น) ใช้ตอนปิดบิลทันทีถ้าถ่ายสลิปพร้อมกันเลย หรือ
                                          ย้อนกลับมาแนบทีหลังจากหน้ารายงานก็ได้ (เช่น พนักงานยุ่งตอนปิดบิล)
                                          รูปที่อัปโหลดเก็บที่ ./uploads/slips เปิดดูตรงผ่าน
                                          /uploads/slips/<ไฟล์> (static route ไม่ผ่าน auth — ใช้ภายในร้าน
                                          เท่านั้น ชื่อไฟล์สุ่มด้วย timestamp กันเดา URL ได้ยากขึ้นระดับหนึ่ง)

GET    /reservations?status=             รายการจอง/กันโต๊ะ (ไม่ส่ง status มา = เฉพาะที่ยัง active/รอดำเนินการ)
POST   /reservations                     กันโต๊ะ/จองโต๊ะใหม่ (table_id, customer_name, customer_phone,
                                          party_size, reserved_for ไม่ส่ง=กันตอนนี้เลย, note) โต๊ะต้องว่างก่อน
PUT    /reservations/:id/cancel          ยกเลิกรายการจอง ปล่อยโต๊ะกลับเป็นว่าง
PUT    /reservations/:id/no-show         บันทึกว่าลูกค้าไม่มาตามนัด ปล่อยโต๊ะกลับเป็นว่าง (แยกสถานะจาก cancel
                                          ไว้เผื่ออยากดูสถิติย้อนหลังว่าลูกค้าไม่มาบ่อยแค่ไหน)

GET    /reports/daily?date=YYYY-MM-DD    ยอดขายรวม, จำนวนบิล, เมนูขายดี, และ orders[] (รายบิลที่ปิดวันนั้น
                                          พร้อม created_by_name/paid_by_name/payment_method/transfer_ref/
                                          slip_image_path) ให้ admin เช็คย้อนหลังได้ว่าใครเปิด/ปิดบิล จ่ายด้วย
                                          วิธีไหน มีสลิปแนบหรือยัง — Order preload CreatedByUser, Payment
                                          preload PaidByUser ด้วยเช่นกัน (ใช้ที่หน้าออเดอร์/คิดเงิน/รายงาน)
GET    /reports/range?from=&to=          ยอดขายรวมรายวันของทุกวันในช่วง (YYYY-MM-DD ทั้งคู่ รวมสูงสุด 366
                                          วัน) วันไหนไม่มีบิลเลยจะเติมยอดขาย 0 ให้ (กราฟจะได้ไม่ขาดช่วง)
                                          ใช้ทำกราฟแนวโน้มที่แท็บ "แนวโน้ม" ของหน้ารายงาน (preset 7/30 วัน
                                          หรือเดือนนี้ คำนวณช่วง from/to ฝั่ง frontend เอง)

GET    /members?search=                  รายชื่อสมาชิกทั้งหมด (รวมที่ปิดใช้งานอยู่ด้วย) กรองด้วยชื่อ/เบอร์โทรได้
GET    /members/:id                      ข้อมูลสมาชิกรายเดียว
GET    /members/by-phone/:phone          ค้นหาสมาชิกด้วยเบอร์โทรตรงๆ (ใช้เร็วที่หน้าขาย)
POST   /members                          สมัครสมาชิกใหม่ (name, phone — เบอร์โทรต้องเป็นตัวเลข 9-10 หลัก
                                          และไม่ซ้ำกับสมาชิกที่มีอยู่) เริ่มต้นแต้ม 0/tier bronze/ยอดใช้จ่าย 0
PUT    /members/:id                      แก้ชื่อ/เบอร์โทร/เปิด-ปิดใช้งานสมาชิกภาพ ทุกฟิลด์เลือกส่งเฉพาะที่จะแก้ได้
GET    /members/:id/history              ประวัติการเปลี่ยนแปลงแต้มของสมาชิกรายนี้ (ใหม่สุดไว้บนสุด)
POST   /members/:id/adjust-points        ปรับแต้มด้วยมือ (admin) — change(+/-), reason บันทึกลง
                                          member_point_history ทุกครั้ง กันแต้มติดลบ (แลก/หักเกินกว่าที่มี)

GET    /loyalty-settings                 ค่าตั้งค่าระบบสะสมแต้มทั้งร้าน ให้ทุกคนที่ login แล้วเรียกได้
PUT    /loyalty-settings                 แก้ไขค่าตั้งค่าระบบสะสมแต้ม (admin เท่านั้น) — แก้ is_enabled,
                                          accumulation, redemption, tier_rules ทั้งชุดพร้อมกันเสมอ
```

## 6. Roadmap หลังจาก MVP

**Phase 2 — ต่อฮาร์ดแวร์:** เชื่อมเครื่องพิมพ์ใบเสร็จจริง (ESC/POS ผ่าน USB/LAN หรือ cloud print — component
`app-receipt` ที่มีอยู่แล้วออกแบบข้อมูล (order + shop settings) ไว้ให้แปลงเป็นคำสั่ง ESC/POS ต่อได้เลย ไม่ต้อง
ออกแบบเลย์เอาต์บิลใหม่ แค่เปลี่ยนจาก `window.print()` เป็นส่งคำสั่งไปเครื่องพิมพ์จริง เช่น ผ่าน raw TCP socket
จาก backend ถ้าใช้เครื่องพิมพ์ LAN หรือผ่านตัวกลางอย่าง QZ Tray ถ้าต้องพิมพ์จากหลายเครื่อง/แท็บเล็ต),
เปิดลิ้นชักเงินอัตโนมัติตอนปิดบิลเงินสด (สั่งงานผ่านเครื่องพิมพ์ที่รองรับ cash-drawer kick), เชื่อม QR
พร้อมเพย์ผ่านผู้ให้บริการ (เช่น เจ้าธนาคาร/ผู้ให้บริการ payment gateway) และเครื่องรูดบัตร EDC

**Phase 3 — ขยายฟีเจอร์:** ระบบสต๊อกวัตถุดิบ (ตัดสต๊อกอัตโนมัติตามสูตร), เชื่อมระบบสมาชิก/สะสมแต้มเข้ากับหน้าขาย/คิดเงินจริง (สะสมแต้มอัตโนมัติตามยอดซื้อ, แลกแต้มเป็นส่วนลดตอนปิดบิล, อัปเดต tier อัตโนมัติตามยอดใช้จ่ายสะสม — ตัว CRUD สมาชิก/ตั้งค่ากฎ ทำไว้แล้วใน Phase 1), พิมพ์ใบสั่งครัวแยกโซน (บาร์กาแฟ vs ครัวอาหาร), รองรับหลายสาขา, เชื่อม LINE/เดลิเวอรี, dashboard รายงานเชิงลึก

## 7. โครงสร้างโปรเจกต์ที่สร้างให้

```
pos-project/
  docs/ARCHITECTURE.md      เอกสารนี้
  backend/                  Go + Gin + GORM + SQLite (pure-Go driver)
  frontend/                 Angular + Tailwind CSS
```

ดูวิธีรันแต่ละฝั่งใน `backend/README.md` และ `frontend/README.md`
