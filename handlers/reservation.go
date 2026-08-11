package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

type createReservationRequest struct {
	TableID       uint       `json:"table_id" binding:"required"`
	CustomerName  string     `json:"customer_name" binding:"required"`
	CustomerPhone string     `json:"customer_phone"`
	PartySize     int        `json:"party_size"`
	ReservedFor   *time.Time `json:"reserved_for"`
	Note          string     `json:"note"`
}

// ListReservations คืนรายการจอง/กันโต๊ะ เรียงจากใหม่ไปเก่า
// ถ้าไม่ส่ง ?status= มา จะคืนเฉพาะรายการที่ยัง active อยู่ (ยังไม่ seated/cancelled/no_show)
func ListReservations(c *gin.Context) {
	var reservations []models.Reservation
	query := db.DB.Preload("Table").Preload("CreatedByUser").Order("created_at desc")
	if statusParam := c.Query("status"); statusParam != "" {
		query = query.Where("status = ?", statusParam)
	} else {
		query = query.Where("status = ?", "active")
	}
	query.Find(&reservations)
	c.JSON(http.StatusOK, reservations)
}

// CreateReservation กันโต๊ะไว้ให้ลูกค้า (ตอนนี้เลย ถ้าไม่ส่ง reserved_for มา หรือจองล่วงหน้าถ้าส่งเวลามา)
// โต๊ะต้องว่างอยู่ก่อน ถ้าสำเร็จโต๊ะจะถูกเปลี่ยนสถานะเป็น reserved ทันที (ซ่อนจาก POS สำหรับลูกค้าคนอื่น
// แต่พนักงานยังเลือกโต๊ะนี้เปิดบิลให้ลูกค้าที่จองไว้ได้ตามปกติ ระบบจะจับคู่ปิด reservation ให้อัตโนมัติ)
func CreateReservation(c *gin.Context) {
	var req createReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var table models.DiningTable
	if err := db.DB.First(&table, req.TableID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	if table.Status != "available" {
		c.JSON(http.StatusConflict, gin.H{"error": "โต๊ะนี้ไม่ว่าง เลือกโต๊ะอื่นแทน"})
		return
	}

	userID, _ := c.Get("user_id")
	reservation := models.Reservation{
		TableID:       req.TableID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		PartySize:     req.PartySize,
		ReservedFor:   req.ReservedFor,
		Note:          req.Note,
		Status:        "active",
		CreatedBy:     userID.(uint),
	}

	tx := db.DB.Begin()
	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Model(&models.DiningTable{}).Where("id = ?", req.TableID).Update("status", "reserved").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tx.Commit()

	db.DB.Preload("Table").Preload("CreatedByUser").First(&reservation, reservation.ID)
	c.JSON(http.StatusCreated, reservation)
}

// releaseTableIfStillReserved ปล่อยโต๊ะกลับเป็นว่าง เฉพาะกรณีโต๊ะยังอยู่ในสถานะ reserved อยู่จริง
// (กันเคสที่โต๊ะถูกเปิดบิลไปแล้วกลายเป็น occupied ไม่ควรไปทับสถานะให้กลายเป็นว่างผิดๆ)
func releaseTableIfStillReserved(tableID uint) {
	var table models.DiningTable
	if err := db.DB.First(&table, tableID).Error; err == nil && table.Status == "reserved" {
		db.DB.Model(&table).Update("status", "available")
	}
}

// CancelReservation ยกเลิกการจอง/เลิกกันโต๊ะ แล้วปล่อยโต๊ะกลับเป็นว่าง
func CancelReservation(c *gin.Context) {
	id := c.Param("id")
	var reservation models.Reservation
	if err := db.DB.First(&reservation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
		return
	}
	if reservation.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รายการจองนี้ไม่ได้อยู่ในสถานะรอดำเนินการแล้ว"})
		return
	}

	reservation.Status = "cancelled"
	db.DB.Save(&reservation)
	releaseTableIfStillReserved(reservation.TableID)

	c.JSON(http.StatusOK, reservation)
}

// MarkNoShow เหมือน CancelReservation แต่บันทึกสถานะแยกไว้ต่างหาก เผื่ออยากดูสถิติย้อนหลังว่าลูกค้าไม่มาบ่อยแค่ไหน
func MarkNoShow(c *gin.Context) {
	id := c.Param("id")
	var reservation models.Reservation
	if err := db.DB.First(&reservation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
		return
	}
	if reservation.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รายการจองนี้ไม่ได้อยู่ในสถานะรอดำเนินการแล้ว"})
		return
	}

	reservation.Status = "no_show"
	db.DB.Save(&reservation)
	releaseTableIfStillReserved(reservation.TableID)

	c.JSON(http.StatusOK, reservation)
}
