package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

// GetShopSettings คืนข้อมูลร้านค้า (ชื่อ/ที่อยู่/เบอร์โทร/เลขผู้เสียภาษี) ให้ทุกคนที่ login แล้วเรียกได้
// (ไม่ใช่แค่ admin เพราะต้องใช้โชว์บนหัวใบเสร็จตอนพนักงานพิมพ์บิลด้วย)
func GetShopSettings(c *gin.Context) {
	var settings models.ShopSettings
	if err := db.DB.First(&settings, 1).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ยังไม่มีข้อมูลร้านค้า"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

type updateShopSettingsRequest struct {
	Name          string `json:"name" binding:"required"`
	Address       string `json:"address"`
	Phone         string `json:"phone"`
	TaxID         string `json:"tax_id"`
	PromptPayID   string `json:"promptpay_id"`
	PromptPayName string `json:"promptpay_name"`
}

// UpdateShopSettings แก้ไขข้อมูลร้านค้า (admin เท่านั้น) — เป็นแถวเดียว (id=1) เสมอ ไม่มีการเพิ่ม/ลบแถว
func UpdateShopSettings(c *gin.Context) {
	var settings models.ShopSettings
	if err := db.DB.First(&settings, 1).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ยังไม่มีข้อมูลร้านค้า"})
		return
	}

	var req updateShopSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings.Name = req.Name
	settings.Address = req.Address
	settings.Phone = req.Phone
	settings.TaxID = req.TaxID
	settings.PromptPayID = req.PromptPayID
	settings.PromptPayName = req.PromptPayName
	db.DB.Save(&settings)

	c.JSON(http.StatusOK, settings)
}
