package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

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

func main() {
	// Load JSON from file
	data, err := os.ReadFile("../test/actual-requests/attend_test.json")
	if err != nil {
		log.Fatalf("failed to read JSON: %v", err)
	}

	var records struct {
		Data []Attendance `json:"data"`
	}
	if err := json.Unmarshal(data, &records); err != nil {
		log.Fatalf("failed to parse JSON: %v", err)
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
				fmt.Printf("Conflict in JSON: att_id=%d date=%s\n", record.ID, v.Date)
				conflicts++
			}
			seen[record.ID][v.Date] = true
		}
	}

	fmt.Printf("total conflicts: %d\n", conflicts)
}
