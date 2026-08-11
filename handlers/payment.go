package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
	"pos-backend/utils"
)

// UpdatePaymentSlip แนบ/แก้ไขเลขอ้างอิงการโอน (ref) และ/หรือรูปสลิปของบิลที่ปิดไปแล้ว (method=transfer)
// เรียกได้ทั้งตอนปิดบิลทันที (ถ้าถ่ายรูปสลิปพร้อมกันเลย) หรือย้อนกลับมาแนบ/แก้ไขทีหลังก็ได้ผ่านหน้ารายงาน
// (เช่น พนักงานยุ่งตอนปิดบิล ถ่ายสลิปเก็บไว้ก่อน ค่อยมาอัปโหลดทีหลัง)
// รับแบบ multipart/form-data: ref (text, ไม่บังคับ), slip (ไฟล์รูป, ไม่บังคับ) — ส่งมาแค่ฟิลด์ไหนก็แก้ไข
// เฉพาะฟิลด์นั้น ไม่ต้องส่งครบทั้งคู่
func UpdatePaymentSlip(c *gin.Context) {
	orderID := c.Param("id")
	var payment models.Payment
	if err := db.DB.Where("order_id = ?", orderID).First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ยังไม่มีการชำระเงินของบิลนี้"})
		return
	}

	if ref, ok := c.GetPostForm("ref"); ok {
		payment.TransferRef = ref
	}

	file, err := c.FormFile("slip")
	if err == nil && file != nil {
		slipPath, saveErr := utils.SaveUpload(file, "slips", fmt.Sprintf("payment_%d", payment.ID))
		if saveErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "อัปโหลดสลิปไม่สำเร็จ"})
			return
		}
		payment.SlipImagePath = slipPath
	}

	db.DB.Save(&payment)
	db.DB.Preload("PaidByUser").First(&payment, payment.ID)
	c.JSON(http.StatusOK, payment)
}
