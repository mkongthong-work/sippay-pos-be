package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pos-backend/db"
	"pos-backend/models"
)

type orderItemInput struct {
	MenuItemID      uint   `json:"menu_item_id" binding:"required"`
	Quantity        int    `json:"quantity" binding:"required,min=1"`
	Note            string `json:"note"`
	OptionChoiceIDs []uint `json:"option_choice_ids"`
	IsTakeaway      bool   `json:"is_takeaway"`
}

type createOrderRequest struct {
	OrderType  string           `json:"order_type" binding:"required,oneof=dine_in takeaway"`
	TableID    *uint            `json:"table_id"`
	GuestCount int              `json:"guest_count"`
	Note       string           `json:"note"`
	Items      []orderItemInput `json:"items" binding:"required,min=1"`
}

// calcDiscountAmount คำนวณยอดส่วนลดเป็นบาทจากยอดรวมก่อนหักส่วนลด
// ครอบไว้ไม่ให้ต่ำกว่า 0 หรือมากกว่ายอดรวม (ลดเกินไม่ได้)
func calcDiscountAmount(subtotal float64, discountType string, discountValue float64) float64 {
	var amount float64
	switch discountType {
	case "percent":
		amount = subtotal * discountValue / 100
	case "amount":
		amount = discountValue
	default:
		amount = 0
	}
	if amount < 0 {
		amount = 0
	}
	if amount > subtotal {
		amount = subtotal
	}
	return amount
}

// buildOrderItem สร้างแถวรายการสินค้า 1 บรรทัด พร้อมตัวเลือกที่แนบมา (option_choice_ids)
// คำนวณราคาต่อหน่วยรวมส่วนเพิ่มจากตัวเลือก และตรวจว่ากลุ่มตัวเลือกที่บังคับเลือกถูกเลือกครบหรือยัง
func buildOrderItem(tx *gorm.DB, orderID uint, in orderItemInput) (models.OrderItem, error) {
	var menuItem models.MenuItem
	if err := tx.Preload("OptionGroups.Choices").First(&menuItem, in.MenuItemID).Error; err != nil {
		return models.OrderItem{}, errors.New("menu item not found")
	}

	// จับคู่ choice ที่เลือกมากับกลุ่มของมัน — กลุ่มหนึ่งเลือกได้หลายตัวเลือกพร้อมกันได้ถ้าเป็นกลุ่มแบบ
	// multi-select (MaxSelect > 1) จึงเก็บเป็น slice ต่อกลุ่ม ไม่ใช่ค่าเดียว (เดิมเก็บแบบค่าเดียวทำให้
	// ถ้าเลือกหลายตัวเลือกในกลุ่มเดียวกัน ตัวที่เลือกไว้จะถูกทับเหลือแค่ตัวสุดท้าย ราคา/ตัวเลือกที่บันทึกจึง
	// ขาดหายไปจากที่หน้าขายคำนวณไว้)
	chosenByGroup := map[uint][]models.MenuOptionChoice{}
	for _, group := range menuItem.OptionGroups {
		for _, choice := range group.Choices {
			for _, id := range in.OptionChoiceIDs {
				if choice.ID == id {
					chosenByGroup[group.ID] = append(chosenByGroup[group.ID], choice)
				}
			}
		}
	}

	for _, group := range menuItem.OptionGroups {
		if group.IsRequired {
			if len(chosenByGroup[group.ID]) == 0 {
				return models.OrderItem{}, fmt.Errorf("please select an option for %s", group.Name)
			}
		}
	}

	unitPrice := menuItem.Price
	for _, choices := range chosenByGroup {
		for _, choice := range choices {
			unitPrice += choice.PriceDelta
		}
	}

	orderItem := models.OrderItem{
		OrderID:    orderID,
		MenuItemID: in.MenuItemID,
		Quantity:   in.Quantity,
		UnitPrice:  unitPrice,
		Note:       in.Note,
		Status:     "pending",
		IsTakeaway: in.IsTakeaway,
	}
	if err := tx.Create(&orderItem).Error; err != nil {
		return models.OrderItem{}, err
	}

	for _, group := range menuItem.OptionGroups {
		choices, ok := chosenByGroup[group.ID]
		if !ok {
			continue
		}
		for _, choice := range choices {
			option := models.OrderItemOption{
				OrderItemID:   orderItem.ID,
				OptionGroupID: group.ID,
				GroupName:     group.Name,
				ChoiceID:      choice.ID,
				ChoiceName:    choice.Name,
				PriceDelta:    choice.PriceDelta,
			}
			tx.Create(&option)
		}
	}

	return orderItem, nil
}

