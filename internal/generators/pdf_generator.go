package generators

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/qr"
	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"

	"pdf-gen-simple/internal/cache"
	"pdf-gen-simple/internal/models"
	"pdf-gen-simple/internal/utils"
)

// PDFGenerator handles PDF generation with enhanced features
type PDFGenerator struct {
	fontCache      *cache.FontCache
	tempDir        string
	pdfPool        sync.Pool
	lastYPositions map[string]float64
	mu             sync.RWMutex
}

// GeneratorConfig contains configuration for the PDF generator
type GeneratorConfig struct {
	FontDir     string
	TempDir     string
	DefaultFont string
	PageSize    string
	Orientation string
}

// NewPDFGenerator creates a new PDF generator with configuration
func NewPDFGenerator(config GeneratorConfig) *PDFGenerator {
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	if config.DefaultFont == "" {
		config.DefaultFont = "Tahoma"
	}
	if config.PageSize == "" {
		config.PageSize = "A4"
	}
	if config.Orientation == "" {
		config.Orientation = "P"
	}

	generator := &PDFGenerator{
		fontCache:      cache.GetFontCache(),
		tempDir:        config.TempDir,
		lastYPositions: make(map[string]float64),
	}

	// Initialize PDF object pool for better performance
	generator.pdfPool = sync.Pool{
		New: func() interface{} {
			pdf := fpdf.New(config.Orientation, "mm", config.PageSize, config.FontDir)
			generator.setupFonts(pdf)
			return pdf
		},
	}

	return generator
}

// GeneratePDF generates a PDF from elements and data
func (g *PDFGenerator) GeneratePDF(elements []models.PDFElement, data map[string]interface{}, outputFile string) error {
	// Get PDF instance from pool
	pdf := g.pdfPool.Get().(*fpdf.Fpdf)
	defer func() {
		// Reset PDF for reuse
		pdf = fpdf.New("P", "mm", "A4", "./fonts")
		g.setupFonts(pdf)
		g.pdfPool.Put(pdf)
	}()

	pdf.AddPage()
	g.setupFonts(pdf)

	utils.LogInfo("Generating PDF with %d elements", len(elements))

	// Process elements
	for i, element := range elements {
		utils.LogDebug("Processing element %d: %s", i+1, element.Type)

		if err := g.processElement(pdf, element, data); err != nil {
			utils.LogError("Error processing element %d: %v", i+1, err)
			continue
		}
	}

	// Save PDF
	utils.LogInfo("Saving PDF to: %s", outputFile)
	return pdf.OutputFileAndClose(outputFile)
}

// GeneratePDFToBytes generates a PDF and returns it as bytes
func (g *PDFGenerator) GeneratePDFToBytes(elements []models.PDFElement, data map[string]interface{}) ([]byte, error) {
	// Get PDF instance from pool
	pdf := g.pdfPool.Get().(*fpdf.Fpdf)
	defer g.pdfPool.Put(pdf)

	pdf.AddPage()
	g.setupFonts(pdf)

	// Process elements
	for _, element := range elements {
		if err := g.processElement(pdf, element, data); err != nil {
			utils.LogError("Error processing element: %v", err)
			continue
		}
	}

	// Output to bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

// setupFonts sets up the fonts for the PDF
func (g *PDFGenerator) setupFonts(pdf *fpdf.Fpdf) {
	if g.fontCache.IsSystemLoaded() {
		return
	}

	// Add UTF8 fonts
	fonts := map[string]string{
		"Tahoma":  "tahoma.ttf",
		"TahomaB": "tahomabd.TTF",
	}

	for fontName, fontFile := range fonts {
		if !g.fontCache.IsLoaded(fontName) {
			fontPath := filepath.Join("./fonts", fontFile)
			if _, err := os.Stat(fontPath); err == nil {
				if fontName == "TahomaB" {
					pdf.AddUTF8Font("Tahoma", "B", fontFile)
				} else {
					pdf.AddUTF8Font(fontName, "", fontFile)
				}
				g.fontCache.MarkLoaded(fontName)
				utils.LogDebug("Loaded font: %s", fontName)
			}
		}
	}

	pdf.SetFont("Tahoma", "", 10)
	g.fontCache.MarkSystemLoaded()
}

// processElement processes a single PDF element
func (g *PDFGenerator) processElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	// Validate element
	if err := element.Validate(); err != nil {
		return fmt.Errorf("element validation failed: %w", err)
	}

	// Handle loop elements
	if element.IsLoopElement() {
		return g.processLoopElement(pdf, element, data)
	}

	// Process based on element type
	switch element.Type {
	case models.ElementTypeText:
		return g.processTextElement(pdf, element, data)
	case models.ElementTypeBox:
		return g.processBoxElement(pdf, element, data)
	case models.ElementTypeImage:
		return g.processImageElement(pdf, element, data)
	case models.ElementTypeQR:
		return g.processQRElement(pdf, element, data)
	case models.ElementTypeBarcode:
		return g.processBarcodeElement(pdf, element, data)
	case models.ElementTypeTable:
		return g.processTableElement(pdf, element, data)
	default:
		return fmt.Errorf("unsupported element type: %s", element.Type)
	}
}

