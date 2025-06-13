package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"pdf-gen-simple/internal/generators"
	"pdf-gen-simple/internal/models"
	"pdf-gen-simple/internal/parsers"
	"pdf-gen-simple/internal/utils"
)

// CSVTemplateHandler handles CSV template-based PDF generation
type CSVTemplateHandler struct {
	parser    *parsers.CSVParser
	generator *generators.PDFGenerator
}

// NewCSVTemplateHandler creates a new CSV template handler
func NewCSVTemplateHandler() *CSVTemplateHandler {
	generator := generators.NewPDFGenerator(generators.GeneratorConfig{
		FontDir:     "./fonts",
		TempDir:     os.TempDir(),
		DefaultFont: "Tahoma",
		PageSize:    "A4",
		Orientation: "P",
	})

	return &CSVTemplateHandler{
		parser:    parsers.NewCSVParser(),
		generator: generator,
	}
}

// HandleCSVTemplate handles POST /invoice/template_csv
func (h *CSVTemplateHandler) HandleCSVTemplate(c *gin.Context) {
	utils.LogInfo("Received request for CSV template-based PDF generation")

	var req models.CSVTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	utils.LogDebug("Processing CSV template request with %d fields", len(req.Fields))

	// Parse CSV template
	templatePath := "./assets/pdf_template_1.csv"
	elements, err := h.parser.ParseCSV(templatePath)
	if err != nil {
		utils.LogError("Error parsing CSV template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse CSV template",
		})
		return
	}

	utils.LogInfo("Successfully parsed %d elements from CSV template", len(elements))

	// Generate PDF in memory
	pdfBytes, err := h.generator.GeneratePDFToBytes(elements, req.Fields)
	if err != nil {
		utils.LogError("Error generating PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "PDF generation failed",
		})
		return
	}

	utils.LogInfo("Successfully generated PDF of size: %d bytes", len(pdfBytes))

	// Set headers for PDF download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=invoice.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Return PDF as downloadable file
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// HandleCSVTemplateToFile handles POST /invoice/template_csv/file (saves to file)
func (h *CSVTemplateHandler) HandleCSVTemplateToFile(c *gin.Context) {
	utils.LogInfo("Received request for CSV template-based PDF generation (file output)")

	var req models.CSVTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Parse CSV template
	templatePath := "./assets/pdf_template_1.csv"
	elements, err := h.parser.ParseCSV(templatePath)
	if err != nil {
		utils.LogError("Error parsing CSV template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse CSV template",
		})
		return
	}

	// Generate PDF to file
	outputFile := filepath.Join(os.TempDir(), fmt.Sprintf("invoice_%d.pdf",
		c.Request.Context().Value("timestamp")))
	err = h.generator.GeneratePDF(elements, req.Fields, outputFile)
	if err != nil {
		utils.LogError("Error generating PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "PDF generation failed",
		})
		return
	}

	// Read generated PDF
	pdfBytes, err := os.ReadFile(outputFile)
	if err != nil {
		utils.LogError("Error reading generated PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read generated PDF",
		})
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
}

// HandleCacheStats handles GET /cache/stats
func (h *CSVTemplateHandler) HandleCacheStats(c *gin.Context) {
	stats := h.parser.GetCacheStats()
	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
	})
}

// HandleCacheClear handles POST /cache/clear
func (h *CSVTemplateHandler) HandleCacheClear(c *gin.Context) {
	h.parser.ClearCache()
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
	})
}

// HandleCustomTemplate handles POST /invoice/custom_template
func (h *CSVTemplateHandler) HandleCustomTemplate(c *gin.Context) {
	utils.LogInfo("Received request for custom template-based PDF generation")

	// Get template path from query parameter
	templatePath := c.Query("template")
	if templatePath == "" {
		templatePath = "./assets/pdf_template_1.csv" // Default template
	}

	var req models.CSVTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Invalid request format: %v", err),
		})
		return
	}

	// Validate template path (security check)
	if !h.isValidTemplatePath(templatePath) {
		utils.LogError("Invalid template path: %s", templatePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid template path",
		})
		return
	}

	// Parse CSV template
	elements, err := h.parser.ParseCSV(templatePath)
	if err != nil {
		utils.LogError("Error parsing CSV template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to parse CSV template",
		})
		return
	}

	// Generate PDF
	pdfBytes, err := h.generator.GeneratePDFToBytes(elements, req.Fields)
	if err != nil {
		utils.LogError("Error generating PDF: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "PDF generation failed",
		})
		return
	}

	utils.LogInfo("Successfully generated custom template PDF of size: %d bytes", len(pdfBytes))

	// Set headers for PDF download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename=custom_invoice.pdf")
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Return PDF
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// isValidTemplatePath validates that the template path is safe
func (h *CSVTemplateHandler) isValidTemplatePath(templatePath string) bool {
	// Only allow files in the assets directory
	absPath, err := filepath.Abs(templatePath)
	if err != nil {
		return false
	}

	assetsDir, err := filepath.Abs("./assets")
	if err != nil {
		return false
	}

	// Check if the path is within the assets directory
	relPath, err := filepath.Rel(assetsDir, absPath)
	if err != nil {
		return false
	}

	// Prevent directory traversal
	if strings.Contains(relPath, "..") {
		return false
	}

	// Check if file exists and is a CSV file
	if !strings.HasSuffix(strings.ToLower(templatePath), ".csv") {
		return false
	}

	_, err = os.Stat(templatePath)
	return err == nil
}

