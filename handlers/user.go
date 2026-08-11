package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"pos-backend/db"
	"pos-backend/models"
)

// PIN ต้องเป็นตัวเลขล้วน 4-6 หลัก (ใช้ทั้งตอนสร้าง/แก้ไขพนักงาน)
var pinFormat = regexp.MustCompile(`^\d{4,6}$`)

// ListUsers คืนรายชื่อพนักงาน/แอดมินทั้งหมด (รวมบัญชีที่ปิดใช้งานอยู่ด้วย ให้เปิดกลับได้)
// PasswordHash มี json:"-" อยู่แล้วในโมเดล จึงไม่หลุดออกไปกับ response นี้
func ListUsers(c *gin.Context) {
	var users []models.User
	db.DB.Order("name asc").Find(&users)
	c.JSON(http.StatusOK, users)
}

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin staff"`
	// PIN 4-6 หลัก ไว้ล็อกอินเร็วที่เครื่อง POS ร่วม (ไม่บังคับตอนสร้าง มาตั้งทีหลังตอนแก้ไขพนักงานก็ได้)
	Pin string `json:"pin"`
}

func CreateUser(c *gin.Context) {
	var input createUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Pin != "" && !pinFormat.MatchString(input.Pin) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN ต้องเป็นตัวเลข 4-6 หลัก"})
		return
	}

	var existing models.User
	if db.DB.Where("username = ?", input.Username).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "มีชื่อผู้ใช้นี้อยู่แล้ว"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ตั้งรหัสผ่านไม่สำเร็จ"})
		return
	}

	user := models.User{
		Username:     input.Username,
		PasswordHash: string(hash),
		Name:         input.Name,
		Role:         input.Role,
		IsActive:     true,
	}
	if input.Pin != "" {
		pinHash, err := bcrypt.GenerateFromPassword([]byte(input.Pin), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ตั้ง PIN ไม่สำเร็จ"})
			return
		}
		user.PinHash = string(pinHash)
		now := time.Now()
		user.PinUpdatedAt = &now
	}
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

type updateUserRequest struct {
	Name     *string `json:"name"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin staff"`
	Password *string `json:"password"`
	IsActive *bool   `json:"is_active"`
	// PIN 4-6 หลัก — ส่งค่ามาตั้ง/แก้ไข PIN ใหม่, ส่งเป็นสตริงว่าง "" เพื่อล้าง PIN ทิ้ง (ปิดการล็อกอินด้วย PIN
	// ของคนนี้), ไม่ส่งฟิลด์นี้มาเลย (nil) แปลว่าไม่แตะ PIN เดิม
	Pin *string `json:"pin"`
}

// UpdateUser ใช้แก้ชื่อ/role/รีเซ็ตรหัสผ่าน/เปิด-ปิดใช้งานบัญชี (ทุกฟิลด์เป็น pointer เลือกส่งเฉพาะที่จะแก้ได้)
// กันแอดมินพลาดปิดใช้งาน หรือลด role ตัวเองออกจาก admin จนอาจไม่เหลือใครจัดการระบบต่อได้
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบผู้ใช้นี้"})
		return
	}

	var input updateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Password != nil && *input.Password != "" && len(*input.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร"})
		return
	}
	if input.Pin != nil && *input.Pin != "" && !pinFormat.MatchString(*input.Pin) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PIN ต้องเป็นตัวเลข 4-6 หลัก"})
		return
	}

	currentUserIDRaw, _ := c.Get("user_id")
	currentUserID, _ := currentUserIDRaw.(uint)
	targetID, _ := strconv.ParseUint(id, 10, 64)
	isSelf := currentUserID == uint(targetID)

	if isSelf && input.IsActive != nil && !*input.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ปิดใช้งานบัญชีตัวเองไม่ได้"})
		return
	}
	if isSelf && input.Role != nil && *input.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "เปลี่ยน role ตัวเองออกจาก admin ไม่ได้"})
		return
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}
	if input.Password != nil && *input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ตั้งรหัสผ่านไม่สำเร็จ"})
			return
		}
		user.PasswordHash = string(hash)
	}
	if input.Pin != nil {
		if *input.Pin == "" {
			// ส่งค่าว่างมา = ล้าง PIN ทิ้ง (ปิดการล็อกอินด้วย PIN ของคนนี้ไปเลย)
			user.PinHash = ""
			user.PinUpdatedAt = nil
		} else {
			pinHash, err := bcrypt.GenerateFromPassword([]byte(*input.Pin), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "ตั้ง PIN ไม่สำเร็จ"})
				return
			}
			user.PinHash = string(pinHash)
			now := time.Now()
			user.PinUpdatedAt = &now
		}
	}

	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