// recalcTotal คำนวณยอดรวมก่อนหักส่วนลด ยอดส่วนลด และยอดสุทธิของออเดอร์ใหม่
// จากรายการทั้งหมดที่มีอยู่ + ส่วนลดที่ตั้งไว้บนออเดอร์นั้น (ถ้ามี)
func recalcTotal(orderID uint) {
	var order models.Order
	if err := db.DB.First(&order, orderID).Error; err != nil {
		return
	}

	var items []models.OrderItem
	db.DB.Where("order_id = ?", orderID).Find(&items)
	var subtotal float64
	for _, it := range items {
		subtotal += it.UnitPrice * float64(it.Quantity)
	}

	discountAmount := calcDiscountAmount(subtotal, order.DiscountType, order.DiscountValue)
	total := subtotal - discountAmount

	db.DB.Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
		"subtotal":        subtotal,
		"discount_amount": discountAmount,
		"total_amount":    total,
	})
}

func CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.OrderType == "dine_in" && req.TableID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table_id is required for dine_in orders"})
		return
	}

	userID, _ := c.Get("user_id")

	order := models.Order{
		OrderType:  req.OrderType,
		TableID:    req.TableID,
		GuestCount: req.GuestCount,
		Note:       req.Note,
		Status:     "open",
		CreatedBy:  userID.(uint),
	}

	tx := db.DB.Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, it := range req.Items {
		if _, err := buildOrderItem(tx, order.ID, it); err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if req.OrderType == "dine_in" && req.TableID != nil {
		tx.Model(&models.DiningTable{}).Where("id = ?", *req.TableID).Update("status", "occupied")
		// ถ้าโต๊ะนี้มีรายการจอง/กันโต๊ะที่ยัง active อยู่ ให้ปิดเป็น seated พร้อมผูก order นี้ให้อัตโนมัติ
		// (พนักงานแค่เลือกโต๊ะที่จองไว้แล้วเปิดบิลตามปกติ ไม่ต้องมากดปิด reservation เองอีกขั้นตอน)
		tx.Model(&models.Reservation{}).
			Where("table_id = ? AND status = ?", *req.TableID, "active").
			Updates(map[string]interface{}{"status": "seated", "order_id": order.ID})
	}

	tx.Commit()

	recalcTotal(order.ID)

	db.DB.Preload("Items.MenuItem").Preload("Items.Options").Preload("Table").Preload("CreatedByUser").First(&order, order.ID)
	c.JSON(http.StatusCreated, order)
}

func ListOrders(c *gin.Context) {
	var orders []models.Order
	query := db.DB.
		Preload("Items.MenuItem").
		Preload("Items.MenuItem.Category").
		Preload("Items.Options").
		Preload("Table").
		Preload("Payment").
		Preload("Payment.PaidByUser").
		Preload("CreatedByUser").
		Order("created_at desc")
	if statusParam := c.Query("status"); statusParam != "" {
		statuses := strings.Split(statusParam, ",")
		query = query.Where("status IN ?", statuses)
	}
	query.Find(&orders)

	// เช็คซ้ำทุกครั้งที่ดึงรายการ เผื่อบิลไหนสถานะค้างไม่ตรงกับรายชิ้น (เช่น บิลเก่าก่อนจะมี auto-sync)
	// จะได้ถูกต้องทันทีโดยไม่ต้องรอให้ครัวกดปุ่มรายชิ้นซ้ำอีกที
	for i := range orders {
		if newStatus, changed := autoSyncOrderStatusFromItems(orders[i].ID); changed {
			orders[i].Status = newStatus
		}
	}

	c.JSON(http.StatusOK, orders)
}

func GetOrder(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := db.DB.
		Preload("Items.MenuItem").
		Preload("Items.MenuItem.Category").
		Preload("Items.Options").
		Preload("Table").
		Preload("Payment").
		Preload("Payment.PaidByUser").
		Preload("CreatedByUser").
		First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if newStatus, changed := autoSyncOrderStatusFromItems(order.ID); changed {
		order.Status = newStatus
	}
	c.JSON(http.StatusOK, order)
}

