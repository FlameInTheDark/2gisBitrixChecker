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
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Check if multiple sites in one row
	if strings.Contains(newOrg.Site, ",") {
		siteArr := strings.Split(newOrg.Site, ",")
		newOrg.Site = strings.Replace(siteArr[0], " ", "", -1)
	}

	resp, errGet1 := client.Get(fmt.Sprintf("https://%s/bitrix/themes/.default/modules.css", newOrg.Site))
	resp2, errGet2 := client.Get(fmt.Sprintf("http://%s/bitrix/themes/.default/modules.css", newOrg.Site))
	if errGet1 == nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			bodyBytes, err2 := ioutil.ReadAll(resp.Body)
			if err2 == nil {
				bodyString := string(bodyBytes)
				if len(bodyString) == 0 {
					if errGet2 != nil {
						defer resp2.Body.Close()
						if resp2.StatusCode == http.StatusOK {
							body2Bytes, err3 := ioutil.ReadAll(resp2.Body)
							if err3 == nil {
								bodyString = string(body2Bytes)
							}
						}
					}
				}
				if strings.Contains(bodyString, "sale") {
					newOrg.Bitrix = "sale"
					newOrg.ToSave = true
					bitrixes++
				} else if strings.Contains(bodyString, "bitrix") {
					newOrg.Bitrix = "bitrix"
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
	org.Store(rowNum, newOrg)
	checked++
	complete++
	active--
}

// Percentage shows percentage of progress in console
func Percentage() {
	fmt.Println("Rows loaded, checking...")
	sc := 0
	for complete < count {
		per := int((float64(complete) / float64(count)) * 100)
		for i := 0; i < sc; i++ {
			fmt.Print("\b")
		}
		str := fmt.Sprintf("Complete: %d%% | %d\\%d | Bitrixes: %d   ", per, complete, count, bitrixes)
		sc = len(str)
		fmt.Print(str)
		time.Sleep(time.Second)
	}
}