// HandleDynamicTemplate handles GET/POST /invoice/template/:template_name
func (h *CSVTemplateHandler) HandleDynamicTemplate(c *gin.Context) {
	templateName := c.Param("template_name")
	utils.LogInfo("Received request for dynamic template: %s", templateName)

	// Only allow POST method for PDF generation
	if c.Request.Method != "POST" {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":    "Only POST method is allowed for PDF generation",
			"template": templateName,
			"usage":    "POST /invoice/template/" + templateName + " with JSON body containing 'fields'",
		})
		return
	}

	// Validate template name
	if templateName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Template name is required",
			"usage": "POST /invoice/template/{template_name}",
		})
		return
	}

	// Construct full template path
	templatePath := h.buildTemplatePath(templateName)

	// Validate template path for security
	if !h.isValidTemplatePath(templatePath) {
		utils.LogError("Invalid or unsafe template path: %s", templatePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Invalid template name or template not found",
			"template": templateName,
			"note":     "Template must exist in assets directory and be a .csv file",
		})
		return
	}

	// Parse request body
	var req models.CSVTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    fmt.Sprintf("Invalid request format: %v", err),
			"template": templateName,
		})
		return
	}

	utils.LogDebug("Processing dynamic template request for %s with %d fields", templateName, len(req.Fields))

	// Parse CSV template
	elements, err := h.parser.ParseCSV(templatePath)
	if err != nil {
		utils.LogError("Error parsing CSV template %s: %v", templatePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to parse CSV template",
			"template": templateName,
			"details":  err.Error(),
		})
		return
	}

	utils.LogInfo("Successfully parsed %d elements from template: %s", len(elements), templateName)

	// Generate PDF in memory
	pdfBytes, err := h.generator.GeneratePDFToBytes(elements, req.Fields)
	if err != nil {
		utils.LogError("Error generating PDF for template %s: %v", templateName, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "PDF generation failed",
			"template": templateName,
			"details":  err.Error(),
		})
		return
	}

	utils.LogInfo("Successfully generated PDF from template %s, size: %d bytes", templateName, len(pdfBytes))

	// Set headers for PDF download
	filename := fmt.Sprintf("invoice_%s.pdf", templateName)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Return PDF as downloadable file
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// HandleTemplateInfo handles GET /invoice/template/:template_name (for template info)
func (h *CSVTemplateHandler) HandleTemplateInfo(c *gin.Context) {
	templateName := c.Param("template_name")
	utils.LogInfo("Received template info request for: %s", templateName)

	// Construct full template path
	templatePath := h.buildTemplatePath(templateName)

	// Validate template path
	if !h.isValidTemplatePath(templatePath) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":    "Template not found",
			"template": templateName,
		})
		return
	}

	// Get template info without full parsing (for performance)
	fileInfo, err := os.Stat(templatePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":    "Template file not accessible",
			"template": templateName,
		})
		return
	}

	// Try to parse template to get element count
	elements, err := h.parser.ParseCSV(templatePath)
	elementCount := 0
	var parseError string
	if err != nil {
		parseError = err.Error()
	} else {
		elementCount = len(elements)
	}

	// Get cache stats for this template
	cacheStats := h.parser.GetCacheStats()

	c.JSON(http.StatusOK, gin.H{
		"template": templateName,
		"path":     templatePath,
		"file_info": gin.H{
			"size":     fileInfo.Size(),
			"modified": fileInfo.ModTime(),
		},
		"elements":    elementCount,
		"parse_error": parseError,
		"cache_stats": cacheStats,
		"usage": gin.H{
			"method":       "POST",
			"url":          fmt.Sprintf("/invoice/template/%s", templateName),
			"content_type": "application/json",
			"body_example": gin.H{
				"fields": gin.H{
					"invoiceNumber": "INV-001",
					"customerName":  "John Doe",
					"qrData":        "https://example.com",
					"barcodeData":   "INV001",
				},
			},
		},
	})
}

// buildTemplatePath constructs the full path to a template file
func (h *CSVTemplateHandler) buildTemplatePath(templateName string) string {
	// Clean the template name
	templateName = strings.TrimSpace(templateName)

	// Remove any existing .csv extension to avoid double extension
	templateName = strings.TrimSuffix(templateName, ".csv")

	// Add .csv extension
	filename := templateName + ".csv"

	// Construct full path
	return filepath.Join("./assets", filename)
}
