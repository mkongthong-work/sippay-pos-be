package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User คือพนักงาน/แอดมินที่ล็อกอินเข้าระบบ
type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string `json:"-" gorm:"not null"`
	Name         string `json:"name"`
	Role         string `json:"role" gorm:"not null;default:staff"` // admin | staff
	// ปิดใช้งานบัญชีได้ (เช่น พนักงานลาออก) โดยไม่ต้องลบทิ้งจริง เพราะยังมี order/payment เก่าที่อ้างอิง
	// user id นี้อยู่ (created_by/paid_by) ถ้าลบทิ้งจะเช็คย้อนหลังไม่ได้ว่าใครเปิด/ปิดบิลนั้น
	IsActive bool `json:"is_active" gorm:"not null;default:true"`
	// PIN 4-6 หลัก สำหรับ "เข้าระบบด้วย PIN" บนเครื่อง POS ที่ใช้ร่วมกัน (ไม่บังคับ — แอดมินเป็นคนตั้ง/แก้ให้
	// ที่หน้า "จัดการพนักงาน" เท่านั้น) เก็บเป็น hash เหมือน PasswordHash ไม่เก็บ PIN ดิบ ไม่ส่งออกไปกับ response
	PinHash string `json:"-" gorm:"column:pin_hash"`
	// เวลาที่ตั้ง/แก้ไข PIN ล่าสุด (เป็น null ถ้ายังไม่เคยตั้งหรือถูกลบ PIN ทิ้งไปแล้ว) โชว์ที่หน้าแก้ไขพนักงาน
	// ให้แอดมินเห็นว่า PIN นี้อัปเดตครั้งล่าสุดเมื่อไหร่
	PinUpdatedAt *time.Time `json:"pin_updated_at"`
	// คำนวณจาก PinHash อัตโนมัติทุกครั้งที่โหลดจาก DB (ดู AfterFind ด้านล่าง) ไม่ได้เก็บเป็นคอลัมน์จริง
	// (gorm:"-") ใช้บอก frontend ว่าคนนี้ตั้ง PIN ไว้แล้วหรือยัง โดยไม่ต้องหลุด hash ออกไปกับ response
	HasPin    bool      `json:"has_pin" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// AfterFind ทำงานอัตโนมัติทุกครั้งที่ GORM โหลด User ขึ้นมาจาก DB (First/Find/...) ใช้เติมค่า HasPin
// แบบ derived field เดียว ไม่ต้องไปเซ็ตซ้ำเองทุก handler ที่ query user
func (u *User) AfterFind(tx *gorm.DB) error {
	u.HasPin = u.PinHash != ""
	return nil
}

// Category คือหมวดหมู่เมนู เช่น กาแฟ, อาหาร, ของหวาน
// Station คือสถานที่ทำของหมวดนี้ เช่น "ครัว"/"บาร์"/"อื่นๆ" — หน้าจอครัว (kitchen) ยังอ่านค่านี้อยู่
// เพื่อจัดกลุ่มรายการ แต่หน้าจัดการเมนู (menu-admin) รุ่นใหม่เลิกให้แก้ไขฟิลด์นี้แล้วตามดีไซน์ใหม่
// (คอลัมน์ยังอยู่ในฐานข้อมูล ค่าเก่าที่เคยตั้งไว้จะยังกรุ๊ปได้ตามปกติ แค่ตั้งใหม่ไม่ได้จาก UI แล้ว)
type Category struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"not null"`
	// Description คือคำอธิบายสั้นๆ ของหมวดหมู่ โชว์ในหน้าจัดการเมนู ไม่บังคับกรอก
	Description string `json:"description"`
	// Color คือสีที่ใช้แสดงจุดสีหน้ารายการหมวดหมู่ในหน้าจัดการเมนู เก็บเป็น hex เช่น "#4f46e5"
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	// IsEnabled คือเปิด/ปิดใช้งานหมวดหมู่นี้ในหน้าขาย (POS) โดยไม่ต้องลบทิ้ง เช่น หมวดที่ยังไม่พร้อมขาย
	IsEnabled bool `json:"is_enabled" gorm:"not null;default:true"`
	// IsArchived คือเก็บหมวดหมู่นี้เข้าคลังเก็บถาวร (ซ่อนจากหน้าจัดการเมนูปกติ แต่กู้คืนได้ ต่างจากลบถาวร)
	IsArchived bool   `json:"is_archived" gorm:"not null;default:false"`
	Station    string `json:"station"`
}

