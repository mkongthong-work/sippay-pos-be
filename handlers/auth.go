package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"pos-backend/db"
	"pos-backend/models"
	"pos-backend/utils"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "บัญชีนี้ถูกปิดใช้งานแล้ว ติดต่อผู้ดูแลระบบ"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"name":      user.Name,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

// pinLoginUser คือข้อมูลย่อของพนักงานที่ใช้แสดงในหน้าเลือกคนก่อนกด PIN (เครื่อง POS ที่ใช้ร่วมกันหลายคน)
// ตั้งใจให้เบาที่สุด ไม่มี username/บทบาทอ่อนไหว แค่พอให้แคชเชียร์แตะเลือกตัวเองได้ถูกคน
type pinLoginUser struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ListPinLoginUsers คืนรายชื่อพนักงานที่ "ตั้ง PIN ไว้แล้ว" และยังเปิดใช้งานบัญชีอยู่เท่านั้น — ใช้แสดงเป็น
// การ์ดให้แตะเลือกที่หน้า "เข้าด้วย PIN" ก่อนกรอกตัวเลข ไม่ต้อง login ก่อนจึงเป็น route สาธารณะ (ไม่ผ่าน AuthRequired)
func ListPinLoginUsers(c *gin.Context) {
	var users []models.User
	db.DB.Where("is_active = ? AND pin_hash <> ''", true).Order("name asc").Find(&users)

	result := make([]pinLoginUser, 0, len(users))
	for _, u := range users {
		result = append(result, pinLoginUser{ID: u.ID, Name: u.Name, Role: u.Role})
	}
	c.JSON(http.StatusOK, result)
}

type pinLoginRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Pin    string `json:"pin" binding:"required"`
}

// PinLogin ล็อกอินด้วยการเลือกพนักงาน (จากรายชื่อของ ListPinLoginUsers) แล้วกรอก PIN 6 หลักแทนรหัสผ่าน
// คืนรูปแบบ response เดียวกับ Login ปกติทุกประการ (token + user) ใช้แทนกันได้ที่ฝั่ง frontend
func PinLogin(c *gin.Context) {
	var req pinLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ไม่พบผู้ใช้นี้ หรือยังไม่ได้ตั้ง PIN"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "บัญชีนี้ถูกปิดใช้งานแล้ว ติดต่อผู้ดูแลระบบ"})
		return
	}

	if user.PinHash == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ผู้ใช้นี้ยังไม่ได้ตั้ง PIN"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(req.Pin)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "PIN ไม่ถูกต้อง"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"username":  user.Username,
			"name":      user.Name,
			"role":      user.Role,
			"is_active": user.IsActive,
		},
	})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"name":      user.Name,
		"role":      user.Role,
		"is_active": user.IsActive,
	})
}
