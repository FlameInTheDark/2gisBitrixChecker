package main

import "sync"

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
	ToSave        bool   `xlsx:"-"`
}

// OrgContainer data container for concurrency
type OrgContainer struct {
	mx sync.Mutex
	m  map[int]*Organization
}

// MakeContainer creates new OrgContainer
func MakeContainer() OrgContainer {
	return OrgContainer{m: make(map[int]*Organization)}
}

// Store saves data to map in OrgContainer
func (o *OrgContainer) Store(key int, org *Organization) {
	o.mx.Lock()
	o.m[key] = org
	o.mx.Unlock()
}

// Get returns Organization from OrgContainer by key
func (o *OrgContainer) Get(key int) *Organization {
	return o.m[key]
}

// Len returns size of OrgContainer map
func (o *OrgContainer) Len() int {
	return len(o.m)
}

// Map returns pointer to map in OrgContainer
func (o *OrgContainer) Map() *map[int]*Organization {
	return &o.m
}

type Company struct {
	Title        string   `json:"TITLE"`
	CompanyType  string   `json:"COMPANY_TYPE"`
	Opened       string   `json:"OPENED"`
	AssignedByID string   `json:"ASSIGNED_BY_ID"`
	Phone        []Phone  `json:"PHONE"`
	Site         []Site   `json:"WEB"`
	Emails       []Email  `json:"EMAIL"`
	From         string   `json:"UF_CRM_5AF5A321D405E"`
	Country      []string `json:"UF_CRM_1551766188"`
	License      []string `json:"UF_CRM_1551766310"`
}

type Phone struct {
	Value     string `json:"VALUE"`
	ValueType string `json:"VALUE_TYPE"`
}

type Site struct {
	Value     string `json:"VALUE"`
	ValueType string `json:"VALUE_TYPE"`
}

type Email struct {
	Value     string `json:"VALUE"`
	ValueType string `json:"VALUE_TYPE"`
}

type Fields struct {
	Fields Company `json:"fields"`
}
