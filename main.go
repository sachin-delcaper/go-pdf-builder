package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

type ChargeItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type InvoiceRequest struct {
	InvoiceName   string       `json:"invoice_name"`
	InvoiceNumber string       `json:"invoice_number"`
	Date          string       `json:"date"`
	Charges       []ChargeItem `json:"charges"`
}

func generateInvoice(data InvoiceRequest) ([]byte, error) {

	pdf := fpdf.New("P", "mm", "A4", "./fonts")
	pdf.SetMargins(72, 72, 72)

	pdf.SetAutoPageBreak(true, 10)

	pdf.SetFont("Tahoma", "B", 10)

	pdf.AddUTF8Font("Tahoma", "", "tahoma.ttf")
	pdf.AddUTF8Font("Tahoma", "B", "tahomabd.TTF")
	pdf.AddPage()
	pdf.SetFont("Tahoma", "B", 10)
	pdf.MultiCell(90, 8, strings.ToUpper(data.InvoiceName), "", "L", false)
	pdf.Ln(10)
	pdf.SetFont("Tahoma", "", 10)
	pdf.MultiCell(90, 8, "Tax Invoice No.: "+strings.ToUpper(data.InvoiceNumber), "", "R", false)
	pdf.Cell(40, 10, "Tahoma: "+data.Date)
	pdf.Ln(12)

	// Table Header
	pdf.CellFormat(100, 10, "Descriptions", "1", 0, "", false, 0, "")
	pdf.CellFormat(40, 10, "Amount", "1", 1, "", false, 0, "")

	var total float64
	for _, item := range data.Charges {
		pdf.CellFormat(100, 10, item.Description, "1", 0, "", false, 0, "")
		pdf.CellFormat(40, 10, formatAmount(item.Amount), "1", 1, "", false, 0, "")
		total += item.Amount
	}

	// Total
	pdf.CellFormat(100, 10, "Total", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 10, formatAmount(total), "1", 1, "", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

type InvoiceData struct {
	InvoiceName      string            `json:"InvoiceName"`
	InvoiceNumber    string            `json:"InvoiceNumber"`
	Date             string            `json:"Date"`
	Charges          []ChargeItem      `json:"Charges"`
	FullName         string            `json:"FullName"`
	Mobile           string            `json:"Mobile"`
	Email            string            `json:"Email"`
	CustomerAddress  string            `json:"CustomerAddress"`
	ConsignmentNo    string            `json:"ConsignmentNo"`
	Origin           string            `json:"Origin"`
	Destination      string            `json:"Destination"`
	Product          string            `json:"Product"`
	ValueOfGoods     string            `json:"ValueOfGoods"`
	ServiceDate      string            `json:"ServiceDate"`
	GSTIN            string            `json:"GSTIN"`
	HSNCode          string            `json:"HSNCode"`
	StateCode        string            `json:"StateCode"`
	State            string            `json:"State"`
	AmountInWords    string            `json:"AmountInWords"`
	Weight           string            `json:"Weight"`
	ChargeDetails    map[string]string `json:"ChargeDetails"`
	TotalCharges     string            `json:"TotalCharges"`
	QRImagePath      string            `json:"QRImagePath"`
	BarcodeImagePath string            `json:"BarcodeImagePath"`
	LogoPath         string            `json:"LogoPath"`
}

func generatePDF(data InvoiceData) ([]byte, error) {
	// Validate required fields
	if data.InvoiceNumber == "" || data.FullName == "" {
		return nil, fmt.Errorf("invoice number and full name are required")
	}

	pdf := fpdf.New("P", "mm", "A4", "./fonts")
	pdf.SetMargins(5, 5, 5)
	pdf.AddUTF8Font("Tahoma", "", "tahoma.ttf")
	pdf.AddUTF8Font("Tahoma", "B", "tahomabd.TTF")
	pdf.AddPage()
	pdf.SetFont("Tahoma", "", 10)

	// Helper
	addField := func(label string, value string, labelColor, valueColor [3]int) {
		pdf.SetTextColor(labelColor[0], labelColor[1], labelColor[2])
		pdf.CellFormat(40, 6, label, "", 0, "L", false, 0, "")
		pdf.SetTextColor(valueColor[0], valueColor[1], valueColor[2])
		pdf.CellFormat(0, 6, value, "", 1, "L", false, 0, "")
	}

	// Header info
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Tahoma", "B", 12)
	pdf.CellFormat(0, 10, strings.ToUpper(data.FullName), "", 1, "L", false, 0, "")
	pdf.SetFont("Tahoma", "", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(0, 5, data.CustomerAddress, "", "L", false)
	addField("MOBILE:", data.Mobile, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("EMAIL:", data.Email, [3]int{0, 0, 0}, [3]int{0, 0, 255})

	pdf.Ln(4)
	addField("TAX INVOICE NO:", data.InvoiceNumber, [3]int{128, 0, 0}, [3]int{0, 0, 255})
	addField("C.N. NOTE:", data.ConsignmentNo, [3]int{128, 0, 0}, [3]int{0, 0, 255})
	addField("DATE:", data.ServiceDate, [3]int{0, 0, 0}, [3]int{0, 0, 255})

	pdf.Ln(2)
	addField("ORIGIN:", data.Origin, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("DESTINATION:", data.Destination, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("PRODUCT:", data.Product, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("VALUE OF GOODS:", data.ValueOfGoods, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("IN WORDS:", data.AmountInWords, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("WEIGHT:", data.Weight, [3]int{0, 0, 0}, [3]int{0, 0, 255})

	pdf.Ln(2)
	addField("HSN/SSC:", data.HSNCode, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("SERVICE:", "COURIER SERVICE", [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("GST#:", data.GSTIN, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("STATE-CODE:", data.StateCode, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	addField("STATE:", data.State, [3]int{0, 0, 0}, [3]int{0, 0, 255})

	// Charges Table
	pdf.Ln(6)
	pdf.SetTextColor(128, 0, 0)
	pdf.CellFormat(0, 8, "CHARGES", "B", 1, "L", false, 0, "")

	pdf.SetTextColor(0, 0, 0)
	for key, value := range data.ChargeDetails {
		addField(strings.ToUpper(key)+":", value, [3]int{0, 0, 0}, [3]int{0, 0, 255})
	}
	addField("TOTAL CHARGES:", data.TotalCharges, [3]int{128, 0, 0}, [3]int{0, 0, 255})

	// QR and Barcode
	// Reserve space for images even if they're not present
	imageY := 50.0
	imageHeight := 40.0

	// Try to add QR code if path exists and is valid
	if data.QRImagePath != "" {
		pdf.ImageOptions(data.QRImagePath, 150, imageY, 40, imageHeight, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}
	imageY += imageHeight + 10

	// Try to add barcode if path exists and is valid
	if data.BarcodeImagePath != "" {
		pdf.ImageOptions(data.BarcodeImagePath, 150, imageY, 40, 15, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	// Try to add logo if path exists and is valid
	if data.LogoPath != "" {
		pdf.ImageOptions(data.LogoPath, 10, 10, 30, 0, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	// Output
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func formatAmount(val float64) string {
	return fmt.Sprintf("%.2f", val)
}

func main() {
	// Enable Gin's debug mode and logging
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// Add a middleware for request logging
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] | %s | %d | %s | %s | %s | %s | %d\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.ClientIP,
			param.StatusCode,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
			param.BodySize,
		)
	}))

	// Simple invoice endpoint
	r.POST("/invoice", func(c *gin.Context) {
		log.Printf("Received request for simple invoice")
		var req InvoiceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("Processing invoice request: %+v", req)

		pdfBytes, err := generateInvoice(req)
		if err != nil {
			log.Printf("Error generating PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
			return
		}

		log.Printf("Successfully generated PDF of size: %d bytes", len(pdfBytes))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	})

	// Detailed invoice endpoint
	r.POST("/invoice/detailed", func(c *gin.Context) {
		log.Printf("Received request for detailed invoice at path: %s", c.Request.URL.Path)
		log.Printf("Request method: %s", c.Request.Method)
		log.Printf("Request headers: %v", c.Request.Header)

		var req InvoiceData
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
			return
		}
		log.Printf("Processing detailed invoice request: %+v", req)

		pdfBytes, err := generatePDF(req)
		if err != nil {
			log.Printf("Error generating PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
			return
		}

		log.Printf("Successfully generated PDF of size: %d bytes", len(pdfBytes))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	})

	// Template-based invoice endpoint
	r.POST("/invoice/template", func(c *gin.Context) {
		log.Printf("Received request for template-based invoice")
		var req InvoiceTemplateData
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
			return
		}
		log.Printf("Processing template invoice request for: %s", req.InvoiceNumber)

		// Calculate totals if not provided
		if req.SubTotal == 0 && len(req.ChargeItems) > 0 {
			for _, item := range req.ChargeItems {
				req.SubTotal += item.Amount
			}
		}

		// Calculate tax amounts if rates are provided but amounts are not
		if req.CGSTRate > 0 && req.CGSTAmount == 0 {
			req.CGSTAmount = req.SubTotal * req.CGSTRate / 100
		}
		if req.SGSTRate > 0 && req.SGSTAmount == 0 {
			req.SGSTAmount = req.SubTotal * req.SGSTRate / 100
		}
		if req.IGSTRate > 0 && req.IGSTAmount == 0 {
			req.IGSTAmount = req.SubTotal * req.IGSTRate / 100
		}

		// Calculate total if not provided
		if req.TotalAmount == 0 {
			req.TotalAmount = req.SubTotal + req.CGSTAmount + req.SGSTAmount + req.IGSTAmount
		}

		pdfBytes, err := GenerateInvoiceFromTemplate(req)
		if err != nil {
			log.Printf("Error generating PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate PDF: %v", err)})
			return
		}

		log.Printf("Successfully generated template PDF of size: %d bytes", len(pdfBytes))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	})

	// CSV Template-based invoice endpoint
	r.POST("/invoice/template_csv", func(c *gin.Context) {
		log.Printf("Received request for CSV template-based invoice")

		var req CSVTemplateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
			return
		}

		// Parse CSV template
		templatePath := "./assets/pdf_template_1.csv"
		elements, err := ParseCSV(templatePath)
		if err != nil {
			log.Printf("Error parsing CSV template: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse CSV template"})
			return
		}

		// Generate PDF
		outputFile := filepath.Join(os.TempDir(), "invoice.pdf")
		err = GeneratePDF(elements, req.Fields, outputFile)
		if err != nil {
			log.Printf("Error generating PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF generation failed"})
			return
		}

		// Read generated PDF
		pdfBytes, err := os.ReadFile(outputFile)
		if err != nil {
			log.Printf("Error reading generated PDF: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read generated PDF"})
			return
		}

		// Clean up temporary file
		defer os.Remove(outputFile)

		// Set headers for PDF download
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename=invoice.pdf")
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

		// Return PDF as downloadable file
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	})

	// Test endpoint
	r.GET("/test", func(c *gin.Context) {
		log.Printf("Received test request")
		c.JSON(200, gin.H{
			"message": "Server is running",
			"endpoints": []string{
				"POST /invoice",
				"POST /invoice/detailed",
				"POST /invoice/template",
				"POST /invoice/template_csv",
			},
		})
	})

	log.Printf("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
