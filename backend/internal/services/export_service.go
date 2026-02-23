package services

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

var (
	ErrExportFailed = errors.New("gagal mengekspor data")
	ErrInvalidData  = errors.New("data tidak valid untuk ekspor")
)

// ExportService handles data export operations
type ExportService struct {
	organizationName string
}

// NewExportService creates a new export service
func NewExportService(organizationName string) *ExportService {
	return &ExportService{
		organizationName: organizationName,
	}
}

// ExportData represents generic data for export
type ExportData struct {
	Title       string              // Report title
	Headers     []string            // Column headers in Indonesian
	Rows        [][]string          // Data rows
	DateRange   string              // Date range for the report
	GeneratedBy string              // User who generated the report
	Metadata    map[string]string   // Additional metadata
}

// ExportToPDF exports data to PDF format
func (s *ExportService) ExportToPDF(data *ExportData) (*bytes.Buffer, error) {
	if data == nil || len(data.Headers) == 0 {
		return nil, ErrInvalidData
	}

	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape orientation
	pdf.SetAutoPageBreak(true, 10)
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "", 12)

	// Organization header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, s.organizationName)
	pdf.Ln(8)

	// Report title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, data.Title)
	pdf.Ln(8)

	// Date range
	if data.DateRange != "" {
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 6, fmt.Sprintf("Periode: %s", data.DateRange))
		pdf.Ln(6)
	}

	// Generated info
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(0, 5, fmt.Sprintf("Dibuat oleh: %s | Tanggal: %s", 
		data.GeneratedBy, 
		time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Table headers
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	
	// Calculate column width
	pageWidth, _ := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	availableWidth := pageWidth - leftMargin - rightMargin
	colWidth := availableWidth / float64(len(data.Headers))

	// Draw headers
	for _, header := range data.Headers {
		pdf.CellFormat(colWidth, 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)
	
	for i, row := range data.Rows {
		// Alternate row colors
		if i%2 == 0 {
			pdf.SetFillColor(245, 245, 245)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		for j, cell := range row {
			// Ensure we don't exceed header count
			if j >= len(data.Headers) {
				break
			}
			pdf.CellFormat(colWidth, 7, cell, "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}

	// Footer with page numbers
	pdf.AliasNbPages("")
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")

	// Write to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExportFailed, err)
	}

	return &buf, nil
}

// ExportToExcel exports data to Excel format
func (s *ExportService) ExportToExcel(data *ExportData) (*bytes.Buffer, error) {
	if data == nil || len(data.Headers) == 0 {
		return nil, ErrInvalidData
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Data"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExportFailed, err)
	}
	f.SetActiveSheet(index)

	// Set column widths
	for i := range data.Headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, col, col, 15)
	}

	// Organization header
	f.SetCellValue(sheetName, "A1", s.organizationName)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14},
	})
	f.SetCellStyle(sheetName, "A1", "A1", headerStyle)

	// Report title
	f.SetCellValue(sheetName, "A2", data.Title)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
	})
	f.SetCellStyle(sheetName, "A2", "A2", titleStyle)

	// Date range
	currentRow := 3
	if data.DateRange != "" {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("Periode: %s", data.DateRange))
		currentRow++
	}

	// Generated info
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", currentRow), 
		fmt.Sprintf("Dibuat oleh: %s | Tanggal: %s", 
			data.GeneratedBy, 
			time.Now().Format("02/01/2006 15:04")))
	currentRow += 2 // Add spacing

	// Table headers
	headerRowStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	for i, header := range data.Headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s%d", col, currentRow)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerRowStyle)
	}
	currentRow++

	// Table data
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	alternateStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#F5F5F5"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	for rowIdx, row := range data.Rows {
		for colIdx, cell := range row {
			// Ensure we don't exceed header count
			if colIdx >= len(data.Headers) {
				break
			}
			col, _ := excelize.ColumnNumberToName(colIdx + 1)
			cellRef := fmt.Sprintf("%s%d", col, currentRow)
			f.SetCellValue(sheetName, cellRef, cell)
			
			// Apply alternating row style
			if rowIdx%2 == 0 {
				f.SetCellStyle(sheetName, cellRef, cellRef, alternateStyle)
			} else {
				f.SetCellStyle(sheetName, cellRef, cellRef, dataStyle)
			}
		}
		currentRow++
	}

	// Auto-filter
	lastCol, _ := excelize.ColumnNumberToName(len(data.Headers))
	headerRow := currentRow - len(data.Rows) - 1
	f.AutoFilter(sheetName, fmt.Sprintf("A%d:%s%d", headerRow, lastCol, currentRow-1), nil)

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExportFailed, err)
	}

	return &buf, nil
}

// ExportToExcelMultiSheet exports multiple datasets to Excel with multiple sheets
func (s *ExportService) ExportToExcelMultiSheet(datasets map[string]*ExportData) (*bytes.Buffer, error) {
	if len(datasets) == 0 {
		return nil, ErrInvalidData
	}

	f := excelize.NewFile()
	defer f.Close()

	firstSheet := true
	for sheetName, data := range datasets {
		if data == nil || len(data.Headers) == 0 {
			continue
		}

		var index int
		var err error
		if firstSheet {
			// Rename default sheet
			f.SetSheetName("Sheet1", sheetName)
			index = 0
			firstSheet = false
		} else {
			index, err = f.NewSheet(sheetName)
			if err != nil {
				continue
			}
		}
		f.SetActiveSheet(index)

		// Similar logic as ExportToExcel but for each sheet
		// Set column widths
		for i := range data.Headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			f.SetColWidth(sheetName, col, col, 15)
		}

		// Headers and data (simplified version)
		currentRow := 1
		
		// Table headers
		headerRowStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
		})

		for i, header := range data.Headers {
			col, _ := excelize.ColumnNumberToName(i + 1)
			cell := fmt.Sprintf("%s%d", col, currentRow)
			f.SetCellValue(sheetName, cell, header)
			f.SetCellStyle(sheetName, cell, cell, headerRowStyle)
		}
		currentRow++

		// Table data
		for _, row := range data.Rows {
			for colIdx, cellValue := range row {
				if colIdx >= len(data.Headers) {
					break
				}
				col, _ := excelize.ColumnNumberToName(colIdx + 1)
				cellRef := fmt.Sprintf("%s%d", col, currentRow)
				f.SetCellValue(sheetName, cellRef, cellValue)
			}
			currentRow++
		}
	}

	// Write to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExportFailed, err)
	}

	return &buf, nil
}