func AddOrderItem(c *gin.Context) {
	orderID := c.Param("id")
	var order models.Order
	if err := db.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.Status == "paid" || order.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify a paid or cancelled order"})
		return
	}

	var input orderItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := buildOrderItem(db.DB, order.ID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recalcTotal(order.ID)

	db.DB.Preload("Items.MenuItem").Preload("Items.Options").Preload("Table").First(&order, order.ID)
	c.JSON(http.StatusCreated, order)
}

type updateOrderItemRequest struct {
	Quantity   *int    `json:"quantity"`
	Status     *string `json:"status" binding:"omitempty,oneof=pending preparing served"`
	IsTakeaway *bool   `json:"is_takeaway"`
}

func UpdateOrderItem(c *gin.Context) {
	orderID := c.Param("id")
	itemID := c.Param("itemId")

	var item models.OrderItem
	if err := db.DB.Where("id = ? AND order_id = ?", itemID, orderID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order item not found"})
		return
	}

	var input updateOrderItemRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// รายการที่ครัวติ๊กว่าเสิร์ฟแล้ว (served) ห้ามแก้จำนวนอีก เพราะทำเสร็จ/เสิร์ฟไปแล้วจริง
	// แต่ยังปรับ status กลับได้ (เผื่อครัวติ๊กผิดแล้วอยากติ๊กออก)
	if input.Quantity != nil && item.Status == "served" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change quantity of an item that's already served"})
		return
	}
	if input.Quantity != nil {
		item.Quantity = *input.Quantity
	}
	statusChanged := input.Status != nil
	if input.Status != nil {
		item.Status = *input.Status
	}
	if input.IsTakeaway != nil {
		item.IsTakeaway = *input.IsTakeaway
	}
	db.DB.Save(&item)
	recalcTotal(item.OrderID)

	// เมื่อครัวติ๊ก/ถอนสถานะรายชิ้น ให้เช็คและปรับสถานะบิลโดยรวมตามอัตโนมัติ
	// (ไม่ต้องให้พนักงานหน้าร้านกดปุ่มเลื่อนสถานะเองอีกต่อไปหลังจากเริ่มทำแล้ว)
	if statusChanged {
		autoSyncOrderStatusFromItems(item.OrderID)
	}

	c.JSON(http.StatusOK, item)
}

// autoSyncOrderStatusFromItems ปรับสถานะบิลโดยรวมให้สอดคล้องกับสถานะรายชิ้นจากครัวโดยอัตโนมัติ:
//   - ทุกรายการเสิร์ฟแล้ว (served) -> บิลเป็น "served"
//   - มีอย่างน้อย 1 รายการเริ่มทำแล้ว (preparing/served) แต่ยังไม่ครบ -> บิลเป็น "preparing"
//   - ทุกรายการยังไม่เริ่มทำ (pending) ทั้งหมด แต่บิลเคยขยับสถานะไปแล้ว -> ถอยกลับเป็น "open"
//     (เผื่อครัวกดพลาดแล้วย้อนกลับทุกรายการ)
//
// ไม่ยุ่งกับบิลที่จ่ายเงินแล้วหรือยกเลิกไปแล้ว
//
// คืนค่า (สถานะล่าสุดของบิล, มีการเปลี่ยนแปลงหรือไม่) ให้ผู้เรียกใช้อัปเดตค่าที่โหลดไว้ในหน่วยความจำ
// ต่อได้เลยโดยไม่ต้อง query ซ้ำ — ใช้ทั้งตอนบันทึก (UpdateOrderItem) และตอนอ่าน (ListOrders/GetOrder)
// เพื่อให้บิลเก่าที่สถานะค้างไม่ตรงกับรายชิ้น (เช่น บันทึกไว้ก่อนจะมีฟีเจอร์นี้) ถูกแก้ให้ถูกต้องทันทีที่อ่านครั้งถัดไป
func autoSyncOrderStatusFromItems(orderID uint) (string, bool) {
	var order models.Order
	if err := db.DB.First(&order, orderID).Error; err != nil {
		return "", false
	}
	if order.Status == "paid" || order.Status == "cancelled" {
		return order.Status, false
	}

	var items []models.OrderItem
	db.DB.Where("order_id = ?", orderID).Find(&items)
	if len(items) == 0 {
		return order.Status, false
	}

	allServed := true
	anyStarted := false
	for _, it := range items {
		if it.Status != "served" {
			allServed = false
		}
		if it.Status == "preparing" || it.Status == "served" {
			anyStarted = true
		}
	}

	newStatus := order.Status
	switch {
	case allServed:
		newStatus = "served"
	case anyStarted:
		newStatus = "preparing"
	case order.Status == "served" || order.Status == "preparing":
		newStatus = "open"
	}

	if newStatus != order.Status {
		db.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("status", newStatus)
		return newStatus, true
	}
	return order.Status, false
}

