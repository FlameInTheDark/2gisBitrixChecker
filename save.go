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
	Emails []struct {
		Value     string `json:"VALUE"`
		ValueType string `json:"VALUE_TYPE"`
	}
}

// ReturnResultList contains slice of companies and next offset
type ReturnResultList struct {
	Result []ReturnableCompany `json:"result"`
	Next   int                 `json:"next"`
}

// SaveCRM saves checked data in CRM
func SaveCRM() {
	var results = make(map[string]ReturnableCompany)
	var resultsDomains = make(map[string]string)
	var resultsMails = make(map[string]string)
	var resultsPhones = make(map[string]string)

	// Getting companies from CRM
	v := GetCompanies(0)
	for _, val := range v.Result {
		if len(val.Sites) != 0 {
			results[val.ID] = val
		}
		for _, vs := range val.Sites {
			resultsDomains[trimDomain(vs.Value)] = val.ID
		}
		for _, ve := range val.Emails {
			resultsMails[ve.Value] = val.ID
		}
		for _, vp := range val.Phones {
			resultsPhones[trimPhone(vp.Value)] = val.ID
		}
	}
	for v.Next != 0 {
		for _, val := range v.Result {
			if len(val.Sites) != 0 {
				results[val.ID] = val
			}
			for _, vs := range val.Sites {
				resultsDomains[trimDomain(vs.Value)] = val.ID
			}
			for _, ve := range val.Emails {
				resultsMails[ve.Value] = val.ID
			}
			for _, vp := range val.Phones {
				resultsPhones[trimPhone(vp.Value)] = val.ID
			}
		}
		v = GetCompanies(v.Next)
	}
	created := 0
	for id, v := range *org.Map() {
		// Checking companies and creating new if company not exists
		if v.ToSave {
			var isExists = false
			if _, ok := resultsDomains[trimDomain(v.Site)]; ok {
				isExists = true
				UpdateCompany(org.Get(id), results[resultsDomains[trimDomain(v.Site)]])
			} else if _, ok := resultsMails[v.Email]; ok {
				isExists = true
				UpdateCompany(org.Get(id), results[resultsMails[v.Email]])
			} else if _, ok := resultsPhones[trimPhone(v.Phone)]; ok {
				isExists = true
				UpdateCompany(org.Get(id), results[resultsPhones[trimPhone(v.Phone)]])
			}
			if !isExists {
				CreateCompany(v)
				created++
			}
		}
	}
	for active > 0 {
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("Created companies: %d\n", created)
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
	request.Select = []string{"WEB", "PHONE", "EMAIL"}
	request.Start = next

	rf, _ := json.Marshal(request)
	rb := bytes.NewReader(rf)
	resp, rErr := http.Post(fmt.Sprintf("%s/crm.company.list", bxConn), "application/json", rb)
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
	// License ID
	license["bitrix"] = "485"
	license["sale"] = "487"
	// Bitrix type ID
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

	_, _ = http.Post(fmt.Sprintf("%s/crm.company.add", bxConn), "application/json", r)
	active--
}

func UpdateCompany(org *Organization, company ReturnableCompany) {
	active++
	var phones []Phone
	var sites []Site
	var emails []Email

	// getting phones, emails and sites
	orgPhones := strings.Split(org.Phone, ",")
	orgEmails := strings.Split(org.Email, ",")
	orgSites := strings.Split(org.Site, ",")

	for _, v := range company.Phones {
		phones = append(phones, Phone{v.Value, v.ValueType})
	}
	for _, v := range company.Emails {
		emails = append(emails, Email{v.Value, v.ValueType})
	}
	for _, v := range company.Sites {
		sites = append(sites, Site{v.Value, v.ValueType})
	}

	for _, v := range orgPhones {
		phones = append(phones, Phone{trimPhone(v), "WORK"})
	}
	for _, v := range orgEmails {
		emails = append(emails, Email{v, "WORK"})
	}

	for _, v := range orgSites {
		sites = append(sites, Site{v, "WORK"})
	}

	// Company fields for request
	newFields := CompanyUpdate{
		phones,
		sites,
		emails,
	}

	f, _ := json.Marshal(FieldsUpdate{company.ID, newFields})

	r := bytes.NewReader(f)

	_, _ = http.Post(fmt.Sprintf("%s/crm.company.update", bxConn), "application/json", r)
	active--
}
