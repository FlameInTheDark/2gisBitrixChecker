package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Organization main data structure
type Organization struct {
	Name          string `xlsx:"0"`
	Category      string `xlsx:"1"`
	Subcategories string `xlsx:"2"`
	Rubrics       string `xlsx:"3"`
	City          string `xlsx:"4"`
	Address       string `xlsx:"5"`
	Email         string `xlsx:"6"`
	Phone         string `xlsx:"7"`
	Fax           string `xlsx:"8"`
	Site          string `xlsx:"9"`
	ICQ           string `xlsx:"10"`
	Jabber        string `xlsx:"11"`
	Skype         string `xlsx:"12"`
	Vkontakte     string `xlsx:"13"`
	Facebook      string `xlsx:"14"`
	Twitter       string `xlsx:"15"`
	Instagram     string `xlsx:"16"`
	Additional    string `xlsx:"17"`
	Photo         string `xlsx:"18"`
	UploadedPhoto string `xlsx:"19"`
	Bitrix        string `xlsx:"20"`
}

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

// OrgContainer data container for concurrency
type OrgContainer struct {
	mx sync.Mutex
	m  map[int]Organization
}

// MakeContainer creates new OrgContainer
func MakeContainer() OrgContainer {
	return OrgContainer{m: make(map[int]Organization)}
}

// Store saves data to map in OrgContainer
func (o *OrgContainer) Store(key int, org Organization) {
	o.mx.Lock()
	o.m[key] = org
	o.mx.Unlock()
}

// Get returns Organization from OrgContainer by key
func (o *OrgContainer) Get(key int) Organization {
	return o.m[key]
}

// Len returns size of OrgContainer map
func (o *OrgContainer) Len() int {
	return len(o.m)
}

// Map returns pointer to map in OrgContainer
func (o *OrgContainer) Map() *map[int]Organization {
	return &o.m
}

func main() {
	timeFrom := time.Now()
	fileName := flag.String("file", "./Book.xlsx", ".xlsx file path")
	//sheetName := flag.String("sheet", "Sheet1", "excel sheet name")
	routinesCount := flag.Int("routines", 100, "count of Go routines")
	flag.Parse()

	fmt.Println("Opening file...")
	//xlsx, err := excelize.OpenFile(*fileName)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
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
		fmt.Println("Saving to .csv")
		file, err := os.Create(newFileName + ".csv")

		if err != nil {
			return
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Comma = '\t'
		defer writer.Flush()

		for _, value := range csvData {
			err := writer.Write(value)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	} else {
		fmt.Println("Saving to .xlsx")
		saveerr := SaveToXlsx(&csvData, newFileName+".xlsx")
		if saveerr != nil {
			fmt.Printf(err.Error())
			return
		}
	}

	fmt.Println("Saved!")
	fmt.Printf("Time: %v\nThreads: %v\nRows: %v\nChecked: %v\nUnchecked: %v",
		time.Since(timeFrom), *routinesCount, rowNum, checked, unchecked)
}

// Check checking site and save result in OrgContainer
func Check(row xlsx.Row, rowNum, line int) {
	newOrg := &Organization{}
	rErr := row.ReadStruct(newOrg)
	if rErr != nil {
		fmt.Println(rErr.Error())
	}

	timeout := time.Duration(time.Duration(*timeoutSeconds) * time.Second)
	client := &http.Client{
		Timeout: time.Duration(timeout),
	}
	resp, err := client.Get(fmt.Sprintf("http://%v/bitrix/themes/.default/modules.css", newOrg.Site))
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 == nil {
				bodyString := string(bodyBytes)
				if strings.Contains(bodyString, "sale") {
					newOrg.Bitrix = "Битрикс, Малый бизнес / Бизнес"
				} else if strings.Contains(bodyString, "bitrix") {
					newOrg.Bitrix = "Битрикс, не магазин"
				} else {
					newOrg.Bitrix = "Не Битрикс, открывается"
				}
			} else {
				newOrg.Bitrix = "Не Битрикс"
			}
		} else {
			newOrg.Bitrix = "Не Битрикс"
		}
	} else {
		newOrg.Bitrix = "Не открывается"
	}
	org.Store(rowNum, *newOrg)
	checked++
	complete++
	fmt.Printf("[%v:%v] Result of: %v is [%v]\n", complete, line, newOrg.Site, newOrg.Bitrix)
	active--
}

// SaveToXlsx generates .xlsx file from csv data array
func SaveToXlsx(csvFile *[][]string, XLSXPath string) error {
	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("Sheet1")
	if err != nil {
		return err
	}
	for _, fields := range *csvFile {
		row := sheet.AddRow()
		for _, field := range fields {
			cell := row.AddCell()
			cell.Value = field
		}
	}
	if err != nil {
		fmt.Printf(err.Error())
	}
	return xlsxFile.Save(XLSXPath)
}