// CategoryOptionTemplate คือ "ค่าเริ่มต้น" ของกลุ่มตัวเลือกที่ตั้งไว้ระดับหมวดหมู่ เช่น
// หมวด "กาแฟ" ตั้ง template "ความหวาน" ไว้ครั้งเดียว แล้วนำไปใช้ซ้ำกับเมนูกาแฟหลายๆ อย่างได้
// (เวลานำไปใช้จริงกับเมนู จะ copy ออกมาเป็น MenuOptionGroup ของเมนูนั้นๆ ไม่ได้ผูก reference ตรงๆ
// เพื่อให้แก้ template ภายหลังไม่กระทบเมนูที่เคย apply ไปแล้ว)
// SelectionType/MinSelect/MaxSelect/IsEnabled มีความหมายเดียวกับฟิลด์ชื่อเดียวกันใน MenuOptionGroup
// (ดูคอมเมนต์ที่นั่น) — ทำให้ฟอร์มเพิ่ม/แก้ไข template หน้า "ตัวเลือกเสริม" ใช้ฟิลด์ชุดเดียวกับฟอร์ม
// "เพิ่มกลุ่มตัวเลือก" ของเมนูได้เป๊ะๆ
type CategoryOptionTemplate struct {
	ID            uint                           `json:"id" gorm:"primaryKey"`
	CategoryID    uint                           `json:"category_id"`
	Name          string                         `json:"name" gorm:"not null"`
	Description   string                         `json:"description"`
	SelectionType string                         `json:"selection_type" gorm:"not null;default:single"`
	MinSelect     int                            `json:"min_select"`
	MaxSelect     int                            `json:"max_select" gorm:"default:1"`
	IsRequired    bool                           `json:"is_required" gorm:"default:true"`
	IsEnabled     bool                           `json:"is_enabled" gorm:"not null;default:true"`
	SortOrder     int                            `json:"sort_order"`
	Choices       []CategoryOptionTemplateChoice `json:"choices,omitempty" gorm:"foreignKey:TemplateID"`
}

// CategoryOptionTemplateChoice คือตัวเลือกย่อยของ template ระดับหมวดหมู่
type CategoryOptionTemplateChoice struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	TemplateID uint    `json:"template_id"`
	Name       string  `json:"name" gorm:"not null"`
	PriceDelta float64 `json:"price_delta"`
	SortOrder  int     `json:"sort_order"`
	IsDefault  bool    `json:"is_default" gorm:"not null;default:false"`
	IsEnabled  bool    `json:"is_enabled" gorm:"not null;default:true"`
}

// MenuItem คือรายการอาหาร/เครื่องดื่มแต่ละอย่าง
type MenuItem struct {
	ID          uint     `json:"id" gorm:"primaryKey"`
	CategoryID  uint     `json:"category_id"`
	Category    Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Name        string   `json:"name" gorm:"not null"`
	Price       float64  `json:"price" gorm:"not null"`
	IsAvailable bool     `json:"is_available" gorm:"default:true"`
	// ImagePath คือ path ของรูปเมนูที่อัปโหลดไว้ (เช่น "/uploads/menu-items/xxx.jpg") ว่างได้ถ้ายังไม่มีรูป
	ImagePath string `json:"image_path"`
	// ป้ายสถานะที่โชว์ในหน้าจัดการเมนู/หน้าขาย — เก็บไว้เป็น flag เฉยๆ ยังไม่มี logic พิเศษผูกอยู่
	// (เช่น TrackStock ยังไม่มีระบบคลังสินค้า/หักสต็อกจริงรองรับ เก็บไว้ก่อนเผื่ออนาคต)
	IsFeatured   bool              `json:"is_featured" gorm:"not null;default:false"`   // เมนูแนะนำ
	IsBestseller bool              `json:"is_bestseller" gorm:"not null;default:false"` // ขายดี
	TrackStock   bool              `json:"track_stock" gorm:"not null;default:false"`   // ติดตามสต็อก
	IsArchived   bool              `json:"is_archived" gorm:"not null;default:false"`
	OptionGroups []MenuOptionGroup `json:"option_groups,omitempty" gorm:"foreignKey:MenuItemID"`
	CreatedAt    time.Time         `json:"created_at"`
}

