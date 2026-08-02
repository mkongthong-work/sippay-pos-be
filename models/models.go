package models

import "time"

// User คือพนักงาน/แอดมินที่ล็อกอินเข้าระบบ
type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Name         string    `json:"name"`
	Role         string    `json:"role" gorm:"not null;default:staff"` // admin | staff
	CreatedAt    time.Time `json:"created_at"`
}

// Category คือหมวดหมู่เมนู เช่น กาแฟ, อาหาร, ของหวาน
// Station คือสถานที่ทำของหมวดนี้ เช่น "ครัว"/"บาร์"/"อื่นๆ" หรือชื่อที่แอดมินพิมพ์เพิ่มเอง
// เก็บไว้เผื่ออนาคตจะแยกจอครัว/บาร์ตาม station นี้ ตอนนี้ยังไม่มีหน้าไหนใช้กรองจริง
type Category struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"not null"`
	SortOrder int    `json:"sort_order"`
	Station   string `json:"station"`
}

// CategoryOptionTemplate คือ "ค่าเริ่มต้น" ของกลุ่มตัวเลือกที่ตั้งไว้ระดับหมวดหมู่ เช่น
// หมวด "กาแฟ" ตั้ง template "ความหวาน" ไว้ครั้งเดียว แล้วนำไปใช้ซ้ำกับเมนูกาแฟหลายๆ อย่างได้
// (เวลานำไปใช้จริงกับเมนู จะ copy ออกมาเป็น MenuOptionGroup ของเมนูนั้นๆ ไม่ได้ผูก reference ตรงๆ
// เพื่อให้แก้ template ภายหลังไม่กระทบเมนูที่เคย apply ไปแล้ว)
type CategoryOptionTemplate struct {
	ID         uint                       `json:"id" gorm:"primaryKey"`
	CategoryID uint                       `json:"category_id"`
	Name       string                     `json:"name" gorm:"not null"`
	IsRequired bool                       `json:"is_required" gorm:"default:true"`
	SortOrder  int                        `json:"sort_order"`
	Choices    []CategoryOptionTemplateChoice `json:"choices,omitempty" gorm:"foreignKey:TemplateID"`
}

// CategoryOptionTemplateChoice คือตัวเลือกย่อยของ template ระดับหมวดหมู่
type CategoryOptionTemplateChoice struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	TemplateID uint    `json:"template_id"`
	Name       string  `json:"name" gorm:"not null"`
	PriceDelta float64 `json:"price_delta"`
	SortOrder  int     `json:"sort_order"`
}

// MenuItem คือรายการอาหาร/เครื่องดื่มแต่ละอย่าง
type MenuItem struct {
	ID           uint              `json:"id" gorm:"primaryKey"`
	CategoryID   uint              `json:"category_id"`
	Category     Category          `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Name         string            `json:"name" gorm:"not null"`
	Price        float64           `json:"price" gorm:"not null"`
	IsAvailable  bool              `json:"is_available" gorm:"default:true"`
	OptionGroups []MenuOptionGroup `json:"option_groups,omitempty" gorm:"foreignKey:MenuItemID"`
	CreatedAt    time.Time         `json:"created_at"`
}

// MenuOptionGroup คือกลุ่มตัวเลือกของเมนู เช่น "ความหวาน", "ระดับน้ำแข็ง" ตั้งค่าได้ตอนสร้าง/แก้ไขเมนู
type MenuOptionGroup struct {
	ID         uint               `json:"id" gorm:"primaryKey"`
	MenuItemID uint               `json:"menu_item_id"`
	Name       string             `json:"name" gorm:"not null"`
	IsRequired bool               `json:"is_required" gorm:"default:true"` // ถ้า true ลูกค้าต้องเลือก 1 ตัวเลือกในกลุ่มนี้ก่อนสั่งได้
	SortOrder  int                `json:"sort_order"`
	Choices    []MenuOptionChoice `json:"choices,omitempty" gorm:"foreignKey:OptionGroupID"`
}

