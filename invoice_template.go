package main

import (
	"bytes"
	"log"
	"os"
	"strings"

	"github.com/go-pdf/fpdf"
)

// Initialize logger at package level
var logger = log.New(os.Stdout, "[InvoiceTemplate] ", log.LstdFlags)

// InvoiceTemplateData holds all the dynamic data for the invoice
type InvoiceTemplateData struct {
	// Header Information
	InvoiceTitle   string
	CompanyName    string
	CompanyAddress string
	CompanyGSTIN   string
	CompanyPhone   string
	CompanyEmail   string

	// Invoice Details
	InvoiceNumber string
	InvoiceDate   string
	DueDate       string

	// Customer Information
	CustomerName    string
	CustomerAddress string
	CustomerGSTIN   string
	CustomerPhone   string
	CustomerEmail   string

	// Shipping Details
	ConsignmentNo string
	Origin        string
	Destination   string
	Weight        string
	Product       string
	ServiceDate   string

	// Financial Details
	ValueOfGoods string
	HSNCode      string
	StateCode    string
	State        string

	// Charges
	ChargeItems   []ChargeItem
	SubTotal      float64
	CGSTRate      float64
	CGSTAmount    float64
	SGSTRate      float64
	SGSTAmount    float64
	IGSTRate      float64
	IGSTAmount    float64
	TotalAmount   float64
	AmountInWords string

	// Optional Images
	LogoPath    string
	QRCodePath  string
	BarcodePath string
}

