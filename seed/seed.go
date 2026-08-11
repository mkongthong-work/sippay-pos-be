package seed

import (
	"log"

	"golang.org/x/crypto/bcrypt"

	"pos-backend/db"
	"pos-backend/models"
)

// Run ใส่ข้อมูลตัวอย่างครั้งแรกที่รันระบบ (ทำครั้งเดียว ถ้ามี user อยู่แล้วจะข้าม)
func Run() {
	var count int64
	db.DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	log.Println("seeding initial data...")

	hash, _ := bcrypt.GenerateFromPassword([]byte("admin1234"), bcrypt.DefaultCost)
	admin := models.User{Username: "admin", PasswordHash: string(hash), Name: "Admin", Role: "admin", IsActive: true}
	db.DB.Create(&admin)

	coffee := models.Category{Name: "กาแฟ/เครื่องดื่ม", SortOrder: 1}
	food := models.Category{Name: "อาหาร", SortOrder: 2}
	dessert := models.Category{Name: "ของหวาน", SortOrder: 3}
	db.DB.Create(&coffee)
	db.DB.Create(&food)
	db.DB.Create(&dessert)

	menuItems := []models.MenuItem{
		{CategoryID: coffee.ID, Name: "อเมริกาโน่", Price: 55, IsAvailable: true},
		{CategoryID: coffee.ID, Name: "ลาเต้", Price: 65, IsAvailable: true},
		{CategoryID: coffee.ID, Name: "คาปูชิโน่", Price: 65, IsAvailable: true},
		{CategoryID: food.ID, Name: "ข้าวผัดกะเพราหมู", Price: 65, IsAvailable: true},
		{CategoryID: food.ID, Name: "ข้าวไข่เจียว", Price: 45, IsAvailable: true},
		{CategoryID: dessert.ID, Name: "เค้กช็อกโกแลต", Price: 75, IsAvailable: true},
	}
	for i := range menuItems {
		db.DB.Create(&menuItems[i])
	}

	tables := []models.DiningTable{
		{Name: "T1", Zone: "ในร้าน", Status: "available"},
		{Name: "T2", Zone: "ในร้าน", Status: "available"},
		{Name: "T3", Zone: "ริมหน้าต่าง", Status: "available"},
	}
	for i := range tables {
		db.DB.Create(&tables[i])
	}

	log.Println("seed complete. login with username=admin password=admin1234")
}