// MenuOptionGroup คือกลุ่มตัวเลือกของเมนู เช่น "ความหวาน", "ระดับน้ำแข็ง" ตั้งค่าได้ตอนสร้าง/แก้ไขเมนู
// SelectionType คือ "single" (เลือกได้อย่างเดียว) หรือ "multi" (เลือกได้หลายอย่าง)
// MinSelect/MaxSelect คือจำนวนตัวเลือกย่อยขั้นต่ำ/สูงสุดที่เลือกได้ในกลุ่มนี้ (ใช้ตอน validate ตอนสั่ง)
type MenuOptionGroup struct {
	ID            uint               `json:"id" gorm:"primaryKey"`
	MenuItemID    uint               `json:"menu_item_id"`
	Name          string             `json:"name" gorm:"not null"`
	Description   string             `json:"description"`
	SelectionType string             `json:"selection_type" gorm:"not null;default:single"` // single | multi
	MinSelect     int                `json:"min_select"`
	MaxSelect     int                `json:"max_select" gorm:"default:1"`
	IsRequired    bool               `json:"is_required" gorm:"default:true"` // ถ้า true ลูกค้าต้องเลือกอย่างน้อย 1 ตัวเลือกในกลุ่มนี้ก่อนสั่งได้
	IsEnabled     bool               `json:"is_enabled" gorm:"not null;default:true"`
	SortOrder     int                `json:"sort_order"`
	Choices       []MenuOptionChoice `json:"choices,omitempty" gorm:"foreignKey:OptionGroupID"`
}

// MenuOptionChoice คือตัวเลือกย่อยในกลุ่ม เช่น "หวานน้อย", "หวานปกติ", "หวานมาก"
// IsDefault คือถูกเลือกไว้ล่วงหน้าให้อัตโนมัติตอนเปิดกล่องเลือกตัวเลือก (แค่ค่าเริ่มต้น ลูกค้าเปลี่ยนได้)
// IsEnabled คือปิดการใช้งานตัวเลือกย่อยนี้ชั่วคราวได้โดยไม่ต้องลบทิ้ง (เช่น วัตถุดิบหมดชั่วคราว)
type MenuOptionChoice struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	OptionGroupID uint    `json:"option_group_id"`
	Name          string  `json:"name" gorm:"not null"`
	PriceDelta    float64 `json:"price_delta"` // ราคาที่บวกเพิ่มจากตัวเลือกนี้ (ถ้ามี เช่น ไซส์ใหญ่ +10)
	SortOrder     int     `json:"sort_order"`
	IsDefault     bool    `json:"is_default" gorm:"not null;default:false"`
	IsEnabled     bool    `json:"is_enabled" gorm:"not null;default:true"`
}

// DiningTable คือโต๊ะสำหรับลูกค้านั่งทาน
// Capacity คือจำนวนคนที่นั่งได้ต่อโต๊ะ ใช้เตือน (ไม่บังคับ) ตอนสร้างออเดอร์ถ้าจำนวนคนเกินที่นั่ง
type DiningTable struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"not null"`
	Zone     string `json:"zone"`
	Status   string `json:"status" gorm:"default:available"` // available | occupied | reserved
	Capacity int    `json:"capacity"`
}

// Reservation คือการจองโต๊ะไว้ล่วงหน้า หรือกันโต๊ะชั่วคราวตอนลูกค้ามาถึงร้านแล้วแต่ยังไม่เปิดบิล
// ReservedFor เป็น null หมายถึง "กันไว้ตอนนี้เลย" (ลูกค้ามาถึงร้านแล้ว รอโต๊ะสักครู่)
// ถ้ามีค่า หมายถึงลูกค้าโทร/จองโต๊ะไว้ล่วงหน้าสำหรับเวลานั้น
// พอเปิดบิลจริงที่โต๊ะนี้ (CreateOrder) จะถูกปรับสถานะเป็น seated + ผูก OrderID ให้อัตโนมัติ
type Reservation struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	TableID       uint         `json:"table_id" gorm:"not null"`
	Table         *DiningTable `json:"table,omitempty" gorm:"foreignKey:TableID"`
	CustomerName  string       `json:"customer_name" gorm:"not null"`
	CustomerPhone string       `json:"customer_phone"`
	PartySize     int          `json:"party_size"`
	ReservedFor   *time.Time   `json:"reserved_for"`
	Note          string       `json:"note"`
	Status        string       `json:"status" gorm:"not null;default:active"` // active | seated | cancelled | no_show
	OrderID       *uint        `json:"order_id"`
	CreatedBy     uint         `json:"created_by"`
	// ใครเป็นคนกดกัน/จองโต๊ะนี้ไว้ (แนบมาเฉพาะตอน preload ให้)
	CreatedByUser *User     `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	CreatedAt     time.Time `json:"created_at"`
}

