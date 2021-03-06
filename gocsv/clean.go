package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	EXCEL_CELL_CHAR_LIMIT = 32767
	NUMBERS_ROW_LIMIT     = 65535
)

func GetStringForRowIndex(index int) string {
	if index == 0 {
		return "Header"
	} else {
		return fmt.Sprintf("Row %d", index)
	}
}
func GetStringForColumnIndex(index int) string {
	return fmt.Sprintf("Column %d", index+1)
}

func PrintCleanCheck(rowIndex, columnIndex int, message string) {
	preludeParts := make([]string, 0)
	if rowIndex > -1 {
		rowString := GetStringForRowIndex(rowIndex)
		preludeParts = append(preludeParts, rowString)
	}
	if columnIndex > -1 {
		columnString := GetStringForColumnIndex(columnIndex)
		preludeParts = append(preludeParts, columnString)
	}
	var prelude string
	if len(preludeParts) > 0 {
		prelude = strings.Join(preludeParts, ", ") + ": "
	} else {
		prelude = ""
	}
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s%s\n", prelude, message))
}

func Clean(reader *csv.Reader, noTrim, excel, numbers, verbose bool) {
	writer := csv.NewWriter(os.Stdout)

	// Disable errors when fields are varying length
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	// Read in rows.
	rows, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// Determine how many columns there actually should be.
	numColumns := 0
	trimFromIndex := -1
	for i, row := range rows {
		lastNonEmptyIndex := -1
		for j, elem := range row {
			if elem != "" {
				lastNonEmptyIndex = j
			}
		}
		if lastNonEmptyIndex > -1 {
			trimFromIndex = -1
		} else if trimFromIndex == -1 {
			trimFromIndex = i
		}
		numColumnsInRow := lastNonEmptyIndex + 1
		if numColumns < numColumnsInRow {
			numColumns = numColumnsInRow
		}
	}

	// Fix rows and output them to writer.
	shellRow := make([]string, numColumns)
	for i, row := range rows {
		if numbers && i >= NUMBERS_ROW_LIMIT {
			if verbose {
				PrintCleanCheck(i, -1, fmt.Sprintf("Numbers row limit exceeded. Removing last %d rows.", len(rows)-NUMBERS_ROW_LIMIT))
			}
			break
		}
		if !noTrim && trimFromIndex > -1 && i >= trimFromIndex {
			if verbose {
				PrintCleanCheck(i, -1, fmt.Sprintf("Trimming %d trailing empty rows.", len(rows)-trimFromIndex))
			}
			break
		}
		if len(row) == numColumns {
			// Just write the original row.
			copy(shellRow, row)
		} else if len(row) < numColumns {
			if verbose {
				PrintCleanCheck(i, -1, fmt.Sprintf("Padding with %d cells.", numColumns-len(row)))
			}
			// Pad the row.
			copy(shellRow, row)
			for i := len(row); i < numColumns; i++ {
				shellRow[i] = ""
			}
		} else {
			// Truncate the row.
			if verbose {
				PrintCleanCheck(i, -1, fmt.Sprintf("Trimming %d trailing empty cells.", len(row)-numColumns))
			}
			copy(shellRow, row)
		}
		if excel {
			for j, cell := range shellRow {
				if len(cell) > EXCEL_CELL_CHAR_LIMIT {
					numExtraChars := len(cell) - EXCEL_CELL_CHAR_LIMIT
					shellRow[j] = cell[0:EXCEL_CELL_CHAR_LIMIT]
					if verbose {
						PrintCleanCheck(i, j, fmt.Sprintf("Excel cell character limit exceeded. Removing %d characters from cell.", numExtraChars))
					}
				}
			}
		}
		writer.Write(shellRow)
		writer.Flush()
	}
}

func RunClean(args []string) {
	fs := flag.NewFlagSet("clean", flag.ExitOnError)
	var noTrim, excel, numbers, verbose bool
	fs.BoolVar(&noTrim, "no-trim", false, "Don't trim end of file of empty rows")
	fs.BoolVar(&excel, "excel", false, "Clean for use in Excel")
	fs.BoolVar(&numbers, "numbers", false, "Clean for use in Numbers")
	fs.BoolVar(&verbose, "verbose", false, "Print messages when cleaning")
	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}
	moreArgs := fs.Args()
	if len(moreArgs) > 1 {
		fmt.Fprintln(os.Stderr, "Can only clean one file")
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
	Clean(reader, noTrim, excel, numbers, verbose)
}