// GenerateInvoiceFromTemplate creates a PDF invoice using the template data
func GenerateInvoiceFromTemplate(data InvoiceTemplateData) ([]byte, error) {
	logger.Println("Starting invoice generation")
	logger.Printf("Processing invoice number: %s", data.InvoiceNumber)

	pdf := fpdf.New("P", "mm", "A4", "./fonts")
	pdf.SetMargins(10, 10, 10)
	width, height := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	width = width - left - right
	logger.Printf("Page Size: %.2f x %.2f mm", width, height)

	pdf.AddUTF8Font("Tahoma", "", "tahoma.ttf")
	pdf.AddUTF8Font("Tahoma", "B", "tahomabd.TTF")
	pdf.AddPage()
	logger.Println("Added fonts and created new page")

	// Colors
	blackColor := [3]int{0, 0, 0}  // For static labels
	blueColor := [3]int{0, 0, 255} // For dynamic data
	// redColor := [3]int{128, 0, 0}      // For important labels
	// grayColor := [3]int{128, 128, 128} // For borders

	// Helper functions
	setColor := func(color [3]int) {
		pdf.SetTextColor(color[0], color[1], color[2])
	}

	// drawLine := func(x1, y1, x2, y2 float64) {
	// 	pdf.SetDrawColor(grayColor[0], grayColor[1], grayColor[2])
	// 	pdf.Line(x1, y1, x2, y2)
	// }

	// addLabelValue := func(label string, value string, labelColor, valueColor [3]int, newLine bool) {
	// 	setColor(labelColor)
	// 	pdf.CellFormat(50, 6, label, "", 0, "L", false, 0, "")
	// 	setColor(valueColor)
	// 	if newLine {
	// 		pdf.CellFormat(0, 6, value, "", 1, "L", false, 0, "")
	// 	} else {
	// 		pdf.CellFormat(80, 6, value, "", 0, "L", false, 0, "")
	// 	}
	// }

	// ===== NEW LAYOUT STARTS HERE =====
	// Header Section
	x := 4.0
	y := 4.0
	pdf.SetXY(x, y)

	// Title
	pdf.SetFont("Tahoma", "B", 10)
	setColor(blackColor)
	pdf.CellFormat(0, 2, strings.ToUpper(data.InvoiceTitle), "", 1, "L", false, 0, "")

	// Invoice Number
	x = width - right - 100
	pdf.SetX(x)
	pdf.CellFormat(0, 2, strings.ToUpper("Tax Invoice No."), "", 1, "R", false, 0, "")
	pdf.SetFont("Tahoma", "B", 8)
	setColor(blueColor)
	pdf.CellFormat(0, 6, strings.ToUpper(data.InvoiceNumber), "", 1, "R", false, 0, "")

	// Add vertical space
	// pdf.Ln(10)

	// Table Section
	tableX := 4.0
	tableY := pdf.GetY()
	pdf.SetXY(tableX, tableY)

	col1Width := width/2 + left/2
	col2Width := col1Width
	rowHeight := 10.0

	pdf.Rect(x, y, col1Width, col2Width, "style")

	pdf.SetFont("Tahoma", "", 9)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetTextColor(0, 0, 0)

	// Row 1

	pdf.ImageOptions(
		"./assets/smile-logo_small.png",
		30.4, 15, // X, Y (with slight padding)
		40, 17.5, // width, height inside the 32x32 cell
		false,
		fpdf.ImageOptions{ImageType: "PNG", ReadDpi: true},
		0,
		"",
	)

	// pdf.CellFormat(col1Width, 32, "Customer Name", "1", 0, "L", false, 0, "")
	// pdf.MultiCell(col1Width, 5, "Customer Name is too long and wraps automatically alsdjflasjdf lajsdfl jasldfj laskjdflkasjdflajs dflkjalsdfjlkasj fd", "1", "L", false)

	pdf.CellFormat(col2Width, rowHeight, data.CustomerName, "1", 1, "L", false, 0, "")

	// ===== OLD LAYOUT (COMMENTED FOR REFERENCE) =====
	/*
		// Company Logo (if provided)
		if data.LogoPath != "" {
			logger.Printf("Adding company logo from: %s", data.LogoPath)
			pdf.ImageOptions(data.LogoPath, 10, 10, 40, 0, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		}

		// Company Header
		logger.Println("Adding company header")
		pdf.SetFont("Tahoma", "B", 10)
		setColor(blackColor)

		pdf.SetXY(width/2, 50)
		pdf.CellFormat(0, 8, strings.ToUpper(data.CompanyName), "", 1, "L", false, 0, "")

		pdf.SetFont("Tahoma", "", 10)
		pdf.SetX(60)
		pdf.MultiCell(100, 5, data.CompanyAddress, "", "L", false)

		pdf.SetX(60)
		addLabelValue("GSTIN:", data.CompanyGSTIN, blackColor, blueColor, false)
		pdf.Ln(6)

		// Invoice Title
		pdf.SetFont("Tahoma", "B", 14)
		setColor(redColor)
		pdf.CellFormat(0, 10, "TAX INVOICE", "", 1, "C", false, 0, "")
		pdf.Ln(2)

		// Invoice Details Section
		pdf.SetFont("Tahoma", "", 10)
		y := pdf.GetY()

		// Left side - Invoice details
		pdf.SetXY(10, y)
		addLabelValue("Invoice No:", data.InvoiceNumber, redColor, blueColor, true)
		addLabelValue("Invoice Date:", data.InvoiceDate, blackColor, blueColor, true)
		addLabelValue("Due Date:", data.DueDate, blackColor, blueColor, true)

		// Right side - Service details
		pdf.SetXY(110, y)
		addLabelValue("C.N. Note:", data.ConsignmentNo, redColor, blueColor, true)
		pdf.SetX(110)
		addLabelValue("Service Date:", data.ServiceDate, blackColor, blueColor, true)
		pdf.SetX(110)
		addLabelValue("Origin:", data.Origin, blackColor, blueColor, true)
		pdf.SetX(110)
		addLabelValue("Destination:", data.Destination, blackColor, blueColor, true)

		pdf.Ln(5)
		drawLine(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(5)

		// Customer Details
		pdf.SetFont("Tahoma", "B", 11)
		setColor(blackColor)
		pdf.CellFormat(0, 7, "BILL TO:", "", 1, "L", false, 0, "")

		pdf.SetFont("Tahoma", "B", 10)
		setColor(blueColor)
		pdf.CellFormat(0, 6, strings.ToUpper(data.CustomerName), "", 1, "L", false, 0, "")

		pdf.SetFont("Tahoma", "", 10)
		pdf.MultiCell(100, 5, data.CustomerAddress, "", "L", false)

		addLabelValue("GSTIN:", data.CustomerGSTIN, blackColor, blueColor, false)
		addLabelValue("Phone:", data.CustomerPhone, blackColor, blueColor, true)
		addLabelValue("Email:", data.CustomerEmail, blackColor, blueColor, true)

		pdf.Ln(3)
		drawLine(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(5)

		// Additional Details
		addLabelValue("Product:", data.Product, blackColor, blueColor, false)
		addLabelValue("Weight:", data.Weight, blackColor, blueColor, true)
		addLabelValue("Value of Goods:", data.ValueOfGoods, blackColor, blueColor, false)
		addLabelValue("HSN/SAC:", data.HSNCode, blackColor, blueColor, true)
		addLabelValue("State:", data.State, blackColor, blueColor, false)
		addLabelValue("State Code:", data.StateCode, blackColor, blueColor, true)

		pdf.Ln(5)

		// Charges Table Header
		pdf.SetFont("Tahoma", "B", 10)
		setColor(blackColor)
		pdf.SetFillColor(240, 240, 240)
		pdf.CellFormat(10, 8, "S.No", "1", 0, "C", true, 0, "")
		pdf.CellFormat(100, 8, "Description", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 8, "HSN/SAC", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 8, "Qty", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 8, "Amount", "1", 1, "C", true, 0, "")

		// Charges Table Body
		pdf.SetFont("Tahoma", "", 10)
		setColor(blueColor)
		for i, item := range data.ChargeItems {
			pdf.CellFormat(10, 7, fmt.Sprintf("%d", i+1), "1", 0, "C", false, 0, "")
			pdf.CellFormat(100, 7, item.Description, "1", 0, "L", false, 0, "")
			pdf.CellFormat(30, 7, data.HSNCode, "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 7, "1", "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 7, fmt.Sprintf("%.2f", item.Amount), "1", 1, "R", false, 0, "")
		}

		// Subtotal
		setColor(blackColor)
		pdf.CellFormat(140, 7, "Sub Total", "1", 0, "R", false, 0, "")
		setColor(blueColor)
		pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", data.SubTotal), "1", 1, "R", false, 0, "")

		// Tax rows
		if data.CGSTAmount > 0 {
			setColor(blackColor)
			pdf.CellFormat(140, 7, fmt.Sprintf("CGST @ %.1f%%", data.CGSTRate), "1", 0, "R", false, 0, "")
			setColor(blueColor)
			pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", data.CGSTAmount), "1", 1, "R", false, 0, "")
		}

		if data.SGSTAmount > 0 {
			setColor(blackColor)
			pdf.CellFormat(140, 7, fmt.Sprintf("SGST @ %.1f%%", data.SGSTRate), "1", 0, "R", false, 0, "")
			setColor(blueColor)
			pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", data.SGSTAmount), "1", 1, "R", false, 0, "")
		}

		if data.IGSTAmount > 0 {
			setColor(blackColor)
			pdf.CellFormat(140, 7, fmt.Sprintf("IGST @ %.1f%%", data.IGSTRate), "1", 0, "R", false, 0, "")
			setColor(blueColor)
			pdf.CellFormat(50, 7, fmt.Sprintf("%.2f", data.IGSTAmount), "1", 1, "R", false, 0, "")
		}

		// Total
		pdf.SetFont("Tahoma", "B", 11)
		setColor(redColor)
		pdf.CellFormat(140, 8, "TOTAL AMOUNT", "1", 0, "R", false, 0, "")
		setColor(blueColor)
		pdf.CellFormat(50, 8, fmt.Sprintf("%.2f", data.TotalAmount), "1", 1, "R", false, 0, "")

		// Amount in words
		pdf.Ln(3)
		pdf.SetFont("Tahoma", "", 10)
		setColor(blackColor)
		pdf.CellFormat(30, 6, "Amount in Words:", "", 0, "L", false, 0, "")
		setColor(blueColor)
		pdf.CellFormat(0, 6, data.AmountInWords, "", 1, "L", false, 0, "")

		// QR Code and Barcode (if provided)
		if data.QRCodePath != "" || data.BarcodePath != "" {
			currentY := pdf.GetY()

			if data.QRCodePath != "" {
				pdf.ImageOptions(data.QRCodePath, 150, currentY+5, 30, 30, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
			}

			if data.BarcodePath != "" {
				pdf.ImageOptions(data.BarcodePath, 150, currentY+40, 40, 10, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
			}
		}

		// Footer
		pdf.SetY(-30)
		drawLine(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(2)

		pdf.SetFont("Tahoma", "", 9)
		setColor(grayColor)
		pdf.CellFormat(0, 5, "This is a computer generated invoice and does not require signature.", "", 1, "C", false, 0, "")
		pdf.CellFormat(0, 5, "Thank you for your business!", "", 1, "C", false, 0, "")
	*/

	// Generate PDF
	logger.Println("Generating final PDF")
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		logger.Printf("Error generating PDF: %v", err)
		return nil, err
	}

	logger.Printf("Successfully generated PDF of size: %d bytes", buf.Len())
	return buf.Bytes(), nil
}
