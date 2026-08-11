package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"gorm.io/gorm"

	"pos-backend/assets"
	"pos-backend/db"
	"pos-backend/models"
)

// generateInvoiceNo ออกเลขที่ใบเสร็จ/ใบกำกับภาษีอย่างย่อรูปแบบ INV-YYYYMMDD-00001 (รันต่อวัน ขึ้นวันใหม่
// เริ่ม 00001 ใหม่) ต้องเรียกภายในทรานแซกชันเดียวกับที่จะสร้าง Payment เสมอ (ดู PayOrder ใน order.go)
// เพื่อให้การอ่าน+บวกเลขล่าสุดกับการบันทึกเป็นการดำเนินการเดียวที่ atomic กันเลขที่บิลชนกันถ้ามีการปิดบิล
// พร้อมกันหลายบิล (SQLite ใช้ writer lock ระดับไฟล์อยู่แล้ว ทรานแซกชันที่เขียนพร้อมกันจะถูกรอคิวให้อัตโนมัติ)
func generateInvoiceNo(tx *gorm.DB) (string, error) {
	today := time.Now().Format("20060102")

	var counter models.InvoiceCounter
	err := tx.Where("date_key = ?", today).First(&counter).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		counter = models.InvoiceCounter{DateKey: today, LastNumber: 0}
		if err := tx.Create(&counter).Error; err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	counter.LastNumber++
	if err := tx.Save(&counter).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("INV-%s-%05d", today, counter.LastNumber), nil
}

