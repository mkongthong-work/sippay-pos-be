package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
	"pos-backend/utils"
)

// AuthRequired ตรวจสอบ JWT token จาก header Authorization: Bearer <token>
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// เช็คว่าบัญชียังเปิดใช้งานอยู่ทุกครั้ง (ไม่ใช่แค่ตอน login) เผื่อแอดมินปิดใช้งานบัญชีนี้กลางคัน
		// ขณะ token เดิมยังไม่หมดอายุ (token อายุ 12 ชม.) จะได้ตัดสิทธิ์ทันที ไม่ต้องรอ token หมดอายุเอง
		var user models.User
		if err := db.DB.First(&user, claims.UserID).Error; err != nil || !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminOnly อนุญาตเฉพาะผู้ใช้ที่มี role เป็น admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin only"})
			return
		}
		c.Next()
	}
}
