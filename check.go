package main

import (
	"fmt"
	"github.com/tealeg/xlsx"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

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

	// Check if multiple sites in one row
	if strings.Contains(newOrg.Site, ",") {
		siteArr := strings.Split(newOrg.Site, ",")
		newOrg.Site = strings.Replace(siteArr[0], " ", "", -1)
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
					newOrg.ToSave = true
					bitrixes++
				} else if strings.Contains(bodyString, "bitrix") {
					newOrg.Bitrix = "Битрикс, не магазин"
					newOrg.ToSave = true
					bitrixes++
				} else {
					newOrg.Bitrix = "Не Битрикс, открывается"

				}
			} else {
				newOrg.Bitrix = "Не Битрикс, открывается"
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
	//fmt.Printf("\r[%v:%v] Result of: %v is [%v]\t\t\t\t\t", complete, line, newOrg.Site, newOrg.Bitrix)
	active--
}

func Percentage() {
	fmt.Println("Rows loaded, checking...")
	sc := 0
	for complete < count {
		per := int((float64(complete ) / float64(count)) * 100)
		for i := 0; i < sc; i++ {
			fmt.Print("\b")
		}
		str := fmt.Sprintf("Complete: %v%% | %v\\%v | Bitrixes: %v", per, complete, count, bitrixes)
		sc = len(str)
		fmt.Print(str)
		time.Sleep(1 * time.Second)
	}
}
