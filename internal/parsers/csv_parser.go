package parsers

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"pdf-gen-simple/internal/cache"
	"pdf-gen-simple/internal/models"
	"pdf-gen-simple/internal/utils"
)

// CSVParser handles parsing CSV templates
type CSVParser struct {
	cache *cache.TemplateCache
}

// NewCSVParser creates a new CSV parser with caching
func NewCSVParser() *CSVParser {
	return &CSVParser{
		cache: cache.GetTemplateCache(),
	}
}

// ParseCSV parses a CSV template file and returns PDF elements
func (p *CSVParser) ParseCSV(filePath string) ([]models.PDFElement, error) {
	// Check cache first
	if elements, found := p.cache.Get(filePath); found {
		utils.LogDebug("CSV template loaded from cache: %s", filePath)
		return elements, nil
	}

	utils.LogInfo("Parsing CSV template: %s", filePath)

	// Open and parse CSV file
	elements, err := p.parseCSVFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file: %w", err)
	}

	// Cache the parsed elements
	p.cache.Set(filePath, elements)

	utils.LogInfo("Successfully parsed %d elements from CSV", len(elements))
	return elements, nil
}

// parseCSVFile performs the actual CSV parsing
func (p *CSVParser) parseCSVFile(filePath string) ([]models.PDFElement, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true // Performance optimization

	// Read header
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV headers: %w", err)
	}

	utils.LogDebug("CSV Headers: %v", headers)

	var elements []models.PDFElement
	rowIndex := 1

	// Process each row
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV row %d: %w", rowIndex+1, err)
		}

		rowIndex++

		if len(record) != len(headers) {
			utils.LogWarn("Row %d has incorrect number of columns (expected %d, got %d)",
				rowIndex, len(headers), len(record))
			continue
		}

		// Create element from row data
		element, err := p.createElementFromRow(headers, record, rowIndex)
		if err != nil {
			utils.LogError("Error creating element from row %d: %v", rowIndex, err)
			continue
		}

		// Validate element
		if err := element.Validate(); err != nil {
			utils.LogWarn("Invalid element at row %d: %v", rowIndex, err)
			continue
		}

		elements = append(elements, *element)
		utils.LogDebug("Created element from row %d: %s", rowIndex, element.Type)
	}

	return elements, nil
}

// createElementFromRow creates a PDFElement from CSV row data
func (p *CSVParser) createElementFromRow(headers, record []string, rowIndex int) (*models.PDFElement, error) {
	// Create a map for easier access
	data := make(map[string]string)
	for i, header := range headers {
		if i < len(record) {
			data[header] = record[i]
		}
	}

	// Parse basic properties
	element := &models.PDFElement{
		Type:         p.parseElementType(data["type"], data["method"]),
		Method:       data["method"],
		Text:         data["text"],
		VariableName: data["variableName"],
		LoopField:    data["loopField"],

		Position: models.Position{
			X: utils.ParseFloat(data["x"]),
			Y: utils.ParseFloat(data["y"]),
		},

		Size: models.Size{
			Width:  utils.ParseFloat(data["width"]),
			Height: utils.ParseFloat(data["height"]),
		},

		Style: models.Style{
			Font: models.Font{
				Family: utils.Coalesce(data["font"], "Tahoma"),
				Style:  data["fontStyle"],
				Size:   utils.ParseFloat(data["fontSize"]),
			},
			Border:       data["border"],
			Align:        utils.NormalizeAlign(data["align"]),
			RotateDegree: utils.ParseInt(data["rotateDegree"]),
			RotateType:   data["rotateType"],
			TextColor: models.Color{
				R:     utils.ParseInt(data["colorR"]),
				G:     utils.ParseInt(data["colorG"]),
				B:     utils.ParseInt(data["colorB"]),
				IsSet: data["colorR"] != "" || data["colorG"] != "" || data["colorB"] != "",
			},
			Background: models.Color{
				R:     utils.ParseInt(data["bgColorR"]),
				G:     utils.ParseInt(data["bgColorG"]),
				B:     utils.ParseInt(data["bgColorB"]),
				IsSet: data["background"] == "1",
			},
			ImageSrc: data["imageSrc"],
		},

		// QR/Barcode specific fields
		QRContent:      data["qrContent"],
		BarcodeFormat:  utils.Coalesce(data["barcodeFormat"], "Code128"),
		BarcodeContent: data["barcodeContent"],
	}

	// Set default font size if not specified
	if element.Style.Font.Size == 0 {
		element.Style.Font.Size = 10
	}

	// Parse table columns if present
	if columnsData := data["columns"]; columnsData != "" {
		columns, err := p.parseColumns(columnsData)
		if err != nil {
			utils.LogWarn("Error parsing columns for row %d: %v", rowIndex, err)
		} else {
			element.Columns = columns
		}
	}

	utils.LogDebug("Created element: Type=%s, Method=%s, Position=(%.1f,%.1f), Size=(%.1f,%.1f)",
		element.Type, element.Method, element.Position.X, element.Position.Y,
		element.Size.Width, element.Size.Height)

	return element, nil
}