// GetOrderInvoicePDF สร้างไฟล์ PDF ใบเสร็จ/ใบกำกับภาษีอย่างย่อของบิลที่จ่ายเงินแล้ว ให้ดาวน์โหลด/เปิดดูได้
// ใช้เลขที่ใบเสร็จ (invoice_no) ที่ออกไว้ตอนปิดบิล ต่างจากใบเสร็จที่พิมพ์จากเครื่อง POS โดยตรง (ผ่าน
// window.print() ฝั่ง frontend สำหรับกระดาษ thermal 80mm) — อันนี้เป็นเอกสาร PDF มาตรฐานสำหรับเก็บ/ส่งอีเมล/
// พิมพ์ซ้ำย้อนหลังด้วยเครื่องพิมพ์ทั่วไป
func GetOrderInvoicePDF(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := db.DB.
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Preload("Table").
		Preload("Payment").
		Preload("Payment.PaidByUser").
		Preload("CreatedByUser").
		First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	if order.Payment == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "บิลนี้ยังไม่ได้ชำระเงิน ยังไม่มีใบเสร็จ"})
		return
	}

	var shop models.ShopSettings
	db.DB.First(&shop, 1)

	pdfBytes, err := buildInvoicePDF(order, shop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "สร้าง PDF ไม่สำเร็จ"})
		return
	}

	filename := order.Payment.InvoiceNo + ".pdf"
	c.Header("Content-Disposition", `inline; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// thaiMonthsShort คือชื่อเดือนย่อภาษาไทย ใช้แปลงวันที่แบบเดียวกับที่ frontend แสดงบนใบเสร็จ (ดู
// receipt.component.ts formatThaiDate) ให้ PDF ฝั่ง backend โชว์รูปแบบวันที่เหมือนกัน
var thaiMonthsShort = []string{
	"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
	"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
}

func formatThaiDateTime(t time.Time) string {
	return fmt.Sprintf("%d %s %d %02d:%02d",
		t.Day(), thaiMonthsShort[int(t.Month())-1], t.Year()+543, t.Hour(), t.Minute())
}

func formatBaht(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

// buildInvoicePDF วางเลย์เอาต์เอกสาร A5 หนึ่งหน้า: หัวเอกสาร (ร้าน+เลขที่บิล+วันที่) -> ข้อมูลบิล (โต๊ะ/
// พนักงาน) -> ตารางรายการสินค้า -> สรุปยอด -> ข้อมูลการชำระเงิน -> ท้ายเอกสาร โครงสร้างข้อมูลเดียวกับที่
// ReceiptComponent ฝั่ง frontend ใช้ (order + payment + shop settings) เพื่อให้ตัวเลขตรงกันเป๊ะ
func buildInvoicePDF(order models.Order, shop models.ShopSettings) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A5", "")
	pdf.AddUTF8FontFromBytes("Sarabun", "", assets.SarabunRegularTTF)
	pdf.AddUTF8FontFromBytes("Sarabun", "B", assets.SarabunBoldTTF)
	pdf.SetMargins(12, 12, 12)
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	contentWidth := pageWidth - 24

	// --- หัวเอกสาร: ชื่อร้าน/ที่อยู่/เบอร์โทร/เลขผู้เสียภาษี ---
	pdf.SetFont("Sarabun", "B", 16)
	pdf.CellFormat(contentWidth, 8, shop.Name, "", 1, "C", false, 0, "")

	pdf.SetFont("Sarabun", "", 10)
	if shop.Address != "" {
		pdf.CellFormat(contentWidth, 5, shop.Address, "", 1, "C", false, 0, "")
	}
	line := ""
	if shop.Phone != "" {
		line = "โทร " + shop.Phone
	}
	if shop.TaxID != "" {
		if line != "" {
			line += "   "
		}
		line += "เลขผู้เสียภาษี " + shop.TaxID
	}
	if line != "" {
		pdf.CellFormat(contentWidth, 5, line, "", 1, "C", false, 0, "")
	}

	pdf.Ln(2)
	pdf.SetFont("Sarabun", "B", 13)
	pdf.CellFormat(contentWidth, 7, "ใบเสร็จรับเงิน / ใบกำกับภาษีอย่างย่อ", "", 1, "C", false, 0, "")
	pdf.Ln(2)

	drawDivider(pdf, contentWidth)

	// --- ข้อมูลบิล ---
	pdf.SetFont("Sarabun", "", 10)
	drawRow(pdf, contentWidth, "เลขที่ใบเสร็จ", order.Payment.InvoiceNo)
	drawRow(pdf, contentWidth, "เลขที่ออเดอร์", fmt.Sprintf("#%d", order.ID))
	drawRow(pdf, contentWidth, "วันที่ชำระเงิน", formatThaiDateTime(order.Payment.PaidAt))
	tableLabel := "ซื้อกลับ"
	if order.OrderType == "dine_in" {
		if order.Table != nil {
			tableLabel = order.Table.Name
		} else {
			tableLabel = "-"
		}
	}
	drawRow(pdf, contentWidth, "โต๊ะ", tableLabel)
	if order.CreatedByUser != nil {
		drawRow(pdf, contentWidth, "พนักงาน", order.CreatedByUser.Name)
	}

	pdf.Ln(1)
	drawDivider(pdf, contentWidth)

	// --- ตารางรายการสินค้า ---
	pdf.SetFont("Sarabun", "B", 10)
	colItem := contentWidth * 0.55
	colQty := contentWidth * 0.15
	colAmt := contentWidth * 0.30
	pdf.CellFormat(colItem, 7, "รายการ", "", 0, "L", false, 0, "")
	pdf.CellFormat(colQty, 7, "จำนวน", "", 0, "C", false, 0, "")
	pdf.CellFormat(colAmt, 7, "ราคา (บาท)", "", 1, "R", false, 0, "")
	drawDivider(pdf, contentWidth)

	pdf.SetFont("Sarabun", "", 10)
	for _, item := range order.Items {
		name := item.MenuItem.Name
		if item.IsTakeaway {
			name += " (กลับบ้าน)"
		}
		amount := item.UnitPrice * float64(item.Quantity)
		pdf.CellFormat(colItem, 6, name, "", 0, "L", false, 0, "")
		pdf.CellFormat(colQty, 6, fmt.Sprintf("%d", item.Quantity), "", 0, "C", false, 0, "")
		pdf.CellFormat(colAmt, 6, formatBaht(amount), "", 1, "R", false, 0, "")

		pdf.SetFont("Sarabun", "", 9)
		for _, opt := range item.Options {
			optLine := "  + " + opt.ChoiceName
			if opt.PriceDelta != 0 {
				optLine += fmt.Sprintf(" (+%s)", formatBaht(opt.PriceDelta))
			}
			pdf.CellFormat(contentWidth, 5, optLine, "", 1, "L", false, 0, "")
		}
		if item.Note != "" {
			pdf.CellFormat(contentWidth, 5, "  หมายเหตุ: "+item.Note, "", 1, "L", false, 0, "")
		}
		pdf.SetFont("Sarabun", "", 10)
	}

	pdf.Ln(1)
	drawDivider(pdf, contentWidth)

	// --- สรุปยอด ---
	drawRow(pdf, contentWidth, "ยอดก่อนส่วนลด", formatBaht(order.Subtotal)+" บาท")
	if order.DiscountAmount > 0 {
		drawRow(pdf, contentWidth, "ส่วนลด", "-"+formatBaht(order.DiscountAmount)+" บาท")
	}
	pdf.SetFont("Sarabun", "B", 12)
	pdf.CellFormat(contentWidth/2, 8, "รวมสุทธิ", "", 0, "L", false, 0, "")
	pdf.CellFormat(contentWidth/2, 8, formatBaht(order.TotalAmount)+" บาท", "", 1, "R", false, 0, "")
	pdf.SetFont("Sarabun", "", 10)

	drawDivider(pdf, contentWidth)

	// --- การชำระเงิน ---
	p := order.Payment
	methodLabel := "เงินสด"
	if p.Method == "transfer" {
		methodLabel = "โอนเงิน"
	}
	drawRow(pdf, contentWidth, "วิธีชำระ", methodLabel)
	if p.Method == "cash" {
		drawRow(pdf, contentWidth, "รับเงิน", formatBaht(p.ReceivedAmount)+" บาท")
		drawRow(pdf, contentWidth, "เงินทอน", formatBaht(p.ChangeAmount)+" บาท")
	} else if p.TransferRef != "" {
		drawRow(pdf, contentWidth, "เลขอ้างอิง", p.TransferRef)
	}

	pdf.Ln(4)
	pdf.SetFont("Sarabun", "", 10)
	pdf.CellFormat(contentWidth, 6, "ขอบคุณที่ใช้บริการ", "", 1, "C", false, 0, "")

	return pdfOutputBuffer(pdf)
}

// drawRow วาดแถวข้อมูลแบบ "ป้ายชื่อ : ค่า" ชิดซ้าย/ขวาคนละฝั่ง ใช้ซ้ำหลายจุดในเอกสาร (ข้อมูลบิล/สรุปยอด/การชำระเงิน)
func drawRow(pdf *fpdf.Fpdf, width float64, label string, value string) {
	pdf.CellFormat(width*0.4, 6, label, "", 0, "L", false, 0, "")
	pdf.CellFormat(width*0.6, 6, value, "", 1, "R", false, 0, "")
}

// drawDivider วาดเส้นคั่นบางๆ เต็มความกว้าง content แล้วเว้นบรรทัดเล็กน้อยก่อนเนื้อหาถัดไป
func drawDivider(pdf *fpdf.Fpdf, width float64) {
	x, y := pdf.GetXY()
	pdf.Line(x, y, x+width, y)
	pdf.Ln(3)
}

// pdfOutputBuffer ดึงไบต์ของ PDF ที่สร้างเสร็จแล้วออกมาเป็น []byte เพื่อส่งกลับใน response โดยตรง
// (ไม่ต้องเขียนลงไฟล์ชั่วคราวบนดิสก์)
func pdfOutputBuffer(pdf *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