// ShopSettings คือข้อมูลร้านค้า (ชื่อ/ที่อยู่/เบอร์โทร/เลขผู้เสียภาษี) เก็บเป็นแถวเดียว (id=1 เสมอ)
// ใช้โชว์บนหัวใบเสร็จตอนพิมพ์บิล แก้ไขได้เฉพาะ admin ที่หน้า "ตั้งค่าร้านค้า"
type ShopSettings struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	TaxID   string `json:"tax_id"`
	// PromptPayID คือเลขที่ใช้รับเงินผ่านพร้อมเพย์ของร้าน กรอกได้ 2 แบบ: เบอร์โทร 10 หลัก (เช่น 0812345678)
	// หรือเลขบัตรประชาชน/เลขผู้เสียภาษี 13 หลัก — ใช้สร้าง QR โค้ดพร้อมเพย์แบบระบุจำนวนเงินตายตัวที่หน้าคิดเงิน
	// (ดู promptpay.ts ฝั่ง frontend ที่ประกอบ payload ตามสเปก EMVCo Thai QR Payment) ไม่บังคับกรอก ถ้าว่างจะไม่
	// แสดง QR ให้เลือก
	PromptPayID string `json:"promptpay_id"`
	// PromptPayName คือชื่อผู้รับเงินที่จะฝังไปใน QR (ไม่บังคับ ถ้าไม่กรอกจะ fallback ไปใช้ชื่อร้าน Name แทน)
	PromptPayName string `json:"promptpay_name"`
}

