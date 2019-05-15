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

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
	defer writer.Flush()

	for _, value := range *csvData {
		writeErr := writer.Write(value)
		if writeErr != nil {
			return writeErr
		}
	}
	_ = file.Close()
	return nil
}

// ReturnableCompany contains company data
type ReturnableCompany struct {
	ID    string `json:"ID"`
	Sites []struct {
		Value     string `json:"VALUE"`
		ValueType string `json:"VALUE_TYPE"`
	} `json:"WEB"`
	Phones []struct {
		Value     string `json:"VALUE"`
		ValueType string `json:"VALUE_TYPE"`
	} `json:"PHONE"`
}

// ReturnResultList contains slice of companies and next offset
type ReturnResultList struct {
	Result []ReturnableCompany `json:"result"`
	Next   int                 `json:"next"`
}

// SaveCRM saves checked data in CRM
func SaveCRM() {
	var results = make(map[string]ReturnableCompany)
	// Getting companies from CRM
	v := GetCompanies(0)
	for _, val := range v.Result {
		if len(val.Sites) != 0 {
			results[trimDomain(val.Sites[0].Value)] = val
		}
	}
	for v.Next != 0 {
		for _, val := range v.Result {
			if len(val.Sites) != 0 {
				results[trimDomain(val.Sites[0].Value)] = val
			}
		}
		v = GetCompanies(v.Next)
	}
	created := 0
	for _, v := range *org.Map() {
		// Checking companies and creating new if company not exists
		if _, ok := results[trimDomain(v.Site)]; !ok && v.ToSave {
			var isPhoneExists = false
			for _,p := range results[trimDomain(v.Site)].Phones {
				if trimPhone(p.Value) == trimPhone(v.Phone) {
					isPhoneExists = true
				}
			}
			if !isPhoneExists {
				CreateCompany(v)
				created++
			}
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
	request.Select = []string{"WEB", "PHONE"}
	request.Start = next

	rf, _ := json.Marshal(request)
	rb := bytes.NewReader(rf)
	resp, rErr := http.Post(fmt.Sprintf("%v/crm.company.list", bxConn), "application/json", rb)
	if rErr != nil {
		result.Next = next
		fmt.Println(rErr.Error())
		return result
	}
	b, bErr := ioutil.ReadAll(resp.Body)
	if bErr != nil {
		result.Next = next
		fmt.Println(bErr.Error())
		return result
	}
	_ = json.Unmarshal(b, &result)
	return result
}

// CreateCompany creates new company in CRM
func CreateCompany(org *Organization) {
	active++
	var phones []Phone
	var sites []Site
	var emails []Email

	// getting phones, emails and sites
	orgPhones := strings.Split(org.Phone, ",")
	orgEmails := strings.Split(org.Email, ",")
	orgSites := strings.Split(org.Site, ",")

	for _, v := range orgPhones {
		phones = append(phones, Phone{trimPhone(v), "WORK"})
	}
	for _, v := range orgEmails {
		emails = append(emails, Email{v, "WORK"})
	}

	for _, v := range orgSites {
		sites = append(sites, Site{v, "WORK"})
	}

	// TODO: change to existing ids
	var license = make(map[string]string)
	var bxtype = make(map[string]string)
	license["bitrix"] = "485"
	license["sale"] = "487"
	bxtype["bitrix"] = "507"
	bxtype["sale"] = "511"

	// Company fields for request
	newFields := Company{
		org.Name,
		"CUSTOMER",
		"Y",
		"147",
		phones,
		sites,
		emails,
		"447",
		[]string{"471"},
		[]string{license[org.Bitrix], bxtype[org.Bitrix]},
	}

	f, _ := json.Marshal(Fields{newFields})

	r := bytes.NewReader(f)

	_, _ = http.Post(fmt.Sprintf("%v/crm.company.add", bxConn), "application/json", r)
	active--
}
