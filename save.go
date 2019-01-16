package main

import (
	"encoding/csv"
	"fmt"
	"github.com/tealeg/xlsx"
	"os"
)

// SaveToXlsx generates .xlsx file from csv data array
func SaveToXlsx(csvData *[][]string, filePath string) error {
	fmt.Println("Saving to .xlsx")
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("Sheet1")
	if err != nil {
		return err
	}
	for _, fields := range *csvData {
		row := sheet.AddRow()
		for _, field := range fields {
			cell := row.AddCell()
			cell.Value = field
		}
	}
	if err != nil {
		fmt.Printf(err.Error())
	}
	return xlsxFile.Save(filePath)
}

// SaveToCsv saves data to .csv file
func SaveToCsv(csvData *[][]string, filePath string) error {
	fmt.Println("Saving to .csv")
	file, fileErr := os.Create(filePath + ".csv")

	if fileErr != nil {
		return fileErr
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	for _, value := range *csvData {
		writeErr := writer.Write(value)
		if writeErr != nil {
			return writeErr
		}
	}
	return nil
}
