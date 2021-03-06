package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
)

func getCellWidth(cell string) int {
	indexOfNewline := strings.Index(cell, "\n")
	if indexOfNewline > -1 {
		return indexOfNewline + 1
	} else {
		return len(cell)
	}
}

func getTruncatedAndPaddedCell(cell string, width int) string {
	indexOfNewline := strings.Index(cell, "\n")
	var lineString string
	if indexOfNewline > -1 {
		lineString = cell[:indexOfNewline]
	} else {
		lineString = cell
	}
	if len(lineString) == width {
		return lineString
	} else if len(lineString) < width {
		return lineString + strings.Repeat(" ", width-len(lineString))
	} else {
		return lineString[:width-3] + "..."
	}
}

func copyTruncatedAndPaddedCellToOutputRow(outrow, row []string, widths []int) {
	for i, cell := range row {
		outrow[i] = getTruncatedAndPaddedCell(cell, widths[i])
	}
}

func getRowSeparator(widths []int) string {
	cells := make([]string, len(widths))
	for i, width := range widths {
		cells[i] = strings.Repeat("-", width)
	}
	return fmt.Sprintf("+-%s-+", strings.Join(cells, "-+-"))
}

func View(reader *csv.Reader, maxWidth, maxRows int) {

	imc := NewInMemoryCsv(reader)

	// Default to 0
	columnWidths := make([]int, imc.NumColumns())
	for j, cell := range imc.header {
		cellLength := getCellWidth(cell)
		if cellLength > columnWidths[j] {
			if cellLength > maxWidth {
				columnWidths[j] = maxWidth
			} else {
				columnWidths[j] = cellLength
			}
		}
	}

	// Get the actual number of rows to display
	numRowsToView := imc.NumRows()
	if maxRows > 0 && maxRows < numRowsToView {
		numRowsToView = maxRows
	}

	for i := 0; i < numRowsToView; i++ {
		row := imc.rows[i]
		for j, cell := range row {
			if columnWidths[j] == maxWidth {
				continue
			}
			cellLength := getCellWidth(cell)
			if cellLength > columnWidths[j] {
				if cellLength > maxWidth {
					columnWidths[j] = maxWidth
				} else {
					columnWidths[j] = cellLength
				}
			}
		}
	}

	rowSeparator := getRowSeparator(columnWidths)
	outrow := make([]string, imc.NumColumns())

	// Top of table
	fmt.Println(rowSeparator)

	// Print header
	copyTruncatedAndPaddedCellToOutputRow(outrow, imc.header, columnWidths)
	fmt.Printf("| %s |\n", strings.Join(outrow, " | "))
	fmt.Println(rowSeparator)

	// Print rows
	for i := 0; i < numRowsToView; i++ {
		row := imc.rows[i]
		copyTruncatedAndPaddedCellToOutputRow(outrow, row, columnWidths)
		fmt.Printf("| %s |\n", strings.Join(outrow, " | "))
		fmt.Println(rowSeparator)
	}
}

func RunView(args []string) {
	fs := flag.NewFlagSet("view", flag.ExitOnError)
	var maxWidth, maxRows int
	fs.IntVar(&maxWidth, "max-width", 20, "Maximum width per column")
	fs.IntVar(&maxWidth, "w", 20, "Maximum width per column (shorthand)")
	fs.IntVar(&maxRows, "n", 0, "Number of rows to display")
	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}

	if maxWidth < 1 {
		fmt.Fprintln(os.Stderr, "Invalid argument --max-width")
		os.Exit(1)
	}
	if maxRows < 0 {
		maxRows = 0
	}

	// Get input CSV
	moreArgs := fs.Args()
	if len(moreArgs) > 1 {
		fmt.Fprintln(os.Stderr, "Can only view one table")
		os.Exit(1)
	}
	var reader *csv.Reader
	if len(moreArgs) == 1 {
		file, err := os.Open(moreArgs[0])
		if err != nil {
			panic(err)
		}
		defer file.Close()
		reader = csv.NewReader(file)
	} else {
		reader = csv.NewReader(os.Stdin)
	}

	View(reader, maxWidth, maxRows)
}