// processLoopElement processes elements that should be repeated for array data
func (g *PDFGenerator) processLoopElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	parts := strings.Split(element.LoopField, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid loopField format: %s", element.LoopField)
	}

	arrayName := parts[0]
	fieldName := parts[1]

	arrayData, ok := data[arrayName]
	if !ok {
		return fmt.Errorf("array field not found: %s", arrayName)
	}

	items, isArray := arrayData.([]interface{})
	if !isArray {
		return fmt.Errorf("field is not an array: %s", arrayName)
	}

	currentY := element.Position.Y
	spacing := element.Size.Height + 2 // Add small spacing between items

	for _, item := range items {
		// Create a copy of the element for this iteration
		elementCopy := element.Clone()
		elementCopy.Position.Y = currentY

		// Get field value from item
		itemValue := utils.GetArrayFieldValue(item, fieldName)

		// Process the element with the item value
		itemData := make(map[string]interface{})
		for k, v := range data {
			itemData[k] = v
		}
		itemData[element.LoopField] = itemValue

		if err := g.processElement(pdf, *elementCopy, itemData); err != nil {
			utils.LogError("Error processing loop element: %v", err)
		}

		currentY += spacing
	}

	// Update the last Y position for this array
	g.mu.Lock()
	g.lastYPositions[arrayName] = currentY
	g.mu.Unlock()

	return nil
}

// processTextElement processes text elements
func (g *PDFGenerator) processTextElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	// Set font
	g.setFont(pdf, element.Style.Font)

	// Set text color
	if element.Style.TextColor.IsSet {
		pdf.SetTextColor(element.Style.TextColor.R, element.Style.TextColor.G, element.Style.TextColor.B)
	}

	// Get text content with variable replacement
	text := g.replaceVariables(element.Text, element.VariableName, data)

	// Apply rotation if needed
	if element.Style.RotateDegree != 0 {
		rotateX, rotateY := g.calculateRotationPoint(element)
		pdf.TransformBegin()
		pdf.TransformRotate(float64(element.Style.RotateDegree), rotateX, rotateY)
	}

	// Draw text based on method
	pdf.SetXY(element.Position.X, element.Position.Y)

	switch element.Method {
	case "MultiCell":
		lineHeight := element.Style.Font.Size * 0.5
		pdf.MultiCell(element.Size.Width, lineHeight, text, element.Style.Border, element.Style.Align, false)
	case "Cell":
		pdf.CellFormat(element.Size.Width, element.Size.Height, text, element.Style.Border, 0, element.Style.Align, false, 0, "")
	default:
		pdf.CellFormat(element.Size.Width, element.Size.Height, text, element.Style.Border, 0, element.Style.Align, false, 0, "")
	}

	// End rotation if applied
	if element.Style.RotateDegree != 0 {
		pdf.TransformEnd()
	}

	return nil
}

// processBoxElement processes box/rectangle elements
func (g *PDFGenerator) processBoxElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	// Set border color
	if element.Style.TextColor.IsSet {
		pdf.SetDrawColor(element.Style.TextColor.R, element.Style.TextColor.G, element.Style.TextColor.B)
	}

	// Set fill color
	if element.Style.Background.IsSet {
		pdf.SetFillColor(element.Style.Background.R, element.Style.Background.G, element.Style.Background.B)
	}

	// Set line width
	pdf.SetLineWidth(0.2)

	// Draw rectangle
	if element.Style.Background.IsSet {
		pdf.Rect(element.Position.X, element.Position.Y, element.Size.Width, element.Size.Height, "FD")
	} else {
		pdf.Rect(element.Position.X, element.Position.Y, element.Size.Width, element.Size.Height, "D")
	}

	return nil
}

// processImageElement processes image elements
func (g *PDFGenerator) processImageElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	imagePath := element.Style.ImageSrc

	// Check if image path is a variable
	if imagePath == "" && element.VariableName != "" {
		if val, ok := data[element.VariableName]; ok {
			imagePath = fmt.Sprintf("%v", val)
		}
	}

	if imagePath == "" {
		return fmt.Errorf("image path not specified")
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); err != nil {
		return fmt.Errorf("image file not found: %s", imagePath)
	}

	// Add image to PDF
	pdf.Image(imagePath, element.Position.X, element.Position.Y, element.Size.Width, element.Size.Height, false, "", 0, "")

	return nil
}