func DeleteOrderItem(c *gin.Context) {
	orderID := c.Param("id")
	itemID := c.Param("itemId")

	var item models.OrderItem
	if err := db.DB.Where("id = ? AND order_id = ?", itemID, orderID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order item not found"})
		return
	}
	// รายการที่ครัวติ๊กว่าเสิร์ฟแล้ว ห้ามยกเลิก/ลบอีก เพราะทำเสร็จและเสิร์ฟไปแล้วจริง
	if item.Status == "served" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot cancel an item that's already served"})
		return
	}

	if err := db.DB.Delete(&models.OrderItem{}, item.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if oid, err := strconv.Atoi(orderID); err == nil {
		recalcTotal(uint(oid))
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

type updateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=open preparing served paid cancelled"`
}

func UpdateOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	var req updateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order.Status = req.Status
	db.DB.Save(&order)

	// ปิดบิลไม่ว่าจะด้วยการจ่ายเงินหรือยกเลิก ต้องปล่อยโต๊ะกลับมาว่างเสมอ (จุดนี้เป็นทางแก้ไขสถานะแบบทั่วไป
	// แยกจาก PayOrder ที่ปล่อยโต๊ะให้อยู่แล้วตอนจ่ายเงินผ่านหน้าคิดเงินปกติ — เผื่อมีการเรียก endpoint นี้ตรงๆ
	// เพื่อตั้งสถานะเป็น paid ก็ต้องปล่อยโต๊ะเหมือนกัน ไม่งั้นโต๊ะจะค้างสถานะไม่ว่างทั้งที่บิลปิดไปแล้ว)
	if (req.Status == "cancelled" || req.Status == "paid") && order.TableID != nil {
		db.DB.Model(&models.DiningTable{}).Where("id = ?", *order.TableID).Update("status", "available")
	}

	c.JSON(http.StatusOK, order)
}

type updateGuestCountRequest struct {
	GuestCount int `json:"guest_count" binding:"required,min=1"`
}

// UpdateOrderGuestCount แก้ไขจำนวนคนที่มาของออเดอร์ (ตั้งตอนสร้างที่ POS ได้ หรือมาแก้ทีหลังที่หน้าออเดอร์ก็ได้)
func UpdateOrderGuestCount(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.Status == "paid" || order.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify a paid or cancelled order"})
		return
	}

	var req updateGuestCountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order.GuestCount = req.GuestCount
	db.DB.Save(&order)
	c.JSON(http.StatusOK, order)
}

type changeOrderTableRequest struct {
	TableID uint `json:"table_id" binding:"required"`
}

