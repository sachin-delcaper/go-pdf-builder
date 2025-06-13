//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TestClient represents a simple HTTP client for testing
type TestClient struct {
	baseURL string
	client  *http.Client
}

// NewTestClient creates a new test client
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func main() {
	fmt.Println("🧪 Testing Dynamic Template API")
	fmt.Println("===============================")

	client := NewTestClient("http://localhost:8080")

	// Test 1: List available templates
	fmt.Println("\n📋 1. Testing template listing...")
	testTemplateList(client)

	// Test 2: Get template information
	fmt.Println("\n📄 2. Testing template information...")
	testTemplateInfo(client, "pdf_template_enhanced")

	// Test 3: Generate PDF with enhanced template
	fmt.Println("\n🎯 3. Testing PDF generation with enhanced template...")
	testPDFGeneration(client, "pdf_template_enhanced")

	// Test 4: Generate PDF with basic template
	fmt.Println("\n📝 4. Testing PDF generation with basic template...")
	testPDFGeneration(client, "pdf_template_1")

	// Test 5: Test error handling
	fmt.Println("\n❌ 5. Testing error handling...")
	testErrorHandling(client)

	fmt.Println("\n✅ All tests completed!")
}

func testTemplateList(client *TestClient) {
	resp, err := client.client.Get(client.baseURL + "/templates")
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ❌ Expected status 200, got %d\n", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   ❌ Error decoding response: %v\n", err)
		return
	}

	templates := result["templates"].([]interface{})
	fmt.Printf("   ✅ Found %d templates:\n", len(templates))
	for i, template := range templates {
		t := template.(map[string]interface{})
		fmt.Printf("      %d. %s (%s)\n", i+1, t["name"], t["endpoint"])
	}
}

func testTemplateInfo(client *TestClient, templateName string) {
	url := fmt.Sprintf("%s/invoice/template/%s", client.baseURL, templateName)
	resp, err := client.client.Get(url)
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ❌ Expected status 200, got %d\n", resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("   ❌ Error decoding response: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Template: %s\n", result["template"])
	fmt.Printf("   📊 Elements: %.0f\n", result["elements"])
	fmt.Printf("   📁 Path: %s\n", result["path"])

	if fileInfo, ok := result["file_info"].(map[string]interface{}); ok {
		fmt.Printf("   📏 Size: %.0f bytes\n", fileInfo["size"])
	}
}

func testPDFGeneration(client *TestClient, templateName string) {
	// Create test data
	testData := map[string]interface{}{
		"fields": map[string]interface{}{
			"invoiceNumber": "INV-2024-" + templateName,
			"date":          time.Now().Format("2006-01-02"),
			"customerName":  "Test Customer",
			"qrData":        "https://company.com/verify/" + templateName,
			"barcodeData":   "TEST" + templateName,
			"logoPath":      "./assets/smile-logo_small.png",
			"subtotal":      "$250.00",
			"tax":           "$25.00",
			"grandTotal":    "$275.00",
			"items": []map[string]interface{}{
				{
					"description": "Test Product A",
					"quantity":    "2",
					"price":       "$50.00",
					"total":       "$100.00",
				},
				{
					"description": "Test Product B",
					"quantity":    "3",
					"price":       "$50.00",
					"total":       "$150.00",
				},
			},
		},
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		fmt.Printf("   ❌ Error marshaling JSON: %v\n", err)
		return
	}

	url := fmt.Sprintf("%s/invoice/template/%s", client.baseURL, templateName)
	resp, err := client.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("   ❌ Expected status 200, got %d\n", resp.StatusCode)
		fmt.Printf("   📄 Response: %s\n", string(body))
		return
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		fmt.Printf("   ❌ Expected PDF content type, got: %s\n", contentType)
		return
	}

	// Read PDF data
	pdfData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ Error reading PDF data: %v\n", err)
		return
	}

	// Save PDF file
	filename := fmt.Sprintf("test_dynamic_%s.pdf", templateName)
	if err := os.WriteFile(filename, pdfData, 0644); err != nil {
		fmt.Printf("   ❌ Error saving PDF: %v\n", err)
		return
	}

	fmt.Printf("   ✅ Generated PDF: %s (%d bytes)\n", filename, len(pdfData))

	// Check PDF header
	if len(pdfData) >= 4 && string(pdfData[:4]) == "%PDF" {
		fmt.Printf("   ✅ Valid PDF format\n")
	} else {
		fmt.Printf("   ❌ Invalid PDF format\n")
	}
}

