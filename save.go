package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
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

type ReturnableCompany struct {
	ID    string `json:"ID"`
	Sites []struct {
		Value     string `json:"VALUE"`
		ValueType string `json:"VALUE_TYPE"`
	} `json:"WEB"`
}

type ReturnResultList struct {
	Result []ReturnableCompany `json:"result"`
	Next   int                 `json:"next"`
}

// SaveCRM saves checked data in CRM
func SaveCRM() {
	var results = make(map[string]ReturnableCompany)
	v := GetCompanies(0)
	for _, val := range v.Result {
		if len(val.Sites) != 0 {
			results[val.Sites[0].Value] = val
		}
	}
	for v.Next != 0 {
		for _, val := range v.Result {
			if len(val.Sites) != 0 {
				results[val.Sites[0].Value] = val
			}
		}
		v = GetCompanies(v.Next)
	}
	created := 0
	for _, v := range *org.Map() {
		if _, ok := results[v.Site]; !ok && v.ToSave {
			CreateCompany(&v)
			created++
		}
	}
	for active > 0 {
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("Created companies: %v\n", created)
}

// GetCompanies get's offset and returns 50 companies from CRM
func GetCompanies(next int) ReturnResultList {
	var result ReturnResultList

	request := struct {
		Order struct {
			DateCreate string `json:"DATE_CREATE"`
		} `json:"order"`
		Select []string `json:"select"`
		Start  int      `json:"start"`
	}{}

	request.Order.DateCreate = "ASC"
	request.Select = []string{"WEB"}
	request.Start = next

	rf, _ := json.Marshal(request)
	rb := bytes.NewReader(rf)
	resp, _ := http.Post(fmt.Sprintf("%v/crm.company.list", bxConn), "application/json", rb)
	b, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(b, &result)
	return result
}

// CreateCompany creates new company in CRM
func CreateCompany(org *Organization) {
	active++
	var phones []Phone
	var sites []Site
	var emails []Email

	orgPhones := strings.Split(org.Phone, ",")
	orgEmails := strings.Split(org.Email, ",")

	for _, v := range orgPhones {
		phones = append(phones, Phone{v, "WORK"})
	}
	for _, v := range orgEmails {
		emails = append(emails, Email{v, "WORK"})
	}

	orgSites := strings.Split(org.Site, ",")

	for _, v := range orgSites {
		sites = append(sites, Site{v, "WORK"})
	}

	// Company fields for request
	newFields := Company{
		org.Name,
		"CUSTOMER",
		"Y",
		"1",
		phones,
		sites,
		emails,
		"119",
	}

	f, _ := json.Marshal(Fields{newFields})

	r := bytes.NewReader(f)

	_, _ = http.Post(fmt.Sprintf("%v/crm.company.add", bxConn), "application/json", r)
	active--
}
