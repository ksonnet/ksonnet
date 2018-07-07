// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package table

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
)

const (
	// sepChar is the character used to separate the header from the content in a table.
	sepChar = "="
)

// Format is the output format.
type Format int

const (
	// FormatTable prints a table.
	FormatTable Format = iota
	// FormatJSON prints JSON.
	FormatJSON
)

// DefaultFormat is the default format for output. It is a table.
const DefaultFormat = FormatTable

// DetectFormat detects a format from a string.
func DetectFormat(formatName string) (Format, error) {
	switch formatName {
	case "json":
		return FormatJSON, nil
	case "", "table":
		return FormatTable, nil
	default:
		return Format(-1), errors.Errorf("unknown output format %q", formatName)
	}
}

// Table creates an output table. Use the New constructor to ensure
// defaults are set properly.
type Table struct {
	Name   string
	Format Format
	w      io.Writer

	printf func(w io.Writer, format string, a ...interface{}) (int, error)

	header []string
	rows   [][]string
}

// New creates an instance of table.
func New(name string, w io.Writer) *Table {
	return &Table{
		Name:   name,
		Format: DefaultFormat,
		printf: fmt.Fprintf,
		w:      w,
	}
}

// SetHeader sets the header for the table.
func (t *Table) SetHeader(columns []string) {
	t.header = columns
}

// SetFormat sets sets the output format.
func (t *Table) SetFormat(f Format) {
	t.Format = f
}

// Append appends a row to the table.
func (t *Table) Append(row []string) {
	t.rows = append(t.rows, row)
}

// AppendBulk appends multiple rows to the table.
func (t *Table) AppendBulk(rows [][]string) {
	t.rows = append(t.rows, rows...)
}

// Render writes the output to the table's writer.
func (t *Table) Render() error {
	if t.w == nil {
		return errors.New("writer is nil")
	}

	switch t.Format {
	default:
		return t.renderTable()
	case FormatTable:
		return t.renderTable()
	case FormatJSON:
		return t.renderJSON()
	}
}

// jsonOutput is the structure for printing JSON output.
type jsonOutput struct {
	Kind string              `json:"kind"`
	Data []map[string]string `json:"data"`
}

func (t *Table) renderJSON() error {
	if len(t.header) == 0 {
		return errors.New("headers aren't defined for output")
	}

	out := make([]map[string]string, 0)
	for _, row := range t.rows {
		m := make(map[string]string)
		if len(t.header) != len(row) {
			return errors.New("header length doesn't match row length")
		}

		for i, header := range t.header {
			m[header] = row[i]
		}

		out = append(out, m)
	}

	encoder := json.NewEncoder(t.w)
	encoder.SetIndent("", "\t")

	jo := jsonOutput{
		Kind: t.Name,
		Data: out,
	}

	return encoder.Encode(&jo)
}

func (t *Table) renderTable() error {
	var output [][]string

	if len(t.header) > 0 {
		headerRow := make([]string, len(t.header), len(t.header))
		sepRow := make([]string, len(t.header), len(t.header))

		for i := range t.header {
			sepLen := len(t.header[i])
			headerRow[i] = strings.ToUpper(t.header[i])
			sepRow[i] = strings.Repeat(sepChar, sepLen)
		}

		output = append(output, headerRow, sepRow)
	}

	output = append(output, t.rows...)

	counts := colLens(output)

	// print rows
	for _, row := range output {
		hl := len(t.header)
		if hl > 0 && (hl != len(row)) {
			return errors.New("header length doesn't match row length")
		}

		var parts []string
		for i, col := range row {
			val := col
			if i < len(row)-1 {
				format := fmt.Sprintf("%%-%ds", counts[i])
				val = fmt.Sprintf(format, col)
			}
			parts = append(parts, val)

		}

		_, err := t.printf(t.w, "%s\n", strings.TrimSpace(strings.Join(parts, " ")))
		if err != nil {
			return errors.Wrap(err, "writing row to table")
		}
	}

	return nil
}

func colLens(rows [][]string) []int {
	// count the number of columns
	colCount := 0
	for _, row := range rows {
		if l := len(row); l > colCount {
			colCount = l
		}
	}

	// get the max length for each column
	counts := make([]int, colCount, colCount)
	for _, row := range rows {
		for i := range row {
			if l := len(row[i]); l > counts[i] {
				counts[i] = l
			}
		}
	}

	return counts
}