func testErrorHandling(client *TestClient) {
	// Test 1: Non-existent template
	fmt.Printf("   🔍 Testing non-existent template...\n")
	testNonExistentTemplate(client)

	// Test 2: Invalid method (GET for PDF generation)
	fmt.Printf("   🔍 Testing invalid method...\n")
	testInvalidMethod(client)

	// Test 3: Invalid JSON
	fmt.Printf("   🔍 Testing invalid JSON...\n")
	testInvalidJSON(client)
}

func testNonExistentTemplate(client *TestClient) {
	url := fmt.Sprintf("%s/invoice/template/non_existent_template", client.baseURL)

	testData := map[string]interface{}{
		"fields": map[string]interface{}{
			"invoiceNumber": "TEST-001",
		},
	}

	jsonData, _ := json.Marshal(testData)
	resp, err := client.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("      ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		fmt.Printf("      ✅ Correctly returned 400 for non-existent template\n")
	} else {
		fmt.Printf("      ❌ Expected 400, got %d\n", resp.StatusCode)
	}
}

func testInvalidMethod(client *TestClient) {
	url := fmt.Sprintf("%s/invoice/template/pdf_template_1", client.baseURL)

	// Try to use GET method for PDF generation (should fail)
	resp, err := client.client.Get(url)
	if err != nil {
		fmt.Printf("      ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// GET should return template info, not error
	if resp.StatusCode == 200 {
		fmt.Printf("      ✅ GET method correctly returns template info\n")
	} else {
		fmt.Printf("      ❌ Expected 200 for GET, got %d\n", resp.StatusCode)
	}
}

func testInvalidJSON(client *TestClient) {
	url := fmt.Sprintf("%s/invoice/template/pdf_template_1", client.baseURL)

	// Send invalid JSON
	invalidJSON := `{"fields": invalid json}`
	resp, err := client.client.Post(url, "application/json", bytes.NewBuffer([]byte(invalidJSON)))
	if err != nil {
		fmt.Printf("      ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		fmt.Printf("      ✅ Correctly returned 400 for invalid JSON\n")
	} else {
		fmt.Printf("      ❌ Expected 400, got %d\n", resp.StatusCode)
	}
}

// Example usage demonstrating the API
func printUsageExamples() {
	fmt.Println("\n📚 Usage Examples:")
	fmt.Println("==================")

	fmt.Println("\n1. List available templates:")
	fmt.Println("   curl http://localhost:8080/templates")

	fmt.Println("\n2. Get template information:")
	fmt.Println("   curl http://localhost:8080/invoice/template/pdf_template_enhanced")

	fmt.Println("\n3. Generate PDF with enhanced template:")
	fmt.Println(`   curl -X POST http://localhost:8080/invoice/template/pdf_template_enhanced \
     -H "Content-Type: application/json" \
     -d '{
       "fields": {
         "invoiceNumber": "INV-001",
         "customerName": "John Doe",
         "qrData": "https://company.com/verify/INV-001",
         "barcodeData": "INV001"
       }
     }' \
     --output invoice.pdf`)

	fmt.Println("\n4. Generate PDF with basic template:")
	fmt.Println(`   curl -X POST http://localhost:8080/invoice/template/pdf_template_1 \
     -H "Content-Type: application/json" \
     -d '{
       "fields": {
         "invoiceNumber": "INV-002",
         "customerName": "Jane Smith"
       }
     }' \
     --output basic_invoice.pdf`)
}
