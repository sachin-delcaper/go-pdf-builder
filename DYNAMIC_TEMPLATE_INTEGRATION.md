# Dynamic Template Integration Guide

## Adding Dynamic Template Support to Your Existing Application

To add the new dynamic template endpoint `POST /invoice/template/{template_name}` to your existing `main.go`, follow these steps:

## 1. Add the New Route to main.go

In your existing `main.go` file, add these routes after your other endpoints:

```go
// Initialize handlers (if not already done)
csvHandler := handlers.NewCSVTemplateHandler()

// NEW: Dynamic template endpoints
r.POST("/invoice/template/:template_name", csvHandler.HandleDynamicTemplate)
r.GET("/invoice/template/:template_name", csvHandler.HandleTemplateInfo)

// Optional: Template listing endpoint
r.GET("/templates", func(c *gin.Context) {
    templates := []map[string]interface{}{
        {
            "name": "pdf_template_1",
            "filename": "pdf_template_1.csv",
            "endpoint": "/invoice/template/pdf_template_1",
        },
        {
            "name": "pdf_template_enhanced",
            "filename": "pdf_template_enhanced.csv", 
            "endpoint": "/invoice/template/pdf_template_enhanced",
        },
    }
    
    c.JSON(200, gin.H{
        "templates": templates,
        "count": len(templates),
        "usage": "Use template name (without .csv) in /invoice/template/:template_name",
    })
})
```

## 2. Import the New Handler

Make sure your imports include the handlers package:

```go
import (
    // ... your existing imports
    "pdf-gen-simple/internal/handlers"
)
```

## 3. Complete Integration Example

Here's how your route setup section should look:

```go
func main() {
    // ... existing middleware setup
    
    // Initialize handlers
    csvHandler := handlers.NewCSVTemplateHandler()
    
    // Existing endpoints
    r.POST("/invoice/template_csv", csvHandler.HandleCSVTemplate)
    
    // NEW: Dynamic template endpoints
    r.POST("/invoice/template/:template_name", csvHandler.HandleDynamicTemplate)
    r.GET("/invoice/template/:template_name", csvHandler.HandleTemplateInfo)
    r.GET("/templates", listTemplatesHandler)
    
    // Cache management
    r.GET("/cache/stats", csvHandler.HandleCacheStats)
    r.POST("/cache/clear", csvHandler.HandleCacheClear)
    
    // ... rest of your existing routes
}

func listTemplatesHandler(c *gin.Context) {
    templates := []map[string]interface{}{
        {
            "name": "pdf_template_1",
            "filename": "pdf_template_1.csv",
            "endpoint": "/invoice/template/pdf_template_1",
            "info_endpoint": "/invoice/template/pdf_template_1 (GET)",
        },
        {
            "name": "pdf_template_enhanced", 
            "filename": "pdf_template_enhanced.csv",
            "endpoint": "/invoice/template/pdf_template_enhanced",
            "info_endpoint": "/invoice/template/pdf_template_enhanced (GET)",
        },
    }
    
    c.JSON(200, gin.H{
        "templates": templates,
        "count": len(templates),
        "usage": "Use template name (without .csv) in /invoice/template/:template_name",
    })
}
```

## 4. Usage Examples

### Generate PDF with Dynamic Template
```bash
# Using pdf_template_enhanced
curl -X POST http://localhost:8080/invoice/template/pdf_template_enhanced \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "invoiceNumber": "INV-001",
      "customerName": "John Doe", 
      "qrData": "https://company.com/verify/INV-001",
      "barcodeData": "INV001"
    }
  }' \
  --output invoice_enhanced.pdf

# Using original template
curl -X POST http://localhost:8080/invoice/template/pdf_template_1 \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "invoiceNumber": "INV-002",
      "customerName": "Jane Smith"
    }
  }' \
  --output invoice_basic.pdf
```

### Get Template Information
```bash
# Get info about a template
curl http://localhost:8080/invoice/template/pdf_template_enhanced

# List all available templates
curl http://localhost:8080/templates
```

## 5. Template File Requirements

Your CSV templates must be:
1. Located in the `./assets/` directory
2. Have `.csv` extension
3. Follow the enhanced CSV format with support for QR/Barcode elements

Example template structure:
```
assets/
├── pdf_template_1.csv           # Basic template
├── pdf_template_enhanced.csv    # Enhanced with QR/Barcode  
├── custom_invoice.csv           # Custom template
└── receipt_template.csv         # Receipt template
```

## 6. API Response Examples

### Successful PDF Generation
```json
// Returns PDF file as binary data
// Headers:
// Content-Type: application/pdf
// Content-Disposition: attachment; filename=invoice_pdf_template_enhanced.pdf
```

### Template Information (GET)
```json
{
  "template": "pdf_template_enhanced",
  "path": "./assets/pdf_template_enhanced.csv",
  "file_info": {
    "size": 2048,
    "modified": "2024-01-15T10:30:00Z"
  },
  "elements": 28,
  "parse_error": "",
  "cache_stats": {
    "entries": 2,
    "maxSize": 100,
    "ttl": "30m0s"
  },
  "usage": {
    "method": "POST",
    "url": "/invoice/template/pdf_template_enhanced",
    "content_type": "application/json",
    "body_example": {
      "fields": {
        "invoiceNumber": "INV-001",
        "customerName": "John Doe",
        "qrData": "https://example.com",
        "barcodeData": "INV001"
      }
    }
  }
}
```

### Error Responses
```json
// Template not found
{
  "error": "Invalid template name or template not found",
  "template": "non_existent_template",
  "note": "Template must exist in assets directory and be a .csv file"
}

// Invalid request method
{
  "error": "Only POST method is allowed for PDF generation",
  "template": "pdf_template_enhanced",
  "usage": "POST /invoice/template/pdf_template_enhanced with JSON body containing 'fields'"
}
```

## 7. Security Features

The dynamic template endpoint includes several security features:

1. **Path Validation**: Prevents directory traversal attacks
2. **File Extension Validation**: Only allows .csv files
3. **Asset Directory Restriction**: Templates must be in ./assets/ directory
4. **Input Sanitization**: Template names are cleaned and validated

## 8. Performance Benefits

- **Template Caching**: Templates are cached after first parse
- **Path Building**: Efficient path construction with proper validation
- **Error Handling**: Graceful error responses with helpful messages
- **Memory Management**: Uses existing object pooling and caching systems

## 9. Testing the Integration

After adding the routes, test with:

```bash
# Test template listing
curl http://localhost:8080/templates

# Test template info
curl http://localhost:8080/invoice/template/pdf_template_enhanced

# Test PDF generation
curl -X POST http://localhost:8080/invoice/template/pdf_template_enhanced \
  -H "Content-Type: application/json" \
  -d '{"fields":{"invoiceNumber":"TEST-001"}}' \
  --output test.pdf
```

## 10. Troubleshooting

### Common Issues:
1. **404 Not Found**: Check template exists in ./assets/ directory
2. **Invalid Template**: Ensure CSV format is correct
3. **Parse Errors**: Check CSV syntax and required columns
4. **Missing Fields**: Verify JSON request includes required fields

### Debug Steps:
1. Check server logs for detailed error messages
2. Use GET endpoint to verify template information
3. Test with known working templates first
4. Validate CSV format against working examples

This integration provides a flexible, secure, and performant way to use different PDF templates dynamically! 