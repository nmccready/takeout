package model

import (
	"encoding/csv"
	"fmt"
	"strings"
)

type CsvRows [][]string

// map to row
type CsvMapRow map[string][]string

func (c CsvRows) ToMap() CsvMapRow {
	m := CsvMapRow{}
	titleCounter := map[string]int{}

	for _, row := range c {
		title := row[0]
		_, hasValue := m[title]
		if !hasValue {
			m[title] = row
			continue
		}
		titleCounter[title]++
		title = fmt.Sprintf("%s(%d)", title, titleCounter[title])
		m[title] = row
	}
	return m
}

func ReadAllCsvRows(body string) (CsvRows, error) {
	r := csv.NewReader(strings.NewReader(body))
	// r.Comma = ';'
	// r.Comment = '#'
	// r.TrimLeadingSpace = true
	// r.LazyQuotes = true
	return r.ReadAll()
}
