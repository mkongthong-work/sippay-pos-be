package handlers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

// เบอร์โทรสมาชิกต้องเป็นตัวเลขล้วน 9-10 หลัก (ตรงกับ validation ฝั่ง frontend)
var memberPhoneFormat = regexp.MustCompile(`^\d{9,10}$`)

// ListMembers คืนรายชื่อสมาชิกทั้งหมด (รวมที่ปิดใช้งานอยู่ด้วย) กรองด้วยชื่อ/เบอร์โทรได้ผ่าน ?search=
func ListMembers(c *gin.Context) {
	var members []models.Member
	query := db.DB.Order("name asc")
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR phone LIKE ?", like, like)
	}
	query.Find(&members)
	c.JSON(http.StatusOK, members)
}

// GetMember คืนข้อมูลสมาชิกรายเดียวตาม id
func GetMember(c *gin.Context) {
	id := c.Param("id")
	var member models.Member
	if err := db.DB.First(&member, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสมาชิกนี้"})
		return
	}
	c.JSON(http.StatusOK, member)
}

// GetMemberByPhone คืนข้อมูลสมาชิกตามเบอร์โทร ใช้ค้นหาเร็วที่หน้าขาย (POS) ตอนลูกค้าแจ้งเบอร์
func GetMemberByPhone(c *gin.Context) {
	phone := c.Param("phone")
	var member models.Member
	if err := db.DB.Where("phone = ?", phone).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสมาชิกนี้"})
		return
	}
	c.JSON(http.StatusOK, member)
}

type createMemberRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

// CreateMember สมัครสมาชิกใหม่ — เริ่มต้นที่แต้ม 0, tier bronze, ยอดใช้จ่ายสะสม 0 เสมอ
func CreateMember(c *gin.Context) {
	var input createMemberRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !memberPhoneFormat.MatchString(input.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "เบอร์โทรต้องเป็นตัวเลข 9-10 หลัก"})
		return
	}

	var existing models.Member
	if db.DB.Where("phone = ?", input.Phone).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "มีสมาชิกที่ใช้เบอร์โทรนี้อยู่แล้ว"})
		return
	}

	member := models.Member{
		Name:          input.Name,
		Phone:         input.Phone,
		PointsBalance: 0,
		Tier:          "bronze",
		TotalSpent:    0,
		IsActive:      true,
	}
	if err := db.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, member)
}

type updateMemberRequest struct {
	Name     *string `json:"name"`
	Phone    *string `json:"phone"`
	IsActive *bool   `json:"is_active"`
}

// UpdateMember แก้ชื่อ/เบอร์โทร/เปิด-ปิดใช้งานสมาชิกภาพ (ทุกฟิลด์เป็น pointer เลือกส่งเฉพาะที่จะแก้ได้)
func UpdateMember(c *gin.Context) {
	id := c.Param("id")
	var member models.Member
	if err := db.DB.First(&member, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสมาชิกนี้"})
		return
	}

	var input updateMemberRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Phone != nil {
		if !memberPhoneFormat.MatchString(*input.Phone) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "เบอร์โทรต้องเป็นตัวเลข 9-10 หลัก"})
			return
		}
		var existing models.Member
		if db.DB.Where("phone = ? AND id != ?", *input.Phone, member.ID).First(&existing).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "มีสมาชิกที่ใช้เบอร์โทรนี้อยู่แล้ว"})
			return
		}
		member.Phone = *input.Phone
	}
	if input.Name != nil {
		member.Name = *input.Name
	}
	if input.IsActive != nil {
		member.IsActive = *input.IsActive
	}

	if err := db.DB.Save(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, member)
}

// GetMemberHistory คืนประวัติการเปลี่ยนแปลงแต้มของสมาชิกรายนี้ ใหม่สุดไว้บนสุด
func GetMemberHistory(c *gin.Context) {
	id := c.Param("id")
	var member models.Member
	if err := db.DB.First(&member, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสมาชิกนี้"})
		return
	}

	var history []models.MemberPointHistory
	db.DB.Where("member_id = ?", member.ID).Order("created_at desc").Find(&history)
	c.JSON(http.StatusOK, history)
}

type adjustPointsRequest struct {
	Change int    `json:"change" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

// AdjustPoints ปรับแต้มสมาชิกด้วยมือ (บวก = ให้แต้มเพิ่ม, ลบ = หัก/แลกแต้ม) บันทึกลง MemberPointHistory
// ทุกครั้งไว้ตรวจสอบย้อนหลังได้ — กันแต้มติดลบ (แลก/หักเกินกว่าที่มี)
func AdjustPoints(c *gin.Context) {
	id := c.Param("id")
	var member models.Member
	if err := db.DB.First(&member, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบสมาชิกนี้"})
		return
	}

	var input adjustPointsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newBalance := member.PointsBalance + input.Change
	if newBalance < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "แต้มคงเหลือไม่พอ"})
		return
	}

	member.PointsBalance = newBalance
	if err := db.DB.Save(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	history := models.MemberPointHistory{
		MemberID: member.ID,
		Change:   input.Change,
		Reason:   input.Reason,
	}
	if err := db.DB.Create(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

// GetLoyaltySettings คืนค่าตั้งค่าระบบสะสมแต้มทั้งร้าน (แถวเดียว id=1 สร้างให้อัตโนมัติตอน migrate ครั้งแรก)
func GetLoyaltySettings(c *gin.Context) {
	var settings models.LoyaltySettings
	if err := db.DB.First(&settings, 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

type updateLoyaltySettingsRequest struct {
	IsEnabled    bool                         `json:"is_enabled"`
	Accumulation models.PointAccumulationRule `json:"accumulation"`
	Redemption   models.RedemptionRule        `json:"redemption"`
	TierRules    models.TierRules             `json:"tier_rules"`
}

// UpdateLoyaltySettings แก้ไขค่าตั้งค่าระบบสะสมแต้มทั้งร้าน (admin เท่านั้น) — แก้ทั้งชุดพร้อมกันเสมอ
func UpdateLoyaltySettings(c *gin.Context) {
	var settings models.LoyaltySettings
	if err := db.DB.First(&settings, 1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input updateLoyaltySettingsRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings.IsEnabled = input.IsEnabled
	settings.Accumulation = input.Accumulation
	settings.Redemption = input.Redemption
	settings.TierRules = input.TierRules

	if err := db.DB.Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}
