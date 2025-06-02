# Invoice Template Documentation

## Overview

The invoice template system provides a professional PDF invoice generation with the following color scheme:
- **Black text**: Static labels (headers, field names)
- **Blue text**: Dynamic data (values that change per invoice)
- **Red text**: Important labels (invoice number, totals)
- **Gray**: Borders and footer text

## Template Structure

The invoice template includes the following sections:

### 1. Company Header
- Company Logo (optional)
- Company Name
- Company Address
- Company GSTIN

### 2. Invoice Details
- Invoice Number
- Invoice Date
- Due Date
- Consignment Note Number
- Service Date
- Origin & Destination

### 3. Customer Information
- Customer Name
- Customer Address
- Customer GSTIN
- Phone & Email

### 4. Product/Service Details
- Product Description
- Weight
- Value of Goods
- HSN/SAC Code
- State & State Code

### 5. Charges Table
- Item descriptions with amounts
- Subtotal
- CGST/SGST/IGST (as applicable)
- Total Amount
- Amount in Words

### 6. Optional Elements
- QR Code
- Barcode
- Company Logo

## API Endpoint

```
POST /invoice/template
Content-Type: application/json
```

## Sample Request

```json
{
  "CompanyName": "SPEEDEX COURIER SERVICES",
  "CompanyAddress": "123 Business Park, Main Road, Mumbai - 400001",
  "CompanyGSTIN": "27AAAAA0000A1Z5",
  "CompanyPhone": "+91-22-12345678",
  "CompanyEmail": "info@speedexcourier.com",
  
  "InvoiceNumber": "INV/2024/001234",
  "InvoiceDate": "2024-03-20",
  "DueDate": "2024-04-20",
  
  "CustomerName": "John Doe Enterprises",
  "CustomerAddress": "456 Commercial Complex, Andheri East, Mumbai - 400069",
  "CustomerGSTIN": "27BBBBB0000B1Z5",
  "CustomerPhone": "+91-9876543210",
  "CustomerEmail": "billing@johndoe.com",
  
  "ConsignmentNo": "CNS789456123",
  "Origin": "Mumbai",
  "Destination": "Delhi",
  "Weight": "2.5 kg",
  "Product": "Documents",
  "ServiceDate": "2024-03-20",
  
  "ValueOfGoods": "50000.00",
  "HSNCode": "996812",
  "StateCode": "27",
  "State": "Maharashtra",
  
  "ChargeItems": [
    {
      "Description": "Courier Service Charges",
      "Amount": 500.00
    },
    {
      "Description": "Handling Charges",
      "Amount": 100.00
    },
    {
      "Description": "Insurance Premium",
      "Amount": 50.00
    }
  ],
  
  "SubTotal": 650.00,
  "CGSTRate": 9,
  "SGSTRate": 9,
  "AmountInWords": "Eight Hundred and Fifty Rupees Only",
  
  "LogoPath": "/path/to/company-logo.png",
  "QRCodePath": "/path/to/qr-code.png",
  "BarcodePath": "/path/to/barcode.png"
}
```

## Auto-Calculation Features

The template automatically calculates the following if not provided:
- **SubTotal**: Sum of all charge items
- **CGST Amount**: SubTotal × CGST Rate / 100
- **SGST Amount**: SubTotal × SGST Rate / 100
- **IGST Amount**: SubTotal × IGST Rate / 100
- **Total Amount**: SubTotal + All Tax Amounts

## Tax Scenarios

1. **Within State (CGST + SGST)**: Set `CGSTRate` and `SGSTRate`
2. **Inter State (IGST)**: Set only `IGSTRate`
3. **No Tax**: Leave all tax rates as 0

## Image Requirements

- **Logo**: PNG format, recommended size 200x80 pixels
- **QR Code**: PNG format, square aspect ratio
- **Barcode**: PNG format, horizontal aspect ratio

Images are optional. If paths are invalid or files don't exist, the invoice will still generate successfully.

## Testing with cURL

```bash
curl -X POST http://localhost:8080/invoice/template \
  -H "Content-Type: application/json" \
  -d '{
    "CompanyName": "TEST COMPANY",
    "InvoiceNumber": "TEST-001",
    "CustomerName": "Test Customer",
    "ChargeItems": [{"Description": "Service", "Amount": 100}],
    "CGSTRate": 9,
    "SGSTRate": 9
  }' \
  -o test-invoice.pdf
```

## Color Coding in Generated PDF

- **Black (#000000)**: Static labels like "Invoice No:", "Customer Name:", etc.
- **Blue (#0000FF)**: Dynamic values like invoice number, customer details, amounts
- **Red (#800000)**: Important labels like "TAX INVOICE", "TOTAL AMOUNT"
- **Gray (#808080)**: Borders, lines, and footer text

## Notes

1. All monetary values should be provided as numbers (not strings)
2. Tax rates should be provided as percentages (e.g., 9 for 9%)
3. The template uses Tahoma font for consistency
4. Page size is A4 with 10mm margins
5. The invoice includes automatic page breaks for long item lists 