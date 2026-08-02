package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

// ---- Categories ----

func ListCategories(c *gin.Context) {
	var categories []models.Category
	db.DB.Order("sort_order asc").Find(&categories)
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
	cat.SortOrder = input.SortOrder
	cat.Station = input.Station
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

// ---- Menu Items ----

func ListMenuItems(c *gin.Context) {
	var items []models.MenuItem
	query := db.DB.Preload("Category").Preload("OptionGroups.Choices")
	if categoryID := c.Query("category_id"); categoryID != "" {
		if id, err := strconv.Atoi(categoryID); err == nil {
			query = query.Where("category_id = ?", id)
		}
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

// ---- Menu Option Groups / Choices ----
// ตัวเลือกของเมนู เช่น "ความหวาน" (หวานน้อย/หวานปกติ/หวานมาก), "ไซส์" ฯลฯ
// ตั้งค่าได้ตอนสร้าง/แก้ไขเมนู แล้วจะไปโผล่ให้เลือกตอนสั่งที่หน้าขาย

type createOptionGroupRequest struct {
	Name       string `json:"name" binding:"required"`
	IsRequired bool   `json:"is_required"`
	Choices    []struct {
		Name       string  `json:"name" binding:"required"`
		PriceDelta float64 `json:"price_delta"`
	} `json:"choices" binding:"required,min=1"`
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

	group := models.MenuOptionGroup{
		MenuItemID: menuItem.ID,
		Name:       req.Name,
		IsRequired: req.IsRequired,
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
		}
		db.DB.Create(&choice)
	}

	db.DB.Preload("Choices").First(&group, group.ID)
	c.JSON(http.StatusCreated, group)
}

type updateOptionGroupRequest struct {
	Name       *string `json:"name"`
	IsRequired *bool   `json:"is_required"`
}

// UpdateOptionGroup แก้ชื่อ/บังคับให้เลือกของกลุ่มตัวเลือกที่มีอยู่แล้ว (เช่น เปลี่ยนจาก "ต้องเลือก" เป็น "ไม่บังคับ")
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
	if req.IsRequired != nil {
		group.IsRequired = *req.IsRequired
	}
	db.DB.Save(&group)

	db.DB.Preload("Choices").First(&group, group.ID)
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
	}
	if err := db.DB.Create(&choice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, choice)
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

func ListCategoryOptionTemplates(c *gin.Context) {
	categoryID := c.Param("id")
	var templates []models.CategoryOptionTemplate
	db.DB.Preload("Choices").Where("category_id = ?", categoryID).Find(&templates)
	c.JSON(http.StatusOK, templates)
}

type createCategoryOptionTemplateRequest struct {
	Name       string `json:"name" binding:"required"`
	IsRequired bool   `json:"is_required"`
	Choices    []struct {
		Name       string  `json:"name" binding:"required"`
		PriceDelta float64 `json:"price_delta"`
	} `json:"choices" binding:"required,min=1"`
}

func CreateCategoryOptionTemplate(c *gin.Context) {
	categoryID := c.Param("id")
	var category models.Category
	if err := db.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	var req createCategoryOptionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := models.CategoryOptionTemplate{
		CategoryID: category.ID,
		Name:       req.Name,
		IsRequired: req.IsRequired,
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
		}
		db.DB.Create(&choice)
	}

	db.DB.Preload("Choices").First(&template, template.ID)
	c.JSON(http.StatusCreated, template)
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

// ApplyOptionTemplateToMenuItem คัดลอก template ของหมวดหมู่ ไปสร้างเป็น MenuOptionGroup+Choices จริง
// ของเมนูนั้นๆ (ไม่ได้ผูก reference ตรงๆ กับ template เพื่อไม่ให้แก้ template ทีหลังกระทบเมนูที่ apply ไปแล้ว)
func ApplyOptionTemplateToMenuItem(c *gin.Context) {
	menuItemID := c.Param("id")
	templateID := c.Param("templateId")

	var menuItem models.MenuItem
	if err := db.DB.First(&menuItem, menuItemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "menu item not found"})
		return
	}

	var template models.CategoryOptionTemplate
	if err := db.DB.Preload("Choices").First(&template, templateID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "option template not found"})
		return
	}

	group := models.MenuOptionGroup{
		MenuItemID: menuItem.ID,
		Name:       template.Name,
		IsRequired: template.IsRequired,
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
		}
		db.DB.Create(&choice)
	}

	db.DB.Preload("Choices").First(&group, group.ID)
	c.JSON(http.StatusCreated, group)
}
