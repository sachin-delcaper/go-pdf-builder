package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
)

// CSVTemplateRequest represents the JSON input for the CSV template endpoint
type CSVTemplateRequest struct {
	Fields map[string]interface{} `json:"fields"` // Changed to interface{} to support arrays
}

type PDFElement struct {
	Type         string
	Method       string
	Text         string
	VariableName string
	X            float64
	Y            float64
	Width        float64
	Height       float64
	Font         string
	FontStyle    string
	FontSize     float64
	Align        string
	RotateDegree int
	Border       string
	ColorR       int
	ColorG       int
	ColorB       int
	background   string
	BGColorR     int
	BGColorG     int
	BGColorB     int
	ImageSrc     string
	RotateType   string
	Columns      []TableColumn
	LoopField    string // New field for specifying array and field to loop
}

// TableColumn represents a column in a table
type TableColumn struct {
	Field     string  // Field name in the data object
	Width     float64 // Column width
	Align     string  // Column alignment
	FontStyle string  // Column font style
}

// Add a global map to track Y positions
var lastYPositions = make(map[string]float64)

func ParseCSV(path string) ([]PDFElement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	var elements []PDFElement
	headers := records[0]
	log.Printf("CSV Headers: %v", headers)

	for i, row := range records[1:] {
		if len(row) != len(headers) {
			return nil, fmt.Errorf("row %d has incorrect number of columns", i+2)
		}

		data := make(map[string]string)
		for j, val := range row {
			data[headers[j]] = val
			log.Printf("Row %d, Column %s: %s", i+2, headers[j], val)
		}

		// Parse coordinates and dimensions
		x := parseFloat(data["x"])
		y := parseFloat(data["y"])
		width := parseFloat(data["width"])
		height := parseFloat(data["height"])
		fontSize := parseFloat(data["fontSize"])

		// Determine element type based on method
		elementType := data["type"]
		if elementType == "" {
			switch data["method"] {
			case "MultiCell", "Cell":
				elementType = "text"
			case "Rect":
				elementType = "box"
			case "Image":
				elementType = "image"
			}
		}

		log.Printf("Parsing element %d:", i+1)
		log.Printf("  Type: %s", elementType)
		log.Printf("  Method: %s", data["method"])
		log.Printf("  Text: %s", data["text"])
		log.Printf("  Position: (%.1f, %.1f)", x, y)
		log.Printf("  Size: (%.1f, %.1f)", width, height)
		log.Printf("  Font: %s, Style: %s, Size: %.1f", data["font"], data["fontStyle"], fontSize)
		log.Printf("  Border: %s", data["border"])
		log.Printf("  Colors: R=%s, G=%s, B=%s", data["ColorR"], data["ColorG"], data["ColorB"])
		log.Printf("  LoopField: %s", data["loopField"])

		e := PDFElement{
			Type:         elementType,
			Method:       data["method"],
			Text:         data["text"],
			VariableName: data["variableName"],
			X:            x,
			Y:            y,
			Width:        width,
			Height:       height,
			Font:         data["font"],
			FontStyle:    data["fontStyle"],
			FontSize:     fontSize,
			Align:        data["align"],
			RotateDegree: parseInt(data["RotateDegree"]),
			Border:       data["border"],
			ColorR:       parseInt(data["ColorR"]),
			ColorG:       parseInt(data["ColorG"]),
			ColorB:       parseInt(data["ColorB"]),
			background:   data["background"],
			BGColorR:     parseInt(data["BGColorR"]),
			BGColorG:     parseInt(data["BGColorG"]),
			BGColorB:     parseInt(data["BGColorB"]),
			ImageSrc:     data["imageSrc"],
			RotateType:   data["rotateType"],
			LoopField:    data["loopField"],
		}
		log.Printf("Created element: %+v", e)
		elements = append(elements, e)
	}

	log.Printf("Successfully parsed %d elements from CSV", len(elements))
	return elements, nil
}