// ChangeOrderTable ย้ายออเดอร์นั่งทานที่เปิดอยู่ไปยังโต๊ะอื่น เช่น ลูกค้านั่งอยู่แล้วอยากย้ายที่นั่ง
// โต๊ะเดิมจะถูกปล่อยกลับเป็นว่าง โต๊ะใหม่ต้องว่างอยู่ก่อนถึงจะย้ายได้
func ChangeOrderTable(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.OrderType != "dine_in" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ย้ายโต๊ะได้เฉพาะออเดอร์นั่งทานเท่านั้น"})
		return
	}
	if order.Status == "paid" || order.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify a paid or cancelled order"})
		return
	}

	var req changeOrderTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if order.TableID != nil && *order.TableID == req.TableID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "โต๊ะนี้เป็นโต๊ะเดิมของบิลนี้อยู่แล้ว"})
		return
	}

	var newTable models.DiningTable
	if err := db.DB.First(&newTable, req.TableID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	if newTable.Status != "available" {
		c.JSON(http.StatusConflict, gin.H{"error": "โต๊ะใหม่ไม่ว่าง เลือกโต๊ะอื่นแทน"})
		return
	}

	oldTableID := order.TableID

	tx := db.DB.Begin()
	if err := tx.Model(&order).Update("table_id", req.TableID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Model(&models.DiningTable{}).Where("id = ?", req.TableID).Update("status", "occupied").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if oldTableID != nil {
		if err := tx.Model(&models.DiningTable{}).Where("id = ?", *oldTableID).Update("status", "available").Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	tx.Commit()

	db.DB.
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Preload("Table").
		Preload("CreatedByUser").
		First(&order, order.ID)
	c.JSON(http.StatusOK, order)
}

type updateDiscountRequest struct {
	DiscountType  string  `json:"discount_type" binding:"required,oneof=none amount percent"`
	DiscountValue float64 `json:"discount_value"`
}

// UpdateOrderDiscount ตั้ง/แก้ไขส่วนลดของออเดอร์ (ก่อนปิดบิลเท่านั้น) แล้วคำนวณยอดสุทธิใหม่ทันที
func UpdateOrderDiscount(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.Status == "paid" || order.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify a paid or cancelled order"})
		return
	}

	var req updateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DiscountValue < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount_value must not be negative"})
		return
	}
	if req.DiscountType == "percent" && req.DiscountValue > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount percent must not exceed 100"})
		return
	}

	discountValue := req.DiscountValue
	if req.DiscountType == "none" {
		discountValue = 0
	}

	db.DB.Model(&order).Updates(map[string]interface{}{
		"discount_type":  req.DiscountType,
		"discount_value": discountValue,
	})

	recalcTotal(order.ID)

	db.DB.Preload("Items.MenuItem").Preload("Items.Options").Preload("Table").Preload("CreatedByUser").First(&order, order.ID)
	c.JSON(http.StatusOK, order)
}

type payOrderRequest struct {
	Method         string  `json:"method" binding:"required,oneof=cash transfer"`
	ReceivedAmount float64 `json:"received_amount"`
	// TransferRef ใช้เฉพาะ method=transfer เป็นเลขอ้างอิงการโอน (ไม่บังคับ ไม่ใส่ก็ปิดบิลได้ มาเติมทีหลังได้)
	TransferRef string `json:"transfer_ref"`
}

func PayOrder(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := db.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order already paid"})
		return
	}

	var req payOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	// โอนเงินถือว่าลูกค้าโอนมาพอดียอด ไม่มีเงินทอน (ตัวเลขที่ frontend ส่งมาถ้ามีจะถูกมองข้าม กันพลาดคำนวณผิด)
	receivedAmount := req.ReceivedAmount
	if req.Method == "transfer" {
		receivedAmount = order.TotalAmount
	}
	change := receivedAmount - order.TotalAmount
	if change < 0 {
		change = 0
	}

	// ปิดบิล + ออกเลขที่ใบเสร็จ (invoice_no) พร้อมกันในทรานแซกชันเดียว กันเลขที่บิลซ้ำ/กระโดดถ้ามีการ
	// ปิดบิลพร้อมกันหลายบิล (ดู generateInvoiceNo ใน handlers/invoice.go) — ถ้าขั้นตอนไหนพลาด ต้อง rollback
	// ทั้งหมด ไม่ให้ค้างเป็นบิลจ่ายแล้วแต่ไม่มีเลขที่ใบเสร็จ
	tx := db.DB.Begin()

	invoiceNo, err := generateInvoiceNo(tx)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ออกเลขที่ใบเสร็จไม่สำเร็จ"})
		return
	}

	payment := models.Payment{
		OrderID:        order.ID,
		Method:         req.Method,
		Subtotal:       order.Subtotal,
		DiscountAmount: order.DiscountAmount,
		Amount:         order.TotalAmount,
		ReceivedAmount: receivedAmount,
		ChangeAmount:   change,
		TransferRef:    req.TransferRef,
		PaidBy:         userID.(uint),
		PaidAt:         time.Now(),
		InvoiceNo:      invoiceNo,
	}
	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	order.Status = "paid"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if order.TableID != nil {
		tx.Model(&models.DiningTable{}).Where("id = ?", *order.TableID).Update("status", "available")
	}

	tx.Commit()

	db.DB.
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Preload("Table").
		Preload("Payment").
		Preload("Payment.PaidByUser").
		Preload("CreatedByUser").
		First(&order, order.ID)
	c.JSON(http.StatusOK, order)
}