// processQRElement processes QR code elements
func (g *PDFGenerator) processQRElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	// Get QR content
	content := element.GetTextContent(data)
	if content == "" {
		content = g.replaceVariables(element.QRContent, element.VariableName, data)
	}

	if content == "" {
		return fmt.Errorf("QR content is empty")
	}

	// Generate QR code
	qrCode, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	// Save QR code to temporary file
	tempFile := filepath.Join(g.tempDir, fmt.Sprintf("qr_%d.png", time.Now().UnixNano()))
	if err := os.WriteFile(tempFile, qrCode, 0644); err != nil {
		return fmt.Errorf("failed to save QR code: %w", err)
	}
	defer os.Remove(tempFile) // Clean up

	// Add QR code to PDF
	pdf.Image(tempFile, element.Position.X, element.Position.Y, element.Size.Width, element.Size.Height, false, "", 0, "")

	utils.LogDebug("Generated QR code for content: %s", utils.TruncateString(content, 50))
	return nil
}

// processBarcodeElement processes barcode elements
func (g *PDFGenerator) processBarcodeElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	// Get barcode content
	content := element.GetTextContent(data)
	if content == "" {
		content = g.replaceVariables(element.BarcodeContent, element.VariableName, data)
	}

	if content == "" {
		return fmt.Errorf("barcode content is empty")
	}

	// Generate barcode based on format
	var barcodeImg barcode.Barcode
	var err error

	switch strings.ToUpper(element.BarcodeFormat) {
	case "CODE128":
		barcodeImg, err = code128.Encode(content)
	case "CODE39":
		barcodeImg, err = code39.Encode(content, true, true)
	case "EAN13":
		barcodeImg, err = ean.Encode(content)
	case "QR":
		barcodeImg, err = qr.Encode(content, qr.M, qr.Auto)
	default:
		barcodeImg, err = code128.Encode(content) // Default to Code128
	}

	if err != nil {
		return fmt.Errorf("failed to generate barcode: %w", err)
	}

	// Scale barcode to desired size
	scaledBarcode, err := barcode.Scale(barcodeImg, int(element.Size.Width*10), int(element.Size.Height*10))
	if err != nil {
		return fmt.Errorf("failed to scale barcode: %w", err)
	}

	// Convert to PNG and save to temporary file
	tempFile := filepath.Join(g.tempDir, fmt.Sprintf("barcode_%d.png", time.Now().UnixNano()))

	// Create a buffer for the PNG data
	var buf bytes.Buffer
	if err := g.imageToPNG(scaledBarcode, &buf); err != nil {
		return fmt.Errorf("failed to convert barcode to PNG: %w", err)
	}

	if err := os.WriteFile(tempFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to save barcode: %w", err)
	}
	defer os.Remove(tempFile) // Clean up

	// Add barcode to PDF
	pdf.Image(tempFile, element.Position.X, element.Position.Y, element.Size.Width, element.Size.Height, false, "", 0, "")

	utils.LogDebug("Generated %s barcode for content: %s", element.BarcodeFormat, utils.TruncateString(content, 50))
	return nil
}

// processTableElement processes table elements
func (g *PDFGenerator) processTableElement(pdf *fpdf.Fpdf, element models.PDFElement, data map[string]interface{}) error {
	// This is a placeholder for table processing
	// Tables are complex and would need additional implementation
	utils.LogWarn("Table elements are not yet fully implemented")
	return nil
}

// Helper methods

// setFont sets the font for the PDF
func (g *PDFGenerator) setFont(pdf *fpdf.Fpdf, font models.Font) {
	family := font.Family
	if family == "" {
		family = "Tahoma"
	}

	style := font.Style
	size := font.Size
	if size == 0 {
		size = 10
	}

	pdf.SetFont(family, style, size)
}

// replaceVariables replaces variables in text
func (g *PDFGenerator) replaceVariables(text, variableName string, data map[string]interface{}) string {
	if variableName != "" {
		return utils.ReplaceVariablesInArray(text, variableName, data)
	}
	return utils.ReplaceVariables(text, data)
}

// calculateRotationPoint calculates the rotation point based on rotation type
func (g *PDFGenerator) calculateRotationPoint(element models.PDFElement) (float64, float64) {
	switch element.Style.RotateType {
	case "left":
		return element.Position.X, element.Position.Y + (element.Size.Height / 2)
	case "top":
		return element.Position.X + (element.Size.Width / 2), element.Position.Y
	default:
		return element.Position.X + (element.Size.Width / 2), element.Position.Y + (element.Size.Height / 2)
	}
}

// imageToPNG converts an image to PNG format
func (g *PDFGenerator) imageToPNG(img image.Image, buf *bytes.Buffer) error {
	return png.Encode(buf, img)
}
