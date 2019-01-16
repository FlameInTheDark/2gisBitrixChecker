package main

import (
	"flag"
	"fmt"
	"sort"
	"time"

	"github.com/tealeg/xlsx"
)

var (
	rowNum         = 1
	checked        = 0
	unchecked      = 0
	lines          = 0
	complete       = 0
	active         = 0
	timeoutSeconds = flag.Int("timeout", 20, "getting http data timeout")
	siteColumn     = flag.Int("site", 9, "number of column with site domain")
	toCsv          = flag.Bool("csv", false, "save data to .csv")
	org            = MakeContainer()
)

func main() {
	timeFrom := time.Now()
	fileName := flag.String("file", "", ".xlsx file path")
	routinesCount := flag.Int("routines", 100, "count of Go routines")
	flag.Parse()

	if *fileName == "" {
		flag.Usage()
		fmt.Println("Example: ./2gisBitrixChecker -file=\"./Book.xlsx\" -routines=200")
		return
	}

	fmt.Println("Opening file...")
	xlsxFile, err := xlsx.OpenFile(*fileName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("File opened, loading rows...")

	var completedSites = make(map[string]bool)

	// Get all the rows in the Sheet1.
	sheet := xlsxFile.Sheets[0]
	fmt.Println("Rows loaded, checking...")
	for i := 0; i < len(sheet.Rows); i++ {
		if len(sheet.Rows[i].Cells) > *siteColumn {
			if sheet.Rows[i].Cells[*siteColumn].String() != "" {
				if completedSites[sheet.Rows[i].Cells[*siteColumn].String()] != true {
					if active < *routinesCount {
						go Check(*sheet.Rows[i], rowNum, lines)
						lines++
						active++
					} else {
						for active >= *routinesCount {
							time.Sleep(time.Second)
						}
						go Check(*sheet.Rows[i], rowNum, lines)
						lines++
						active++
					}
					completedSites[sheet.Rows[i].Cells[*siteColumn].String()] = true
				}
			} else {
				unchecked++
			}
		}
		rowNum++
	}

	for active > 0 {
		time.Sleep(time.Second)
	}

	fmt.Println("Checked, saving...")
	keys := make([]int, 0, org.Len())
	for k := range *org.Map() {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	var csvData = make([][]string, 0, org.Len())

	for _, k := range keys {
		csvData = append(csvData, []string{
			org.Get(k).Name,
			org.Get(k).Category,
			org.Get(k).Subcategories,
			org.Get(k).Rubrics,
			org.Get(k).City,
			org.Get(k).Address,
			org.Get(k).Email,
			org.Get(k).Phone,
			org.Get(k).Fax,
			org.Get(k).Site,
			org.Get(k).ICQ,
			org.Get(k).Jabber,
			org.Get(k).Skype,
			org.Get(k).Vkontakte,
			org.Get(k).Facebook,
			org.Get(k).Twitter,
			org.Get(k).Instagram,
			org.Get(k).Additional,
			org.Get(k).Photo,
			org.Get(k).UploadedPhoto,
			org.Get(k).Bitrix,
		})
	}

	newFileName := fmt.Sprintf("CheckResult_%v-%v-%v_%v-%v-%v",
		timeFrom.Year(), timeFrom.Month(), timeFrom.Day(),
		timeFrom.Hour(), timeFrom.Minute(), timeFrom.Second())

	// Saving results
	if *toCsv {
		saveErr := SaveToCsv(&csvData, newFileName)
		if saveErr != nil {
			fmt.Println(saveErr.Error())
			return
		}
	} else {
		saveErr := SaveToXlsx(&csvData, newFileName+".xlsx")
		if saveErr != nil {
			fmt.Println(saveErr.Error())
			return
		}
	}

	fmt.Println("Saved!")
	fmt.Printf("Time: %v\nThreads: %v\nRows: %v\nChecked: %v\nUnchecked: %v",
		time.Since(timeFrom), *routinesCount, rowNum, checked, unchecked)
}