func GeneratePDF(elements []PDFElement, input map[string]interface{}, outputFile string) error {
	pdf := fpdf.New("P", "mm", "A4", "./fonts")
	pdf.AddPage()

	// Add fonts
	pdf.AddUTF8Font("Tahoma", "", "tahoma.ttf")
	pdf.AddUTF8Font("Tahoma", "B", "tahomabd.TTF")

	// Set default font
	pdf.SetFont("Tahoma", "", 10)

	log.Printf("Generating PDF with %d elements from CSV", len(elements))
	log.Printf("Input fields: %+v", input)

	// Process actual elements
	for i, el := range elements {
		log.Printf("\nProcessing CSV element %d:", i+1)
		log.Printf("  Type: %s", el.Type)
		log.Printf("  Method: %s", el.Method)
		log.Printf("  Text: %s", el.Text)
		log.Printf("  VariableName: %s", el.VariableName)
		log.Printf("  Position: (%.1f, %.1f)", el.X, el.Y)
		log.Printf("  Size: (%.1f, %.1f)", el.Width, el.Height)
		log.Printf("  Font: %s, Style: %s, Size: %.1f", el.Font, el.FontStyle, el.FontSize)
		log.Printf("  Align: %s", el.Align)
		log.Printf("  Border: %s", el.Border)
		log.Printf("  Background: %s", el.background)
		log.Printf("  Colors: R=%d, G=%d, B=%d", el.ColorR, el.ColorG, el.ColorB)
		log.Printf("  BG Colors: R=%d, G=%d, B=%d", el.BGColorR, el.BGColorG, el.BGColorB)

		// Skip empty elements
		if el.Type == "" || el.Method == "" {
			log.Printf("  Skipping empty element")
			continue
		}

		// Validate coordinates
		if el.X < 0 || el.Y < 0 {
			log.Printf("  Invalid coordinates: (%.1f, %.1f)", el.X, el.Y)
			continue
		}

		// Validate dimensions
		if el.Width <= 0 || el.Height <= 0 {
			log.Printf("  Invalid dimensions: (%.1f, %.1f)", el.Width, el.Height)
			continue
		}

		processElement(pdf, el, input)
	}

	log.Printf("Saving PDF to: %s", outputFile)
	return pdf.OutputFileAndClose(outputFile)
}