// Zone คือโซนของโต๊ะ (เช่น ในร้าน, ริมหน้าต่าง) แยกจัดการเปิด/ปิดใช้งานได้ต่างหาก
// (เช่น ปิดซ่อม หรือมีการจองที่นั่งไว้ทั้งโซน) ผูกกับ DiningTable.Zone แบบชื่อ (string) ไม่ใช่ foreign key
// เพราะโต๊ะเดิมเก็บชื่อโซนเป็น string มาก่อนฟีเจอร์นี้อยู่แล้ว
type Zone struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"not null;uniqueIndex"`
	IsActive bool   `json:"is_active" gorm:"not null;default:true"`
}

// Order คือบิล/ออเดอร์หนึ่งใบ (นั่งทานหรือซื้อกลับ)
type Order struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	OrderType  string       `json:"order_type" gorm:"not null"` // dine_in | takeaway
	TableID    *uint        `json:"table_id"`
	Table      *DiningTable `json:"table,omitempty" gorm:"foreignKey:TableID"`
	GuestCount int          `json:"guest_count"` // จำนวนคนที่มา (ถ้าระบุ) แก้ไขทีหลังได้ที่หน้าออเดอร์
	// คำสั่งพิเศษ/โน้ตระดับทั้งบิล (ต่างจาก OrderItem.Note ที่เป็นโน้ตต่อรายการ) กรอกได้ตั้งแต่หน้าขายก่อนส่งออเดอร์
	Note           string  `json:"note"`
	Status         string  `json:"status" gorm:"not null;default:open"`        // open|preparing|served|paid|cancelled
	Subtotal       float64 `json:"subtotal"`                                   // ยอดรวมก่อนหักส่วนลด
	DiscountType   string  `json:"discount_type" gorm:"not null;default:none"` // none | amount | percent
	DiscountValue  float64 `json:"discount_value"`                             // ค่าที่กรอก (บาท หรือ %)
	DiscountAmount float64 `json:"discount_amount"`                            // ยอดส่วนลดที่คำนวณเป็นบาท
	TotalAmount    float64 `json:"total_amount"`                               // ยอดสุทธิที่ต้องจ่าย = Subtotal - DiscountAmount
	CreatedBy      uint    `json:"created_by"`
	// ใครเป็นคนเปิดบิลนี้ (แนบมาเฉพาะตอน preload ให้ เพื่อให้ admin เช็คย้อนหลังได้ที่หน้าออเดอร์/รายงาน)
	CreatedByUser *User       `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	Items         []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	Payment       *Payment    `json:"payment,omitempty" gorm:"foreignKey:OrderID"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// OrderItem คือรายการสินค้าแต่ละบรรทัดในบิล
type OrderItem struct {
	ID         uint     `json:"id" gorm:"primaryKey"`
	OrderID    uint     `json:"order_id"`
	MenuItemID uint     `json:"menu_item_id"`
	MenuItem   MenuItem `json:"menu_item,omitempty" gorm:"foreignKey:MenuItemID"`
	Quantity   int      `json:"quantity" gorm:"not null"`
	UnitPrice  float64  `json:"unit_price" gorm:"not null"`
	Note       string   `json:"note"`
	Status     string   `json:"status" gorm:"default:pending"` // pending|preparing|served
	// IsTakeaway ใช้กรณีลูกค้านั่งทานที่โต๊ะ (order_type=dine_in) แต่อยากสั่งบางรายการกลับบ้านด้วย
	// อยู่ในบิลเดียวกัน จ่ายเงินครั้งเดียว แต่แยกแท็กไว้ให้ครัว/แคชเชียร์รู้ว่าต้องแพ็คกลับบ้าน
	IsTakeaway bool              `json:"is_takeaway" gorm:"default:false"`
	Options    []OrderItemOption `json:"options,omitempty" gorm:"foreignKey:OrderItemID"`
}

// OrderItemOption คือตัวเลือกที่ถูกเลือกไว้ในรายการนั้น ๆ ตอนสั่ง
// เก็บชื่อ/ราคาแบบ snapshot ไว้ เพื่อไม่ให้บิลเก่าเปลี่ยนไปถ้าแก้เมนู/ตัวเลือกทีหลัง
type OrderItemOption struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	OrderItemID   uint    `json:"order_item_id"`
	OptionGroupID uint    `json:"option_group_id"`
	GroupName     string  `json:"group_name"`
	ChoiceID      uint    `json:"choice_id"`
	ChoiceName    string  `json:"choice_name"`
	PriceDelta    float64 `json:"price_delta"`
}

// Payment คือการปิดบิล/รับเงินของออเดอร์นั้น
// Method เป็น cash (เงินสด) หรือ transfer (โอนเงิน/พร้อมเพย์) ถ้าเป็น transfer จะไม่คำนวณเงินทอน
// (ถือว่าลูกค้าโอนมาพอดียอด) แต่เก็บ TransferRef (เลขอ้างอิงการโอน) แทน และแนบสลิปเป็นรูปได้ทีหลังก็ได้
// (ไม่บังคับตอนปิดบิล เผื่อพนักงานยุ่งจนไม่ได้ถ่ายสลิปตอนนั้น มาแนบย้อนหลังที่หน้ารายงานได้)
type Payment struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	OrderID        uint    `json:"order_id" gorm:"uniqueIndex"`
	Method         string  `json:"method" gorm:"default:cash"` // cash | transfer
	Subtotal       float64 `json:"subtotal"`                   // เก็บยอดรวมก่อนหักส่วนลด ณ ตอนปิดบิล (ไว้ทำใบเสร็จ/รายงานย้อนหลัง)
	DiscountAmount float64 `json:"discount_amount"`            // เก็บยอดส่วนลด ณ ตอนปิดบิล
	Amount         float64 `json:"amount"`
	ReceivedAmount float64 `json:"received_amount"`
	ChangeAmount   float64 `json:"change_amount"`
	// TransferRef คือเลขอ้างอิงการโอน (เช่น เลขท้ายอ้างอิงจากแอปธนาคาร) ใช้เฉพาะ method=transfer
	TransferRef string `json:"transfer_ref"`
	// SlipImagePath คือ path ของรูปสลิปที่แนบไว้ (เช่น "/uploads/slips/xxx.jpg") ว่างได้ถ้ายังไม่ได้แนบ
	SlipImagePath string `json:"slip_image_path"`
	PaidBy        uint   `json:"paid_by"`
	// ใครเป็นคนกดปิดบิล/รับเงิน (แนบมาเฉพาะตอน preload ให้)
	PaidByUser *User     `json:"paid_by_user,omitempty" gorm:"foreignKey:PaidBy;references:ID"`
	PaidAt     time.Time `json:"paid_at"`
	// InvoiceNo คือเลขที่ใบเสร็จ/ใบกำกับภาษีอย่างย่อ รูปแบบ INV-YYYYMMDD-00001 (รันต่อวัน ขึ้นวันใหม่เริ่ม 00001 ใหม่)
	// สร้างอัตโนมัติครั้งเดียวตอนปิดบิลสำเร็จ (ดู generateInvoiceNo ใน handlers/invoice.go) เก็บ persist ไว้ใน DB จริง
	// เพื่อให้เลขที่เรียงต่อเนื่องไม่ซ้ำ/ไม่กระโดด เอาไว้ใช้อ้างอิงทำบัญชี/ภาษี ต่างจาก order.id ที่โชว์บนใบเสร็จเดิม
	// ซึ่งเป็นแค่เลขอ้างอิงออเดอร์ ไม่ใช่เลขที่บิลอย่างเป็นทางการ
	InvoiceNo string `json:"invoice_no" gorm:"uniqueIndex"`
}

// InvoiceCounter คือตัวนับเลขที่ใบเสร็จ/ใบกำกับภาษีแบบรันต่อวัน มีแถวเดียวต่อวัน (DateKey รูปแบบ YYYYMMDD)
// ใช้สร้างเลขที่บิล Payment.InvoiceNo แบบ persist จริงใน DB (บวกทีละ 1 ในทรานแซกชันเดียวกับการสร้าง Payment
// เพื่อกันเลขซ้ำถ้ามีการปิดบิลพร้อมกันหลายบิล) ต่างจากเดิมที่ frontend คำนวณเลขที่บิลเองจาก order.id เฉยๆ
// ซึ่งไม่ persist และไม่การันตีว่าเรียงต่อเนื่อง
type InvoiceCounter struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	DateKey    string `json:"date_key" gorm:"uniqueIndex;not null"`
	LastNumber int    `json:"last_number" gorm:"not null;default:0"`
}

// Member คือลูกค้าที่ลงทะเบียนสมาชิกสะสมแต้ม ผูกด้วยเบอร์โทร (ไม่ซ้ำกัน ใช้ค้นหาที่หน้าขายได้)
// PointsBalance คือแต้มคงเหลือปัจจุบัน ปรับได้ผ่าน AdjustPoints เท่านั้น (ทุกครั้งที่ปรับจะบันทึกลง
// MemberPointHistory ไว้ตรวจสอบย้อนหลังได้ว่าแต้มเปลี่ยนเพราะอะไร) Tier คำนวณจาก TotalSpent เทียบกับ
// LoyaltySettings.TierRules — สมาชิกใหม่เริ่มที่ bronze/แต้ม 0/ยอดใช้จ่าย 0 เสมอ
type Member struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	Name          string  `json:"name" gorm:"not null"`
	Phone         string  `json:"phone" gorm:"not null;uniqueIndex"`
	PointsBalance int     `json:"points_balance" gorm:"not null;default:0"`
	Tier          string  `json:"tier" gorm:"not null;default:bronze"` // bronze | silver | gold | platinum
	TotalSpent    float64 `json:"total_spent" gorm:"not null;default:0"`
	// IsActive คือเปิด/ปิดใช้งานสมาชิกภาพนี้ได้โดยไม่ต้องลบทิ้ง (เช่น ลูกค้าขอยกเลิกสมาชิก) แต่ยังเก็บ
	// ประวัติแต้มเดิมไว้ตรวจสอบย้อนหลังได้
	IsActive  bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemberPointHistory คือประวัติการเปลี่ยนแปลงแต้มของสมาชิกแต่ละครั้ง (ทั้งบวก/ลบ)
// OrderID เป็น null ได้ ถ้าแต้มถูกปรับโดยตรงจากหน้าจัดการสมาชิก ไม่ได้ผูกกับบิลใด (เช่น แอดมินปรับแต้มเอง)
type MemberPointHistory struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	MemberID  uint      `json:"member_id" gorm:"not null;index"`
	OrderID   *uint     `json:"order_id"`
	Change    int       `json:"change" gorm:"not null"` // ค่าที่เปลี่ยนแปลง (บวก = ได้แต้มเพิ่ม, ลบ = แลก/หักแต้ม)
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// PointAccumulationRule คือกฎการสะสมแต้มจากยอดซื้อ เก็บเป็น JSON ในคอลัมน์เดียวของ LoyaltySettings
// (ดู Value/Scan ด้านล่าง) เพราะเป็นค่าตั้งค่าที่แก้ทั้งชุดพร้อมกันเสมอ ไม่มีที่ไหน query แยกเป็นรายฟิลด์
type PointAccumulationRule struct {
	SpendPerPoint    float64 `json:"spend_per_point"`    // ใช้จ่ายกี่บาทถึงได้ 1 แต้ม
	PointsExpiryDays int     `json:"points_expiry_days"` // แต้มหมดอายุกี่วันหลังได้รับ (0 = ไม่หมดอายุ)
	MinSpendToEarn   float64 `json:"min_spend_to_earn"`  // ยอดซื้อขั้นต่ำต่อบิลถึงจะเริ่มได้แต้ม
}

func (r PointAccumulationRule) Value() (driver.Value, error) {
	b, err := json.Marshal(r)
	return string(b), err
}

func (r *PointAccumulationRule) Scan(value interface{}) error {
	return scanJSONColumn(value, r)
}

// RedemptionRule คือกฎการแลกแต้มเป็นส่วนลด เก็บเป็น JSON เช่นเดียวกับ PointAccumulationRule
type RedemptionRule struct {
	PointsPerBaht     float64 `json:"points_per_baht"`      // กี่แต้มแลกได้ส่วนลด 1 บาท
	MinPointsToRedeem int     `json:"min_points_to_redeem"` // ต้องมีแต้มขั้นต่ำเท่าไหร่ถึงแลกได้
	MaxDiscountRatio  float64 `json:"max_discount_ratio"`   // ส่วนลดจากแต้มแลกได้สูงสุดกี่เท่าของยอดบิล (0-1)
}

func (r RedemptionRule) Value() (driver.Value, error) {
	b, err := json.Marshal(r)
	return string(b), err
}

func (r *RedemptionRule) Scan(value interface{}) error {
	return scanJSONColumn(value, r)
}

// TierRule คือเกณฑ์ของสมาชิกระดับหนึ่ง (bronze/silver/gold/platinum) — ยอดใช้จ่ายสะสมขั้นต่ำที่ต้องถึง
// และตัวคูณแต้มที่ได้รับสำหรับระดับนั้น (ระดับสูงขึ้นมักได้แต้มต่อบาทมากขึ้นเป็นสิทธิพิเศษ)
type TierRule struct {
	Tier             string  `json:"tier"`
	Label            string  `json:"label"`
	MinTotalSpent    float64 `json:"min_total_spent"`
	PointsMultiplier float64 `json:"points_multiplier"`
}

// TierRules คือรายการ TierRule ทั้งหมด เก็บเป็น JSON ในคอลัมน์เดียวเช่นเดียวกับกฎอื่นๆ ข้างบน
type TierRules []TierRule

func (r TierRules) Value() (driver.Value, error) {
	b, err := json.Marshal(r)
	return string(b), err
}

func (r *TierRules) Scan(value interface{}) error {
	return scanJSONColumn(value, r)
}

// scanJSONColumn คือ helper ใช้ร่วมกันสำหรับ Scan() ของ PointAccumulationRule/RedemptionRule/TierRules
// แปลงค่าที่ได้จาก driver (SQLite คืนมาเป็น []byte หรือ string ก็ได้ขึ้นกับ query) กลับเป็น struct/slice เดิม
func scanJSONColumn(value interface{}, dest interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, dest)
	case string:
		return json.Unmarshal([]byte(v), dest)
	default:
		return fmt.Errorf("unsupported type for JSON column: %T", value)
	}
}

// LoyaltySettings คือค่าตั้งค่าระบบสะสมแต้มทั้งร้าน เก็บเป็นแถวเดียว (id=1 เสมอ) เหมือน ShopSettings
// IsEnabled คือเปิด/ปิดใช้งานระบบสมาชิกทั้งร้าน แก้ไขได้เฉพาะ admin ที่หน้า "ตั้งค่า" ของหน้าสมาชิก
type LoyaltySettings struct {
	ID           uint                  `json:"id" gorm:"primaryKey"`
	IsEnabled    bool                  `json:"is_enabled" gorm:"not null;default:true"`
	Accumulation PointAccumulationRule `json:"accumulation" gorm:"type:text"`
	Redemption   RedemptionRule        `json:"redemption" gorm:"type:text"`
	TierRules    TierRules             `json:"tier_rules" gorm:"type:text"`
	UpdatedAt    time.Time             `json:"updated_at"`
}
