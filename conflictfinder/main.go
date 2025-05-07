package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Assessment struct {
	Type              string `json:"type,omitempty"`
	Number            string `json:"number,omitempty"`
	Version           int    `json:"version,omitempty"`
	TNV               string `json:"tnv,omitempty"`
	Level             int    `json:"level,omitempty"`
	Credits           int    `json:"credits,omitempty"`
	Weighting         any    `json:"weighting,omitempty"`
	Points            any    `json:"points,omitempty"`
	Title             string `json:"title,omitempty"`
	Description       any    `json:"description,omitempty"`
	Purpose           any    `json:"purpose,omitempty"`
	SchoolRef         any    `json:"schoolref,omitempty"`
	Subfield          string `json:"subfield,omitempty"`
	Internalexternal  string `json:"internalexternal,omitempty"`
	ListenerUpdatedAt string
}

func (a *Assessment) CreateTNV() {
	tnv := strings.Join([]string{a.Type, a.Number, strconv.Itoa(a.Version)}, "_")
	// fmt.Println(tnv)
	a.TNV = tnv
}

type Attendance struct {
	ID     int    `json:"id,omitempty"`
	Nsn    string `json:"nsn,omitempty"`
	Values []struct {
		Date  string `json:"date,omitempty"`
		Codes string `json:"codes,omitempty"`
		Alt   string `json:"alt,omitempty"`
		Hdu   int    `json:"hdu,omitempty"`
		Hdj   int    `json:"hdj,omitempty"`
		Hdp   int    `json:"hdp,omitempty"`
	} `json:"values,omitempty"`
}

type Pastoral struct {
	ID                int    `json:"id,omitempty"`
	Nsn               string `json:"nsn,omitempty"`
	Type              string `json:"type,omitempty"`
	Ref               int    `json:"ref,omitempty"`
	Reason            string `json:"reason,omitempty"`
	ReasonPB          any    `json:"reasonPB,omitempty"`
	Motivation        any    `json:"motivation,omitempty"`
	MotivationPB      any    `json:"motivationPB,omitempty"`
	Location          any    `json:"location,omitempty"`
	LocationPB        any    `json:"locationPB,omitempty"`
	Othersinvolved    any    `json:"othersinvolved,omitempty"`
	Action1           string `json:"action1,omitempty"`
	Action2           any    `json:"action2,omitempty"`
	Action3           any    `json:"action3,omitempty"`
	ActionPB1         any    `json:"actionPB1,omitempty"`
	ActionPB2         any    `json:"actionPB2,omitempty"`
	ActionPB3         any    `json:"actionPB3,omitempty"`
	Teacher           string `json:"teacher,omitempty"`
	Points            int    `json:"points,omitempty"`
	Demerits          int    `json:"demerits,omitempty"`
	Dateevent         string `json:"dateevent,omitempty"`
	Timeevent         string `json:"timeevent,omitempty"`
	Datedue           string `json:"datedue,omitempty"`
	Duestatus         string `json:"duestatus,omitempty"`
	ListenerUpdatedAt string
}

// Check for the number of records in a JSON file that have multiple entries containing a field that should be unique, i.e. updates rather than inserts, to ensure that the number of rows in a database table after insert matches the number of unique entries in the JSON file
func main() {
	// Load JSON from file
	data, err := os.ReadFile("../test/actual-requests/attend_test.json")
	if err != nil {
		log.Fatalf("failed to read attendance JSON: %v", err)
	}

	var records struct {
		Data []Attendance `json:"data"`
	}
	if err := json.Unmarshal(data, &records); err != nil {
		log.Fatalf("failed to parse attendance JSON: %v", err)
	}

	// Track seen (att_id, date) pairs
	seen := make(map[int]map[string]bool)
	conflicts := 0

	for _, record := range records.Data {
		if seen[record.ID] == nil {
			seen[record.ID] = make(map[string]bool)
		}
		for _, v := range record.Values {
			if seen[record.ID][v.Date] {
				fmt.Printf("conflict in attendance JSON: att_id=%d date=%s\n", record.ID, v.Date)
				conflicts++
			}
			seen[record.ID][v.Date] = true
		}
	}

	fmt.Printf("total attendance conflicts: %d\n", conflicts)

	data, err = os.ReadFile("../test/actual-requests/assess_test.json")
	if err != nil {
		log.Fatalf("failed to read assessment JSON: %v", err)
	}

	var assRecords struct {
		Data []Assessment `json:"data"`
	}
	if err := json.Unmarshal(data, &assRecords); err != nil {
		log.Fatalf("failed to parse assessment JSON: %v", err)
	}

	// Track seen assessment TNVs
	seenAss := make(map[string]bool)
	assConflicts := 0

	for _, record := range assRecords.Data {
		// fmt.Println(record.Title)
		record.CreateTNV()
		// fmt.Println(record.TNV)
		if seenAss[record.TNV] {
			fmt.Printf("conflict in assessment JSON: ass_tnv=%s\n", record.TNV)
			assConflicts++
		}
		seenAss[record.TNV] = true
	}

	fmt.Printf("total assessment conflicts: %d\n", assConflicts)

	data, err = os.ReadFile("../test/actual-requests/pastoral_test.json")
	if err != nil {
		log.Fatalf("failed to read pastoral JSON: %v", err)
	}

	var pasRecords struct {
		Data []Pastoral `json:"data"`
	}
	if err := json.Unmarshal(data, &pasRecords); err != nil {
		log.Fatalf("failed to parse pastoral JSON: %v", err)
	}

	// Track seen assessment TNVs
	seenPas := make(map[string]bool)
	pasConflicts := 0

	for _, record := range pasRecords.Data {
		identifier := strconv.Itoa(record.ID) + "_" + record.Type + "_" + strconv.Itoa(record.Ref)
		if seenPas[identifier] {
			fmt.Printf("conflict in pastoral JSON: id_type_ref=%s\n", identifier)
			pasConflicts++
		}
		seenPas[identifier] = true
	}

	fmt.Printf("total pastoral conflicts: %d\n", pasConflicts)
}
