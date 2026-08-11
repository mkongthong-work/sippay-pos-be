package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pos-backend/db"
	"pos-backend/models"
	"pos-backend/utils"
)

// ---- Categories ----

func ListCategories(c *gin.Context) {
	var categories []models.Category
	query := db.DB.Order("sort_order asc")
	if c.Query("include_archived") != "true" {
		query = query.Where("is_archived = ?", false)
	}
	query.Find(&categories)
	c.JSON(http.StatusOK, categories)
}

func CreateCategory(c *gin.Context) {
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.DB.Create(&cat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var cat models.Category
	if err := db.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	var input models.Category
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat.Name = input.Name
	cat.Description = input.Description
	cat.Color = input.Color
	cat.SortOrder = input.SortOrder
	cat.IsEnabled = input.IsEnabled
	// Station เจตนาไม่รับค่าจาก UI ใหม่แล้ว (หน้าจัดการเมนูรุ่นใหม่ไม่มีช่องนี้) แต่ยังคงค่าที่เคยตั้งไว้เดิม
	// ไม่ต้อง overwrite ด้วย input.Station เพราะ input จะเป็นค่าว่างเสมอจาก frontend รุ่นใหม่
	db.DB.Save(&cat)
	c.JSON(http.StatusOK, cat)
}

func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.Category{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ArchiveCategory เก็บหมวดหมู่เข้าคลังเก็บถาวร (ซ่อนจากรายการหลัก แต่กู้คืนได้ ต่างจากลบถาวร)
func ArchiveCategory(c *gin.Context) {
	id := c.Param("id")
	var cat models.Category
	if err := db.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	cat.IsArchived = true
	db.DB.Save(&cat)
	c.JSON(http.StatusOK, cat)
}

// RestoreCategory กู้คืนหมวดหมู่ที่เก็บถาวรไว้กลับมาใช้งานปกติ
func RestoreCategory(c *gin.Context) {
	id := c.Param("id")
	var cat models.Category
	if err := db.DB.First(&cat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}
	cat.IsArchived = false
	db.DB.Save(&cat)
	c.JSON(http.StatusOK, cat)
}

// ---- Menu Items ----

func ListMenuItems(c *gin.Context) {
	var items []models.MenuItem
	query := db.DB.Preload("Category").
		Preload("OptionGroups", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc") }).
		Preload("OptionGroups.Choices", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order asc") })
	if categoryID := c.Query("category_id"); categoryID != "" {
		if id, err := strconv.Atoi(categoryID); err == nil {
			query = query.Where("category_id = ?", id)
		}
	}
	if c.Query("include_archived") != "true" {
		query = query.Where("is_archived = ?", false)
	}
	query.Find(&items)
	c.JSON(http.StatusOK, items)
}

func CreateMenuItem(c *gin.Context) {
	var item models.MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func UpdateMenuItem(c *gin.Context) {
	id := c.Param("id")
	var item models.MenuItem
	if err := db.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}
	var input models.MenuItem
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item.Name = input.Name
	item.Price = input.Price
	item.CategoryID = input.CategoryID
	item.IsAvailable = input.IsAvailable
	item.IsFeatured = input.IsFeatured
	item.IsBestseller = input.IsBestseller
	item.TrackStock = input.TrackStock
	db.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

func DeleteMenuItem(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.MenuItem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ArchiveMenuItem เก็บเมนูเข้าคลังเก็บถาวร (ซ่อนจากรายการหลัก แต่กู้คืนได้ ต่างจากลบถาวร)
func ArchiveMenuItem(c *gin.Context) {
	id := c.Param("id")
	var item models.MenuItem
	if err := db.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}
	item.IsArchived = true
	db.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

// RestoreMenuItem กู้คืนเมนูที่เก็บถาวรไว้กลับมาใช้งานปกติ
func RestoreMenuItem(c *gin.Context) {
	id := c.Param("id")
	var item models.MenuItem
	if err := db.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}
	item.IsArchived = false
	db.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

// UploadMenuItemImage อัปโหลด/แทนที่รูปของเมนูนี้ (multipart form field ชื่อ "image")
// รูปแบบเดียวกับ UpdatePaymentSlip (handlers/payment.go) — เก็บไฟล์แบบสุ่มชื่อกันเดา URL แล้วเซฟ path/URL ไว้
// ในฟิลด์ (utils.SaveUpload สลับเก็บดิสก์เอง/Supabase Storage ให้อัตโนมัติตาม env — ดู utils/storage.go)
func UploadMenuItemImage(c *gin.Context) {
	id := c.Param("id")
	var item models.MenuItem
	if err := db.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบไฟล์รูปภาพ (field 'image')"})
		return
	}

	imagePath, err := utils.SaveUpload(file, "menu-items", fmt.Sprintf("menuitem_%s", id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	item.ImagePath = imagePath
	db.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

// DeleteMenuItemImage ลบรูปเมนูออก (ล้างฟิลด์ path เฉยๆ ไม่ได้ลบไฟล์จริงบนดิสก์)
func DeleteMenuItemImage(c *gin.Context) {
	id := c.Param("id")
	var item models.MenuItem
	if err := db.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}
	item.ImagePath = ""
	db.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

// ---- Menu Option Groups / Choices ----
// ตัวเลือกของเมนู เช่น "ความหวาน" (หวานน้อย/หวานปกติ/หวานมาก), "ไซส์" ฯลฯ
// ตั้งค่าได้ตอนสร้าง/แก้ไขเมนู แล้วจะไปโผล่ให้เลือกตอนสั่งที่หน้าขาย

type optionChoiceInput struct {
	Name       string  `json:"name" binding:"required"`
	PriceDelta float64 `json:"price_delta"`
	IsDefault  bool    `json:"is_default"`
	IsEnabled  bool    `json:"is_enabled"`
}

type createOptionGroupRequest struct {
	Name          string               `json:"name" binding:"required"`
	Description   string               `json:"description"`
	SelectionType string               `json:"selection_type"`
	MinSelect     int                  `json:"min_select"`
	MaxSelect     int                  `json:"max_select"`
	IsRequired    bool                 `json:"is_required"`
	IsEnabled     bool                 `json:"is_enabled"`
	SortOrder     int                  `json:"sort_order"`
	Choices       []optionChoiceInput  `json:"choices" binding:"required,min=1"`
}

func CreateOptionGroup(c *gin.Context) {
	menuItemID := c.Param("id")
	var menuItem models.MenuItem
	if err := db.DB.First(&menuItem, menuItemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}

	var req createOptionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	selectionType := req.SelectionType
	if selectionType != "multi" {
		selectionType = "single"
	}
	maxSelect := req.MaxSelect
	if maxSelect <= 0 {
		if selectionType == "multi" {
			maxSelect = len(req.Choices)
		} else {
			maxSelect = 1
		}
	}

	group := models.MenuOptionGroup{
		MenuItemID:    menuItem.ID,
		Name:          req.Name,
		Description:   req.Description,
		SelectionType: selectionType,
		MinSelect:     req.MinSelect,
		MaxSelect:     maxSelect,
		IsRequired:    req.IsRequired,
		IsEnabled:     true,
		SortOrder:     req.SortOrder,
	}
	if err := db.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, ch := range req.Choices {
		choice := models.MenuOptionChoice{
			OptionGroupID: group.ID,
			Name:          ch.Name,
			PriceDelta:    ch.PriceDelta,
			SortOrder:     i,
			IsDefault:     ch.IsDefault,
			IsEnabled:     true,
		}
		db.DB.Create(&choice)
	}

	orderedChoicesPreload(db.DB).First(&group, group.ID)
	c.JSON(http.StatusCreated, group)
}

type updateOptionGroupRequest struct {
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	SelectionType *string `json:"selection_type"`
	MinSelect     *int    `json:"min_select"`
	MaxSelect     *int    `json:"max_select"`
	IsRequired    *bool   `json:"is_required"`
	IsEnabled     *bool   `json:"is_enabled"`
	SortOrder     *int    `json:"sort_order"`
}

// UpdateOptionGroup แก้ไขกลุ่มตัวเลือกที่มีอยู่แล้ว (ส่งเฉพาะฟิลด์ที่อยากแก้ ฟิลด์ที่ไม่ส่งจะคงค่าเดิม)
func UpdateOptionGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.MenuOptionGroup
	if err := db.DB.First(&group, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option group not found"})
		return
	}

	var req updateOptionGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.SelectionType != nil {
		if *req.SelectionType == "multi" {
			group.SelectionType = "multi"
		} else {
			group.SelectionType = "single"
		}
	}
	if req.MinSelect != nil {
		group.MinSelect = *req.MinSelect
	}
	if req.MaxSelect != nil {
		group.MaxSelect = *req.MaxSelect
	}
	if req.IsRequired != nil {
		group.IsRequired = *req.IsRequired
	}
	if req.IsEnabled != nil {
		group.IsEnabled = *req.IsEnabled
	}
	if req.SortOrder != nil {
		group.SortOrder = *req.SortOrder
	}
	db.DB.Save(&group)

	orderedChoicesPreload(db.DB).First(&group, group.ID)
	c.JSON(http.StatusOK, group)
}

func DeleteOptionGroup(c *gin.Context) {
	id := c.Param("id")
	db.DB.Where("option_group_id = ?", id).Delete(&models.MenuOptionChoice{})
	if err := db.DB.Delete(&models.MenuOptionGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

type addOptionChoiceRequest struct {
	Name       string  `json:"name" binding:"required"`
	PriceDelta float64 `json:"price_delta"`
	IsDefault  bool    `json:"is_default"`
	SortOrder  int     `json:"sort_order"`
}

func AddOptionChoice(c *gin.Context) {
	groupID := c.Param("id")
	var group models.MenuOptionGroup
	if err := db.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option group not found"})
		return
	}
	var req addOptionChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	choice := models.MenuOptionChoice{
		OptionGroupID: group.ID,
		Name:          req.Name,
		PriceDelta:    req.PriceDelta,
		IsDefault:     req.IsDefault,
		IsEnabled:     true,
		SortOrder:     req.SortOrder,
	}
	if err := db.DB.Create(&choice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, choice)
}

type updateOptionChoiceRequest struct {
	Name       *string  `json:"name"`
	PriceDelta *float64 `json:"price_delta"`
	IsDefault  *bool    `json:"is_default"`
	IsEnabled  *bool    `json:"is_enabled"`
	SortOrder  *int     `json:"sort_order"`
}

// UpdateOptionChoice แก้ไขตัวเลือกย่อยที่มีอยู่แล้ว (เช่น ติ๊ก "เริ่มต้น"/"เปิดใช้งาน" หรือแก้ราคา)
func UpdateOptionChoice(c *gin.Context) {
	id := c.Param("id")
	var choice models.MenuOptionChoice
	if err := db.DB.First(&choice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "choice not found"})
		return
	}
	var req updateOptionChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil {
		choice.Name = *req.Name
	}
	if req.PriceDelta != nil {
		choice.PriceDelta = *req.PriceDelta
	}
	if req.IsDefault != nil {
		choice.IsDefault = *req.IsDefault
	}
	if req.IsEnabled != nil {
		choice.IsEnabled = *req.IsEnabled
	}
	if req.SortOrder != nil {
		choice.SortOrder = *req.SortOrder
	}
	db.DB.Save(&choice)
	c.JSON(http.StatusOK, choice)
}

func DeleteOptionChoice(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.MenuOptionChoice{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ---- Category Option Templates ----
// ค่าเริ่มต้นของกลุ่มตัวเลือกที่ตั้งไว้ระดับหมวดหมู่ เช่น หมวด "กาแฟ" ตั้ง "ความหวาน" ไว้ครั้งเดียว
// แล้วนำไป "ใช้" (apply) กับเมนูแต่ละอย่างในหมวดนั้นได้ โดย copy เป็น MenuOptionGroup ของเมนูนั้นจริงๆ

// เรียง Choices ของแต่ละ template ตาม sort_order เสมอ ให้ลำดับที่ตั้งไว้ในหน้าฟอร์ม (ปุ่ม ▲▼) ตรงกับ
// ลำดับที่โชว์เป็น chip ในการ์ดและตอนนำไป apply กับเมนู
func orderedChoicesPreload(query *gorm.DB) *gorm.DB {
	return query.Preload("Choices", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order asc")
	})
}

// ListAllCategoryOptionTemplates คืนกลุ่มตัวเลือก (template) ของทุกหมวดหมู่รวมกัน — ใช้กับหน้าแท็บ
// "ตัวเลือกเสริม" ที่โชว์เป็นการ์ดรวมทุกกลุ่มในที่เดียว ไม่ต้องเลือกหมวดหมู่ก่อนถึงจะเห็น
// ไม่รวมกลุ่มที่เก็บถาวรไว้ (ดู ListArchivedCategoryOptionTemplates สำหรับแท็บ "เก็บถาวร")
func ListAllCategoryOptionTemplates(c *gin.Context) {
	var templates []models.CategoryOptionTemplate
	orderedChoicesPreload(db.DB).Where("is_archived = ?", false).Order("sort_order asc").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

// ListArchivedCategoryOptionTemplates คืนกลุ่มตัวเลือกที่ถูก "เก็บ" ไว้ — ใช้กับแท็บ "เก็บถาวร"
func ListArchivedCategoryOptionTemplates(c *gin.Context) {
	var templates []models.CategoryOptionTemplate
	orderedChoicesPreload(db.DB).Where("is_archived = ?", true).Order("sort_order asc").Find(&templates)
	c.JSON(http.StatusOK, templates)
}

// ListCategoryOptionTemplates คืนเฉพาะ template ของหมวดหมู่เดียว — ใช้ตอนเปิดหน้าแก้ไขเมนูเพื่อโชว์
// "ทางลัดใช้ค่าเริ่มต้นจากหมวดหมู่นี้" (กรองตามหมวดของเมนูที่กำลังแก้ไขอยู่)
func ListCategoryOptionTemplates(c *gin.Context) {
	categoryID := c.Param("id")
	var templates []models.CategoryOptionTemplate
	orderedChoicesPreload(db.DB).Where("category_id = ? AND is_archived = ?", categoryID, false).Find(&templates)
	c.JSON(http.StatusOK, templates)
}

type createCategoryOptionTemplateRequest struct {
	CategoryID    uint                `json:"category_id" binding:"required"`
	Name          string              `json:"name" binding:"required"`
	Description   string              `json:"description"`
	SelectionType string              `json:"selection_type"`
	MinSelect     int                 `json:"min_select"`
	MaxSelect     int                 `json:"max_select"`
	IsRequired    bool                `json:"is_required"`
	SortOrder     int                 `json:"sort_order"`
	Choices       []optionChoiceInput `json:"choices" binding:"required,min=1"`
}

// CreateCategoryOptionTemplate สร้างกลุ่มตัวเลือก (template) ใหม่ ผูกกับหมวดหมู่ที่ระบุมาใน category_id
func CreateCategoryOptionTemplate(c *gin.Context) {
	var req createCategoryOptionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	if err := db.DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	selectionType := req.SelectionType
	if selectionType != "multi" {
		selectionType = "single"
	}
	maxSelect := req.MaxSelect
	if maxSelect <= 0 {
		if selectionType == "multi" {
			maxSelect = len(req.Choices)
		} else {
			maxSelect = 1
		}
	}

	template := models.CategoryOptionTemplate{
		CategoryID:    category.ID,
		Name:          req.Name,
		Description:   req.Description,
		SelectionType: selectionType,
		MinSelect:     req.MinSelect,
		MaxSelect:     maxSelect,
		IsRequired:    req.IsRequired,
		IsEnabled:     true,
		SortOrder:     req.SortOrder,
	}
	if err := db.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for i, ch := range req.Choices {
		choice := models.CategoryOptionTemplateChoice{
			TemplateID: template.ID,
			Name:       ch.Name,
			PriceDelta: ch.PriceDelta,
			SortOrder:  i,
			IsDefault:  ch.IsDefault,
			IsEnabled:  true,
		}
		db.DB.Create(&choice)
	}

	orderedChoicesPreload(db.DB).First(&template, template.ID)
	c.JSON(http.StatusCreated, template)
}

type updateCategoryOptionTemplateRequest struct {
	CategoryID    *uint   `json:"category_id"`
	Name          *string `json:"name"`
	Description   *string `json:"description"`
	SelectionType *string `json:"selection_type"`
	MinSelect     *int    `json:"min_select"`
	MaxSelect     *int    `json:"max_select"`
	IsRequired    *bool   `json:"is_required"`
	IsEnabled     *bool   `json:"is_enabled"`
	SortOrder     *int    `json:"sort_order"`
}

// UpdateCategoryOptionTemplate แก้ไขกลุ่มตัวเลือก (template) ที่มีอยู่แล้ว (ส่งเฉพาะฟิลด์ที่อยากแก้)
func UpdateCategoryOptionTemplate(c *gin.Context) {
	id := c.Param("id")
	var template models.CategoryOptionTemplate
	if err := db.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option template not found"})
		return
	}

	var req updateCategoryOptionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CategoryID != nil {
		template.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		template.Name = *req.Name
	}
	if req.Description != nil {
		template.Description = *req.Description
	}
	if req.SelectionType != nil {
		if *req.SelectionType == "multi" {
			template.SelectionType = "multi"
		} else {
			template.SelectionType = "single"
		}
	}
	if req.MinSelect != nil {
		template.MinSelect = *req.MinSelect
	}
	if req.MaxSelect != nil {
		template.MaxSelect = *req.MaxSelect
	}
	if req.IsRequired != nil {
		template.IsRequired = *req.IsRequired
	}
	if req.IsEnabled != nil {
		template.IsEnabled = *req.IsEnabled
	}
	if req.SortOrder != nil {
		template.SortOrder = *req.SortOrder
	}
	db.DB.Save(&template)

	orderedChoicesPreload(db.DB).First(&template, template.ID)
	c.JSON(http.StatusOK, template)
}

func DeleteCategoryOptionTemplate(c *gin.Context) {
	id := c.Param("id")
	db.DB.Where("template_id = ?", id).Delete(&models.CategoryOptionTemplateChoice{})
	if err := db.DB.Delete(&models.CategoryOptionTemplate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ArchiveCategoryOptionTemplate เก็บกลุ่มตัวเลือกนี้ไว้ (ไม่ลบถาวร) — ไม่โชว์ในหน้า "ตัวเลือกเสริม" หลัก
// อีกต่อไป แต่กู้คืนได้ภายหลังจากแท็บ "เก็บถาวร"
func ArchiveCategoryOptionTemplate(c *gin.Context) {
	id := c.Param("id")
	var template models.CategoryOptionTemplate
	if err := db.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option template not found"})
		return
	}
	template.IsArchived = true
	db.DB.Save(&template)
	c.JSON(http.StatusOK, template)
}

// RestoreCategoryOptionTemplate กู้คืนกลุ่มตัวเลือกที่เก็บถาวรไว้กลับมาใช้งานปกติ
func RestoreCategoryOptionTemplate(c *gin.Context) {
	id := c.Param("id")
	var template models.CategoryOptionTemplate
	if err := db.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option template not found"})
		return
	}
	template.IsArchived = false
	db.DB.Save(&template)
	c.JSON(http.StatusOK, template)
}

type addTemplateChoiceRequest struct {
	Name       string  `json:"name" binding:"required"`
	PriceDelta float64 `json:"price_delta"`
	IsDefault  bool    `json:"is_default"`
	SortOrder  int     `json:"sort_order"`
}

// AddCategoryOptionTemplateChoice เพิ่มตัวเลือกย่อยใหม่เข้ากลุ่ม template ที่มีอยู่แล้ว
func AddCategoryOptionTemplateChoice(c *gin.Context) {
	templateID := c.Param("id")
	var template models.CategoryOptionTemplate
	if err := db.DB.First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option template not found"})
		return
	}
	var req addTemplateChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	choice := models.CategoryOptionTemplateChoice{
		TemplateID: template.ID,
		Name:       req.Name,
		PriceDelta: req.PriceDelta,
		IsDefault:  req.IsDefault,
		IsEnabled:  true,
		SortOrder:  req.SortOrder,
	}
	if err := db.DB.Create(&choice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, choice)
}

type updateTemplateChoiceRequest struct {
	Name       *string  `json:"name"`
	PriceDelta *float64 `json:"price_delta"`
	IsDefault  *bool    `json:"is_default"`
	IsEnabled  *bool    `json:"is_enabled"`
	SortOrder  *int     `json:"sort_order"`
}

// UpdateCategoryOptionTemplateChoice แก้ไขตัวเลือกย่อยของ template ที่มีอยู่แล้ว
func UpdateCategoryOptionTemplateChoice(c *gin.Context) {
	id := c.Param("id")
	var choice models.CategoryOptionTemplateChoice
	if err := db.DB.First(&choice, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "choice not found"})
		return
	}
	var req updateTemplateChoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil {
		choice.Name = *req.Name
	}
	if req.PriceDelta != nil {
		choice.PriceDelta = *req.PriceDelta
	}
	if req.IsDefault != nil {
		choice.IsDefault = *req.IsDefault
	}
	if req.IsEnabled != nil {
		choice.IsEnabled = *req.IsEnabled
	}
	if req.SortOrder != nil {
		choice.SortOrder = *req.SortOrder
	}
	db.DB.Save(&choice)
	c.JSON(http.StatusOK, choice)
}

func DeleteCategoryOptionTemplateChoice(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.CategoryOptionTemplateChoice{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ApplyOptionTemplateToMenuItem คัดลอก template ของหมวดหมู่ ไปสร้างเป็น MenuOptionGroup+Choices จริง
// ของเมนูนั้นๆ (ไม่ได้ผูก reference ตรงๆ กับ template เพื่อไม่ให้แก้ template ภายหลังไม่กระทบเมนูที่ apply ไปแล้ว)
func ApplyOptionTemplateToMenuItem(c *gin.Context) {
	menuItemID := c.Param("id")
	templateID := c.Param("templateId")

	var menuItem models.MenuItem
	if err := db.DB.First(&menuItem, menuItemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}

	var template models.CategoryOptionTemplate
	if err := orderedChoicesPreload(db.DB).First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option template not found"})
		return
	}

	group := models.MenuOptionGroup{
		MenuItemID:    menuItem.ID,
		Name:          template.Name,
		Description:   template.Description,
		SelectionType: template.SelectionType,
		MinSelect:     template.MinSelect,
		MaxSelect:     template.MaxSelect,
		IsRequired:    template.IsRequired,
		IsEnabled:     true,
	}
	if group.SelectionType == "" {
		group.SelectionType = "single"
	}
	if group.MaxSelect <= 0 {
		group.MaxSelect = 1
	}
	if err := db.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, ch := range template.Choices {
		choice := models.MenuOptionChoice{
			OptionGroupID: group.ID,
			Name:          ch.Name,
			PriceDelta:    ch.PriceDelta,
			SortOrder:     ch.SortOrder,
			IsDefault:     ch.IsDefault,
			IsEnabled:     true,
		}
		db.DB.Create(&choice)
	}

	orderedChoicesPreload(db.DB).First(&group, group.ID)
	c.JSON(http.StatusCreated, group)
}
