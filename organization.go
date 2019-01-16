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
}

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