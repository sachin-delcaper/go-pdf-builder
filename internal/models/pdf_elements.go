package models

import (
	"encoding/json"
	"fmt"
)

// ElementType represents the type of PDF element
type ElementType string

const (
	ElementTypeText    ElementType = "text"
	ElementTypeBox     ElementType = "box"
	ElementTypeImage   ElementType = "image"
	ElementTypeQR      ElementType = "qr"
	ElementTypeBarcode ElementType = "barcode"
	ElementTypeTable   ElementType = "table"
)

// PDFElement represents a single element in the PDF template
type PDFElement struct {
	Type         ElementType   `json:"type" csv:"type"`
	Method       string        `json:"method" csv:"method"`
	Text         string        `json:"text" csv:"text"`
	VariableName string        `json:"variableName" csv:"variableName"`
	Position     Position      `json:"position"`
	Size         Size          `json:"size"`
	Style        Style         `json:"style"`
	LoopField    string        `json:"loopField" csv:"loopField"`
	Columns      []TableColumn `json:"columns,omitempty"`

	// QR/Barcode specific fields
	QRContent      string `json:"qrContent,omitempty" csv:"qrContent"`
	BarcodeFormat  string `json:"barcodeFormat,omitempty" csv:"barcodeFormat"`
	BarcodeContent string `json:"barcodeContent,omitempty" csv:"barcodeContent"`
}

// Position represents the position of an element
type Position struct {
	X float64 `json:"x" csv:"x"`
	Y float64 `json:"y" csv:"y"`
}

// Size represents the dimensions of an element
type Size struct {
	Width  float64 `json:"width" csv:"width"`
	Height float64 `json:"height" csv:"height"`
}

// Style contains styling information for elements
type Style struct {
	Font         Font   `json:"font"`
	Border       string `json:"border" csv:"border"`
	Align        string `json:"align" csv:"align"`
	RotateDegree int    `json:"rotateDegree" csv:"rotateDegree"`
	RotateType   string `json:"rotateType" csv:"rotateType"`
	TextColor    Color  `json:"textColor"`
	Background   Color  `json:"background"`
	ImageSrc     string `json:"imageSrc" csv:"imageSrc"`
}

// Font represents font styling
type Font struct {
	Family string  `json:"family" csv:"font"`
	Style  string  `json:"style" csv:"fontStyle"`
	Size   float64 `json:"size" csv:"fontSize"`
}

// Color represents RGB color values
type Color struct {
	R     int  `json:"r" csv:"colorR"`
	G     int  `json:"g" csv:"colorG"`
	B     int  `json:"b" csv:"colorB"`
	IsSet bool `json:"isSet,omitempty"`
}

// TableColumn represents a column in a table
type TableColumn struct {
	Field     string  `json:"field"`
	Width     float64 `json:"width"`
	Align     string  `json:"align"`
	FontStyle string  `json:"fontStyle"`
}

// CSVTemplateRequest represents the JSON input for the CSV template endpoint
type CSVTemplateRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

// Validate checks if the PDF element has valid values
func (e *PDFElement) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("element type is required")
	}

	if e.Position.X < 0 || e.Position.Y < 0 {
		return fmt.Errorf("invalid position: x=%.2f, y=%.2f", e.Position.X, e.Position.Y)
	}

	if e.Size.Width <= 0 || e.Size.Height <= 0 {
		return fmt.Errorf("invalid size: width=%.2f, height=%.2f", e.Size.Width, e.Size.Height)
	}

	// Type-specific validation
	switch e.Type {
	case ElementTypeQR:
		if e.QRContent == "" && e.VariableName == "" {
			return fmt.Errorf("QR element requires either qrContent or variableName")
		}
	case ElementTypeBarcode:
		if e.BarcodeContent == "" && e.VariableName == "" {
			return fmt.Errorf("barcode element requires either barcodeContent or variableName")
		}
		if e.BarcodeFormat == "" {
			e.BarcodeFormat = "Code128" // Default format
		}
	case ElementTypeImage:
		if e.Style.ImageSrc == "" && e.VariableName == "" {
			return fmt.Errorf("image element requires either imageSrc or variableName")
		}
	}

	return nil
}

// IsLoopElement returns true if this element should be processed in a loop
func (e *PDFElement) IsLoopElement() bool {
	return e.LoopField != ""
}

// GetTextContent returns the text content for the element, processing variables
func (e *PDFElement) GetTextContent(data map[string]interface{}) string {
	content := e.Text

	// Process QR content
	if e.Type == ElementTypeQR {
		if e.QRContent != "" {
			content = e.QRContent
		} else if e.VariableName != "" {
			if val, ok := data[e.VariableName]; ok {
				content = fmt.Sprintf("%v", val)
			}
		}
	}

	// Process barcode content
	if e.Type == ElementTypeBarcode {
		if e.BarcodeContent != "" {
			content = e.BarcodeContent
		} else if e.VariableName != "" {
			if val, ok := data[e.VariableName]; ok {
				content = fmt.Sprintf("%v", val)
			}
		}
	}

	return content
}

// Clone creates a deep copy of the PDFElement
func (e *PDFElement) Clone() *PDFElement {
	clone := *e
	if len(e.Columns) > 0 {
		clone.Columns = make([]TableColumn, len(e.Columns))
		copy(clone.Columns, e.Columns)
	}
	return &clone
}

// String returns a string representation of the element for debugging
func (e *PDFElement) String() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}
