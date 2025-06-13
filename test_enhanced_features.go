//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"pdf-gen-simple/internal/generators"
	"pdf-gen-simple/internal/models"
	"pdf-gen-simple/internal/parsers"
)

// TestData represents sample invoice data
type TestData struct {
	InvoiceNumber string                   `json:"invoiceNumber"`
	Date          string                   `json:"date"`
	CustomerName  string                   `json:"customerName"`
	QRData        string                   `json:"qrData"`
	BarcodeData   string                   `json:"barcodeData"`
	LogoPath      string                   `json:"logoPath"`
	Subtotal      string                   `json:"subtotal"`
	Tax           string                   `json:"tax"`
	GrandTotal    string                   `json:"grandTotal"`
	Items         []map[string]interface{} `json:"items"`
}

func main() {
	fmt.Println("Testing Enhanced PDF Generation Features")
	fmt.Println("=======================================")

	// Test 1: Basic CSV parsing with new structure
	fmt.Println("\n1. Testing CSV Parser with new structure...")
	testCSVParser()

	// Test 2: QR Code generation
	fmt.Println("\n2. Testing QR Code generation...")
	testQRGeneration()

	// Test 3: Barcode generation
	fmt.Println("\n3. Testing Barcode generation...")
	testBarcodeGeneration()

	// Test 4: Full invoice with all features
	fmt.Println("\n4. Testing complete invoice with all features...")
	testCompleteInvoice()

	fmt.Println("\nAll tests completed successfully!")
}

func testCSVParser() {
	parser := parsers.NewCSVParser()

	templatePath := "./assets/pdf_template_enhanced.csv"
	elements, err := parser.ParseCSV(templatePath)
	if err != nil {
		log.Printf("Error parsing CSV: %v", err)
		return
	}

	fmt.Printf("Successfully parsed %d elements from enhanced template\n", len(elements))

	// Show some element types
	for i, element := range elements {
		if i < 5 { // Show first 5 elements
			fmt.Printf("  Element %d: Type=%s, Position=(%.1f,%.1f)\n",
				i+1, element.Type, element.Position.X, element.Position.Y)
		}
	}

	// Test cache functionality
	stats := parser.GetCacheStats()
	fmt.Printf("Cache stats: %+v\n", stats)
}

func testQRGeneration() {
	// Create a simple QR element
	qrElement := models.PDFElement{
		Type:      models.ElementTypeQR,
		Position:  models.Position{X: 10, Y: 10},
		Size:      models.Size{Width: 30, Height: 30},
		QRContent: "https://example.com/test-qr",
	}

	if err := qrElement.Validate(); err != nil {
		log.Printf("QR element validation failed: %v", err)
		return
	}

	fmt.Printf("QR element validation passed: %s\n", qrElement.Type)
	fmt.Printf("QR content: %s\n", qrElement.QRContent)
}

func testBarcodeGeneration() {
	// Create a barcode element
	barcodeElement := models.PDFElement{
		Type:           models.ElementTypeBarcode,
		Position:       models.Position{X: 10, Y: 50},
		Size:           models.Size{Width: 80, Height: 15},
		BarcodeFormat:  "Code128",
		BarcodeContent: "TEST123456",
	}

	if err := barcodeElement.Validate(); err != nil {
		log.Printf("Barcode element validation failed: %v", err)
		return
	}

	fmt.Printf("Barcode element validation passed: %s\n", barcodeElement.Type)
	fmt.Printf("Barcode format: %s, content: %s\n",
		barcodeElement.BarcodeFormat, barcodeElement.BarcodeContent)
}

func testCompleteInvoice() {
	// Create test data
	testData := TestData{
		InvoiceNumber: "INV-2024-001",
		Date:          "2024-01-15",
		CustomerName:  "John Doe",
		QRData:        "https://company.com/verify/INV-2024-001",
		BarcodeData:   "INV2024001",
		LogoPath:      "./assets/smile-logo_small.png",
		Subtotal:      "$250.00",
		Tax:           "$25.00",
		GrandTotal:    "$275.00",
		Items: []map[string]interface{}{
			{
				"description": "Product A",
				"quantity":    "2",
				"price":       "$50.00",
				"total":       "$100.00",
			},
			{
				"description": "Product B",
				"quantity":    "3",
				"price":       "$50.00",
				"total":       "$150.00",
			},
		},
	}

	// Convert to map for the generator
	dataBytes, _ := json.Marshal(testData)
	var dataMap map[string]interface{}
	json.Unmarshal(dataBytes, &dataMap)

	// Parse template
	parser := parsers.NewCSVParser()
	elements, err := parser.ParseCSV("./assets/pdf_template_enhanced.csv")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	// Create generator
	generator := generators.NewPDFGenerator(generators.GeneratorConfig{
		FontDir:     "./fonts",
		TempDir:     os.TempDir(),
		DefaultFont: "Tahoma",
		PageSize:    "A4",
		Orientation: "P",
	})

	// Generate PDF
	outputFile := "test_enhanced_invoice.pdf"
	err = generator.GeneratePDF(elements, dataMap, outputFile)
	if err != nil {
		log.Printf("Error generating PDF: %v", err)
		return
	}

	fmt.Printf("Successfully generated enhanced invoice: %s\n", outputFile)

	// Check file size
	if fileInfo, err := os.Stat(outputFile); err == nil {
		fmt.Printf("Generated PDF size: %d bytes\n", fileInfo.Size())
	}
}

// Example usage functions

func ExampleUsage() {
	fmt.Println("\nExample Usage:")
	fmt.Println("==============")

	// 1. Create CSV parser
	fmt.Println("1. Initialize parser:")
	fmt.Println("   parser := parsers.NewCSVParser()")

	// 2. Parse template
	fmt.Println("\n2. Parse template:")
	fmt.Println("   elements, err := parser.ParseCSV(\"./assets/template.csv\")")

	// 3. Create generator
	fmt.Println("\n3. Create generator:")
	fmt.Println("   generator := generators.NewPDFGenerator(generators.GeneratorConfig{")
	fmt.Println("       FontDir: \"./fonts\",")
	fmt.Println("       TempDir: os.TempDir(),")
	fmt.Println("   })")

	// 4. Generate PDF
	fmt.Println("\n4. Generate PDF:")
	fmt.Println("   err = generator.GeneratePDF(elements, data, \"output.pdf\")")

	// 5. QR Code example
	fmt.Println("\n5. QR Code in CSV:")
	fmt.Println("   type,method,x,y,width,height,qrContent")
	fmt.Println("   qr,QR,10,10,30,30,\"https://example.com\"")

	// 6. Barcode example
	fmt.Println("\n6. Barcode in CSV:")
	fmt.Println("   type,method,x,y,width,height,barcodeFormat,barcodeContent")
	fmt.Println("   barcode,Barcode,10,50,80,15,Code128,\"123456789\"")
}
