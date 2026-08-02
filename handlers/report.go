package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pos-backend/db"
	"pos-backend/models"
)

type topItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Total    float64 `json:"total"`
}

// DailyReport สรุปยอดขายของวันที่ระบุ (ค่าเริ่มต้นคือวันนี้) โดยนับจากบิลที่ "ชำระเงินแล้ว" เท่านั้น
func DailyReport(c *gin.Context) {
	dateParam := c.Query("date")
	var day time.Time
	var err error
	if dateParam == "" {
		day = time.Now()
	} else {
		day, err = time.Parse("2006-01-02", dateParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
			return
		}
	}

	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(24 * time.Hour)

	var payments []models.Payment
	db.DB.Where("paid_at >= ? AND paid_at < ?", start, end).Find(&payments)

	var totalSales float64
	itemTotals := map[uint]*topItem{}

	for _, p := range payments {
		totalSales += p.Amount
		var order models.Order
		db.DB.Preload("Items").First(&order, p.OrderID)
		for _, it := range order.Items {
			if _, ok := itemTotals[it.MenuItemID]; !ok {
				var mi models.MenuItem
				db.DB.First(&mi, it.MenuItemID)
				itemTotals[it.MenuItemID] = &topItem{Name: mi.Name}
			}
			itemTotals[it.MenuItemID].Quantity += it.Quantity
			itemTotals[it.MenuItemID].Total += it.UnitPrice * float64(it.Quantity)
		}
	}

	var topItems []topItem
	for _, v := range itemTotals {
		topItems = append(topItems, *v)
	}

	c.JSON(http.StatusOK, gin.H{
		"date":        start.Format("2006-01-02"),
		"order_count": len(payments),
		"total_sales": totalSales,
		"top_items":   topItems,
	})
}