// MenuOptionChoice คือตัวเลือกย่อยในกลุ่ม เช่น "หวานน้อย", "หวานปกติ", "หวานมาก"
type MenuOptionChoice struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	OptionGroupID uint    `json:"option_group_id"`
	Name          string  `json:"name" gorm:"not null"`
	PriceDelta    float64 `json:"price_delta"` // ราคาที่บวกเพิ่มจากตัวเลือกนี้ (ถ้ามี เช่น ไซส์ใหญ่ +10)
	SortOrder     int     `json:"sort_order"`
}

// DiningTable คือโต๊ะสำหรับลูกค้านั่งทาน
// Capacity คือจำนวนคนที่นั่งได้ต่อโต๊ะ ใช้เตือน (ไม่บังคับ) ตอนสร้างออเดอร์ถ้าจำนวนคนเกินที่นั่ง
type DiningTable struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"not null"`
	Zone     string `json:"zone"`
	Status   string `json:"status" gorm:"default:available"` // available | occupied
	Capacity int    `json:"capacity"`
}

// Order คือบิล/ออเดอร์หนึ่งใบ (นั่งทานหรือซื้อกลับ)
type Order struct {
	ID             uint         `json:"id" gorm:"primaryKey"`
	OrderType      string       `json:"order_type" gorm:"not null"` // dine_in | takeaway
	TableID        *uint        `json:"table_id"`
	Table          *DiningTable `json:"table,omitempty" gorm:"foreignKey:TableID"`
	GuestCount     int          `json:"guest_count"`                         // จำนวนคนที่มา (ถ้าระบุ) แก้ไขทีหลังได้ที่หน้าออเดอร์
	Status         string       `json:"status" gorm:"not null;default:open"` // open|preparing|served|paid|cancelled
	Subtotal       float64      `json:"subtotal"`                                 // ยอดรวมก่อนหักส่วนลด
	DiscountType   string       `json:"discount_type" gorm:"not null;default:none"` // none | amount | percent
	DiscountValue  float64      `json:"discount_value"`                           // ค่าที่กรอก (บาท หรือ %)
	DiscountAmount float64      `json:"discount_amount"`                          // ยอดส่วนลดที่คำนวณเป็นบาท
	TotalAmount    float64      `json:"total_amount"`                             // ยอดสุทธิที่ต้องจ่าย = Subtotal - DiscountAmount
	CreatedBy      uint         `json:"created_by"`
	Items          []OrderItem  `json:"items,omitempty" gorm:"foreignKey:OrderID"`
	Payment        *Payment     `json:"payment,omitempty" gorm:"foreignKey:OrderID"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// OrderItem คือรายการสินค้าแต่ละบรรทัดในบิล
type OrderItem struct {
	ID         uint              `json:"id" gorm:"primaryKey"`
	OrderID    uint              `json:"order_id"`
	MenuItemID uint              `json:"menu_item_id"`
	MenuItem   MenuItem          `json:"menu_item,omitempty" gorm:"foreignKey:MenuItemID"`
	Quantity   int               `json:"quantity" gorm:"not null"`
	UnitPrice  float64           `json:"unit_price" gorm:"not null"`
	Note       string            `json:"note"`
	Status     string            `json:"status" gorm:"default:pending"` // pending|preparing|served
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
type Payment struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	OrderID        uint      `json:"order_id" gorm:"uniqueIndex"`
	Method         string    `json:"method" gorm:"default:cash"`
	Subtotal       float64   `json:"subtotal"`        // เก็บยอดรวมก่อนหักส่วนลด ณ ตอนปิดบิล (ไว้ทำใบเสร็จ/รายงานย้อนหลัง)
	DiscountAmount float64   `json:"discount_amount"` // เก็บยอดส่วนลด ณ ตอนปิดบิล
	Amount         float64   `json:"amount"`
	ReceivedAmount float64   `json:"received_amount"`
	ChangeAmount   float64   `json:"change_amount"`
	PaidBy         uint      `json:"paid_by"`
	PaidAt         time.Time `json:"paid_at"`
}