// parseElementType determines the element type from type and method fields
func (p *CSVParser) parseElementType(typeField, methodField string) models.ElementType {
	// If type is explicitly set, use it
	if typeField != "" {
		switch strings.ToLower(typeField) {
		case "text":
			return models.ElementTypeText
		case "box":
			return models.ElementTypeBox
		case "image":
			return models.ElementTypeImage
		case "qr":
			return models.ElementTypeQR
		case "barcode":
			return models.ElementTypeBarcode
		case "table":
			return models.ElementTypeTable
		}
	}

	// Infer from method if type is not set
	switch methodField {
	case "MultiCell", "Cell":
		return models.ElementTypeText
	case "Rect":
		return models.ElementTypeBox
	case "Image":
		return models.ElementTypeImage
	case "QR":
		return models.ElementTypeQR
	case "Barcode":
		return models.ElementTypeBarcode
	default:
		return models.ElementTypeText // Default fallback
	}
}

// parseColumns parses column definitions from a string format
func (p *CSVParser) parseColumns(columnsData string) ([]models.TableColumn, error) {
	// This is a simple implementation. In a real scenario, you might want
	// to support JSON format or a more sophisticated parsing mechanism

	// For now, assume comma-separated format: "field1:width1:align1,field2:width2:align2"
	var columns []models.TableColumn

	parts := strings.Split(columnsData, ",")
	for _, part := range parts {
		columnParts := strings.Split(strings.TrimSpace(part), ":")
		if len(columnParts) >= 2 {
			column := models.TableColumn{
				Field: strings.TrimSpace(columnParts[0]),
				Width: utils.ParseFloat(columnParts[1]),
			}

			if len(columnParts) >= 3 {
				column.Align = utils.NormalizeAlign(columnParts[2])
			} else {
				column.Align = "L"
			}

			if len(columnParts) >= 4 {
				column.FontStyle = columnParts[3]
			}

			columns = append(columns, column)
		}
	}

	return columns, nil
}

// GetCacheStats returns statistics about the template cache
func (p *CSVParser) GetCacheStats() map[string]interface{} {
	return p.cache.Stats()
}

// ClearCache clears the template cache
func (p *CSVParser) ClearCache() {
	p.cache.Clear()
}

// ParseCSVFromReader parses CSV data from an io.Reader (for testing or dynamic content)
func (p *CSVParser) ParseCSVFromReader(reader io.Reader) ([]models.PDFElement, error) {
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.ReuseRecord = true

	// Read header
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV headers: %w", err)
	}

	var elements []models.PDFElement
	rowIndex := 1

	// Process each row
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV row %d: %w", rowIndex+1, err)
		}

		rowIndex++

		if len(record) != len(headers) {
			continue // Skip malformed rows
		}

		element, err := p.createElementFromRow(headers, record, rowIndex)
		if err != nil {
			continue // Skip invalid elements
		}

		if err := element.Validate(); err != nil {
			continue // Skip invalid elements
		}

		elements = append(elements, *element)
	}

	return elements, nil
}
