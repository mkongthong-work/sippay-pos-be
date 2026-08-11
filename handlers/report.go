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

// orderSummary คือแถวสรุปของบิลแต่ละใบที่ปิดในวันนั้น ใช้ให้ admin เช็คย้อนหลังได้ว่าใครเปิด/ใครปิดบิล
// แนบข้อมูลการชำระเงินไว้ด้วย (วิธีจ่าย/เลขอ้างอิงโอน/สลิป) ให้หน้ารายงานเช็ค หรือแนบสลิปย้อนหลังได้
// ถ้าตอนปิดบิลพนักงานยังไม่ได้ถ่ายสลิป
type orderSummary struct {
	OrderID       uint      `json:"order_id"`
	OrderType     string    `json:"order_type"`
	TableName     string    `json:"table_name"`
	TotalAmount   float64   `json:"total_amount"`
	CreatedByName string    `json:"created_by_name"`
	PaidByName    string    `json:"paid_by_name"`
	PaidAt        time.Time `json:"paid_at"`
	PaymentMethod string    `json:"payment_method"`
	TransferRef   string    `json:"transfer_ref"`
	SlipImagePath string    `json:"slip_image_path"`
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
	var orderSummaries []orderSummary

	for _, p := range payments {
		totalSales += p.Amount
		var order models.Order
		db.DB.
			Preload("Items").
			Preload("Table").
			Preload("CreatedByUser").
			First(&order, p.OrderID)
		for _, it := range order.Items {
			if _, ok := itemTotals[it.MenuItemID]; !ok {
				var mi models.MenuItem
				db.DB.First(&mi, it.MenuItemID)
				itemTotals[it.MenuItemID] = &topItem{Name: mi.Name}
			}
			itemTotals[it.MenuItemID].Quantity += it.Quantity
			itemTotals[it.MenuItemID].Total += it.UnitPrice * float64(it.Quantity)
		}

		tableName := ""
		if order.Table != nil {
			tableName = order.Table.Name
		}
		createdByName := ""
		if order.CreatedByUser != nil {
			createdByName = order.CreatedByUser.Name
		}
		var paidByUser models.User
		paidByName := ""
		if err := db.DB.First(&paidByUser, p.PaidBy).Error; err == nil {
			paidByName = paidByUser.Name
		}

		orderSummaries = append(orderSummaries, orderSummary{
			OrderID:       order.ID,
			OrderType:     order.OrderType,
			TableName:     tableName,
			TotalAmount:   p.Amount,
			CreatedByName: createdByName,
			PaidByName:    paidByName,
			PaidAt:        p.PaidAt,
			PaymentMethod: p.Method,
			TransferRef:   p.TransferRef,
			SlipImagePath: p.SlipImagePath,
		})
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
		"orders":      orderSummaries,
	})
}

// dailySalesPoint คือยอดขายรวมของวันหนึ่งวัน ใช้ทำกราฟแนวโน้ม (เช่น 7 วัน/30 วัน/รายเดือนย้อนหลัง)
type dailySalesPoint struct {
	Date       string  `json:"date"`
	TotalSales float64 `json:"total_sales"`
	OrderCount int     `json:"order_count"`
}

// SalesRange คืนยอดขายรวมรายวันของทุกวันในช่วง from-to (รวมวันทั้งสองฝั่ง) นับจากบิลที่ "ชำระเงินแล้ว"
// เท่านั้นเหมือนกับ DailyReport ใช้ทำกราฟแนวโน้มที่หน้ารายงาน วันไหนไม่มีบิลเลยจะเติมยอดขาย 0 ให้
// เพื่อให้กราฟไม่ขาดช่วง
func SalesRange(c *gin.Context) {
	fromParam := c.Query("from")
	toParam := c.Query("to")
	if fromParam == "" || toParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ต้องระบุ from และ to (YYYY-MM-DD)"})
		return
	}

	from, err := time.Parse("2006-01-02", fromParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format, use YYYY-MM-DD"})
		return
	}
	to, err := time.Parse("2006-01-02", toParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format, use YYYY-MM-DD"})
		return
	}

	rangeStart := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	rangeEnd := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, to.Location()).Add(24 * time.Hour)

	if rangeEnd.Before(rangeStart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from ต้องมาก่อน to"})
		return
	}
	// กันเผลอขอช่วงกว้างเกินไปจนดึงข้อมูลหนักเกินจำเป็น
	if rangeEnd.Sub(rangeStart) > 366*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ช่วงวันที่กว้างเกินไป (สูงสุด 366 วัน)"})
		return
	}

	var payments []models.Payment
	db.DB.Where("paid_at >= ? AND paid_at < ?", rangeStart, rangeEnd).Find(&payments)

	totals := map[string]*dailySalesPoint{}
	for _, p := range payments {
		key := p.PaidAt.Format("2006-01-02")
		if _, ok := totals[key]; !ok {
			totals[key] = &dailySalesPoint{Date: key}
		}
		totals[key].TotalSales += p.Amount
		totals[key].OrderCount++
	}

	days := []dailySalesPoint{}
	for d := rangeStart; d.Before(rangeEnd); d = d.Add(24 * time.Hour) {
		key := d.Format("2006-01-02")
		if point, ok := totals[key]; ok {
			days = append(days, *point)
		} else {
			days = append(days, dailySalesPoint{Date: key})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"from": rangeStart.Format("2006-01-02"),
		"to":   to.Format("2006-01-02"),
		"days": days,
	})
}