func processElement(pdf *fpdf.Fpdf, el PDFElement, input map[string]interface{}) {
	// Handle text content
	text := el.Text

	// Process based on element type
	switch el.Type {
	case "table", "text":
		// Check if we need to loop through an array
		if el.LoopField != "" {
			// Split into array name and field name
			parts := strings.Split(el.LoopField, ".")
			if len(parts) != 2 {
				log.Printf("Invalid loopField format: %s", el.LoopField)
				return
			}

			arrayName := parts[0]
			fieldName := parts[1]

			if data, ok := input[arrayName]; ok {
				if items, isArray := data.([]interface{}); isArray {
					// Get current Y position
					currentY := el.Y
					lineHeight := el.FontSize * 0.5

					// Process each array item
					for _, item := range items {
						// Get the field value if it's a map
						var itemStr string
						if itemMap, isMap := item.(map[string]interface{}); isMap {
							if val, ok := itemMap[fieldName]; ok {
								itemStr = fmt.Sprintf("%v", val)
							}
						}

						// Process any variables in the text
						if el.Text != "" {
							// Replace the field value in the text
							itemStr = strings.ReplaceAll(el.Text, "{{"+el.LoopField+"}}", itemStr)

							// Process any other variables in the text
							for varName, varVal := range input {
								if varName != arrayName { // Skip the array variable itself
									itemStr = strings.ReplaceAll(itemStr, "{{"+varName+"}}", fmt.Sprintf("%v", varVal))
								}
							}
						}

						// Set position for this item
						pdf.SetXY(el.X, currentY)

						// Set font and style for text elements
						if el.Font != "" {
							fontStyle := ""
							if el.FontStyle == "B" {
								fontStyle = "B"
							}
							pdf.SetFont(el.Font, fontStyle, el.FontSize)
						}

						// Set text color
						pdf.SetTextColor(el.ColorR, el.ColorG, el.ColorB)

						// Calculate rotation point based on rotation type
						var rotateX, rotateY float64
						switch el.RotateType {
						case "left":
							rotateX = el.X
							rotateY = currentY + (el.Height / 2)
						case "top":
							rotateX = el.X + (el.Width / 2)
							rotateY = currentY
						default:
							rotateX = el.X + (el.Width / 2)
							rotateY = currentY + (el.Height / 2)
						}

						// Apply rotation if specified
						if el.RotateDegree != 0 {
							pdf.TransformBegin()
							pdf.TransformRotate(float64(el.RotateDegree), rotateX, rotateY)
						}

						// Draw the text
						if el.Method == "MultiCell" {
							pdf.MultiCell(el.Width, lineHeight, itemStr, el.Border, el.Align, false)
							// Update Y position for next item
							currentY = pdf.GetY() + 2 // Add some spacing between items
						} else {
							pdf.CellFormat(el.Width, el.Height, itemStr, el.Border, 0, el.Align, false, 0, "")
							// Update Y position for next item
							currentY += el.Height + 2 // Add some spacing between items
						}

						// End rotation if applied
						if el.RotateDegree != 0 {
							pdf.TransformEnd()
						}

						log.Printf("Processed array item field %s: %s", fieldName, itemStr)
					}

					// Store the final Y position for this variable
					lastYPositions[arrayName] = currentY
					return
				}
			}
		}

		// Normal text processing for non-array variables
		if el.VariableName != "" {
			// Check if variableName is a JSON array
			if strings.HasPrefix(el.VariableName, "[") && strings.HasSuffix(el.VariableName, "]") {
				// Remove brackets and split by comma
				varsStr := strings.Trim(el.VariableName, "[]")
				vars := strings.Split(varsStr, ",")

				// Clean up variable names (remove quotes and spaces)
				for i, v := range vars {
					vars[i] = strings.Trim(strings.Trim(v, "\""), " ")
				}

				// Replace each variable in the text
				for _, varName := range vars {
					// Try exact match first
					if val, ok := input[varName]; ok {
						text = strings.ReplaceAll(text, "{{"+varName+"}}", fmt.Sprintf("%v", val))
						log.Printf("Replaced variable %s with value: %v", varName, val)
						continue
					}

					// Try case-insensitive match
					for inputKey, val := range input {
						// Remove any special characters from input key
						cleanInputKey := strings.TrimRight(inputKey, ":")
						if strings.EqualFold(cleanInputKey, varName) {
							text = strings.ReplaceAll(text, "{{"+varName+"}}", fmt.Sprintf("%v", val))
							log.Printf("Replaced variable %s with value: %v (matched with %s)", varName, val, inputKey)
							break
						}
					}
				}
			} else {
				// Handle single variable
				if val, ok := input[el.VariableName]; ok {
					text = strings.ReplaceAll(text, "{{"+el.VariableName+"}}", fmt.Sprintf("%v", val))
					log.Printf("Replaced variable %s with value: %v", el.VariableName, val)
				} else {
					log.Printf("Warning: Variable %s not found in input", el.VariableName)
				}
			}
		}

		// Set font and style for text elements
		if el.Font != "" {
			fontStyle := ""
			if el.FontStyle == "B" {
				fontStyle = "B"
			}
			pdf.SetFont(el.Font, fontStyle, el.FontSize)
			log.Printf("Set font: %s, style: %s, size: %.1f", el.Font, fontStyle, el.FontSize)
		}

		// Set text color
		pdf.SetTextColor(el.ColorR, el.ColorG, el.ColorB)
		log.Printf("Set text color: R=%d, G=%d, B=%d", el.ColorR, el.ColorG, el.ColorB)

		// Calculate rotation point based on rotation type
		var rotateX, rotateY float64
		switch el.RotateType {
		case "left":
			// Rotate from left side
			rotateX = el.X
			rotateY = el.Y + (el.Height / 2)
			log.Printf("Using left side rotation point: (%.1f, %.1f)", rotateX, rotateY)
		case "top":
			// Rotate from top middle
			rotateX = el.X + (el.Width / 2)
			rotateY = el.Y
			log.Printf("Using top rotation point: (%.1f, %.1f)", rotateX, rotateY)
		default:
			// Default to center rotation
			rotateX = el.X + (el.Width / 2)
			rotateY = el.Y + (el.Height / 2)
			log.Printf("Using center rotation point: (%.1f, %.1f)", rotateX, rotateY)
		}

		// Apply rotation if specified
		if el.RotateDegree != 0 {
			pdf.TransformBegin()
			pdf.TransformRotate(float64(el.RotateDegree), rotateX, rotateY)
			log.Printf("Applied rotation: %d degrees at point (%.1f, %.1f)", el.RotateDegree, rotateX, rotateY)
		}

		// Draw text based on method
		switch el.Method {
		case "MultiCell":
			log.Printf("Drawing MultiCell text at (%.1f, %.1f): %s", el.X, el.Y, text)
			align := "L" // Default to left alignment
			if el.Align != "" {
				align = el.Align
			}
			pdf.SetXY(el.X, el.Y)
			lineHeight := el.FontSize * 0.5
			pdf.MultiCell(el.Width, lineHeight, text, el.Border, align, false)

		case "Cell":
			log.Printf("Drawing Cell text at (%.1f, %.1f): %s", el.X, el.Y, text)
			align := "L" // Default to left alignment
			if el.Align != "" {
				align = el.Align
			}
			pdf.SetXY(el.X, el.Y)
			pdf.CellFormat(el.Width, el.Height, text, el.Border, 0, align, false, 0, "")
		}

		// End rotation if applied
		if el.RotateDegree != 0 {
			pdf.TransformEnd()
		}

	case "box":
		// For boxes, set border and background colors
		pdf.SetDrawColor(el.ColorR, el.ColorG, el.ColorB)

		// Set background color if specified
		if el.background == "1" {
			pdf.SetFillColor(el.BGColorR, el.BGColorG, el.BGColorB)
			log.Printf("Drawing box with background color: R=%d, G=%d, B=%d", el.BGColorR, el.BGColorG, el.BGColorB)
		} else {
			// No background fill
			pdf.SetFillColor(255, 255, 255) // White background
		}

		// Check if we need to loop through an array
		if el.LoopField != "" {
			parts := strings.Split(el.LoopField, ".")
			if len(parts) != 2 {
				log.Printf("Invalid loopField format: %s", el.LoopField)
				return
			}

			arrayName := parts[0]
			fieldName := parts[1]

			if data, ok := input[arrayName]; ok {
				if items, isArray := data.([]interface{}); isArray {
					currentY := el.Y

					for _, item := range items {
						if itemMap, isMap := item.(map[string]interface{}); isMap {
							if val, ok := itemMap[fieldName]; ok {
								// Round Y to avoid float drift gaps
								currentY = math.Round(currentY*10) / 10

								log.Printf("Drawing box for array item at Y=%.1f", currentY)

								pdf.SetLineWidth(0.2)
								if el.background == "1" {
									pdf.Rect(el.X, currentY, el.Width, el.Height, "FD")
								} else {
									pdf.Rect(el.X, currentY, el.Width, el.Height, "D")
								}

								// Optional: draw text inside the box
								pdf.SetFont(el.Font, el.FontStyle, el.FontSize)
								pdf.SetTextColor(el.ColorR, el.ColorG, el.ColorB)
								pdf.SetXY(el.X, currentY)
								pdf.MultiCell(el.Width, el.Height, fmt.Sprintf("%v", val), el.Border, el.Align, false)

								currentY += el.Height
							}
						}
					}

					lastYPositions[arrayName] = currentY
					return
				}
			}
		}

		// Normal box processing for non-array variables
		log.Printf("Drawing box at (%.1f, %.1f) with size (%.1f, %.1f) and border color R=%d, G=%d, B=%d",
			el.X, el.Y, el.Width, el.Height, el.ColorR, el.ColorG, el.ColorB)

		// Set line width to make borders more visible
		pdf.SetLineWidth(0.2)

		// Draw the rectangle with or without fill
		if el.background == "1" {
			pdf.Rect(el.X, el.Y, el.Width, el.Height, "FD") // Fill and Draw
		} else {
			pdf.Rect(el.X, el.Y, el.Width, el.Height, "D") // Draw only
		}

		// Reset line width
		pdf.SetLineWidth(0.2)

	case "image":
		if el.ImageSrc != "" {
			log.Printf("Drawing image from %s at (%.1f, %.1f)", el.ImageSrc, el.X, el.Y)
			if _, err := os.Stat(el.ImageSrc); err == nil {
				pdf.Image(el.ImageSrc, el.X, el.Y, el.Width, el.Height, false, "", 0, "")
			} else {
				log.Printf("Warning: Image file not found: %s", el.ImageSrc)
			}
		}
	}
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("Warning: Error parsing float '%s': %v", s, err)
		return 0
	}
	return f
}

func parseInt(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("Warning: Error parsing int '%s': %v", s, err)
		return 0
	}
	return i
}
