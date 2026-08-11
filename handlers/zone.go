package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

// ListZones คืนโซนทั้งหมด (รวมทั้งที่ปิดใช้งานอยู่ด้วย เพื่อให้หน้าจัดการโซนแสดงครบและสลับเปิดกลับได้)
func ListZones(c *gin.Context) {
	var zones []models.Zone
	db.DB.Order("name asc").Find(&zones)
	c.JSON(http.StatusOK, zones)
}

type createZoneRequest struct {
	Name string `json:"name" binding:"required"`
}

func CreateZone(c *gin.Context) {
	var input createZoneRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Zone
	if db.DB.Where("name = ?", input.Name).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "มีโซนชื่อนี้อยู่แล้ว"})
		return
	}

	zone := models.Zone{Name: input.Name, IsActive: true}
	if err := db.DB.Create(&zone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, zone)
}

type updateZoneRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"is_active"`
}

// UpdateZone ใช้ทั้งแก้ชื่อโซน และเปิด/ปิดใช้งานโซน (เช่น ปิดซ่อม หรือมีการจองที่นั่งไว้ทั้งโซน)
func UpdateZone(c *gin.Context) {
	id := c.Param("id")
	var zone models.Zone
	if err := db.DB.First(&zone, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบโซนนี้"})
		return
	}

	var input updateZoneRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldName := zone.Name

	if input.Name != nil && *input.Name != "" {
		zone.Name = *input.Name
	}
	if input.IsActive != nil {
		zone.IsActive = *input.IsActive
	}

	if err := db.DB.Save(&zone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ถ้าเปลี่ยนชื่อโซน ให้อัปเดตชื่อโซนของโต๊ะทุกตัวที่ผูกกับชื่อเดิมด้วย (Table.Zone เก็บเป็น string)
	if input.Name != nil && *input.Name != "" && *input.Name != oldName {
		db.DB.Model(&models.DiningTable{}).Where("zone = ?", oldName).Update("zone", zone.Name)
	}

	c.JSON(http.StatusOK, zone)
}
