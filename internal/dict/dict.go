// Package dict is the semantic layer: it maps structural positions
// (segment, field, component) and coded values to human meaning for a
// given HL7 v2.x version.
//
// Data provenance (see repo NOTICE):
//   - tables / vocabulary: HL7 Terminology (CC0)
//   - segment/datatype/message structure: hand-curated launch subset,
//     cross-checked against the Apache-2.0 HL7 v2-to-FHIR project and
//     verified against the published v2.5.1 attribute tables.
//
// No HAPI (GPLv2/MPL) data and no transcribed HL7 spec prose are embedded.
package dict

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// Version identifies a loaded dictionary dataset (e.g. "2.5.1").
type Version string

// FieldDef describes one field position within a segment.
// Fields are addressed so that a segment's Fields[i] describes <SEG>-(i+1).
type FieldDef struct {
	Seq         int    `json:"seq"`
	Name        string `json:"name"`
	DataType    string `json:"datatype"`
	Optionality string `json:"optionality"` // R required, O optional, C conditional
	Repeats     bool   `json:"repeats"`
	Table       string `json:"table"` // bound HL7 table id, empty if none
}

// SegmentDef is the ordered field definitions for a segment.
type SegmentDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Fields      []FieldDef `json:"fields"`
}

// ComponentDef describes one component within a composite data type.
type ComponentDef struct {
	Seq      int    `json:"seq"`
	Name     string `json:"name"`
	DataType string `json:"datatype"`
	Table    string `json:"table"`
}

// DataTypeDef describes a data type: primitive, or a composite with
// ordered components.
type DataTypeDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Primitive   bool           `json:"primitive"`
	Components  []ComponentDef `json:"components"`
}

// Table maps a coded value to its display meaning (HL7 table contents).
type Table struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Entries map[string]string `json:"entries"`
}

// SegmentRule is one entry in an abstract message definition (grammar).
type SegmentRule struct {
	Segment     string `json:"segment"`
	Optionality string `json:"optionality"` // R required, O optional
	Repeats     bool   `json:"repeats"`
}

// MessageDef is the abstract message definition (segment grammar) for a
// message structure such as ADT_A01.
type MessageDef struct {
	Structure string        `json:"structure"`
	Triggers  []string      `json:"triggers"`
	Segments  []SegmentRule `json:"segments"`
	// Ordered enables segment-order checking. It is meaningful only for
	// flat message structures (e.g. ADT); messages with repeating segment
	// groups (e.g. ORU) leave it false, since a flat rule list cannot
	// express a segment that legitimately recurs across groups.
	Ordered bool `json:"ordered"`
}

// Dictionary is the full semantic dataset for one version.
type Dictionary struct {
	Version   Version
	Segments  map[string]SegmentDef
	DataTypes map[string]DataTypeDef
	Tables    map[string]Table
	Messages  map[string]MessageDef // keyed by structure id (e.g. ADT_A01)
}

//go:embed all:data
var dataFS embed.FS

// Load returns the embedded dictionary for a version (e.g. "2.5.1").
func Load(v Version) (*Dictionary, error) {
	base := path.Join("data", string(v))
	if _, err := fs.Stat(dataFS, base); err != nil {
		return nil, fmt.Errorf("dict: no embedded dataset for version %q", v)
	}
	d := &Dictionary{
		Version:   v,
		Segments:  map[string]SegmentDef{},
		DataTypes: map[string]DataTypeDef{},
		Tables:    map[string]Table{},
		Messages:  map[string]MessageDef{},
	}
	if err := loadSegments(d, path.Join(base, "segments")); err != nil {
		return nil, err
	}
	if err := loadDataTypes(d, path.Join(base, "datatypes")); err != nil {
		return nil, err
	}
	if err := loadTables(d, path.Join(base, "tables")); err != nil {
		return nil, err
	}
	if err := loadMessages(d, path.Join(base, "messages")); err != nil {
		return nil, err
	}
	return d, nil
}

func loadMessages(d *Dictionary, dir string) error {
	files, _ := jsonFiles(dir)
	for _, f := range files {
		var m MessageDef
		if err := readJSON(f, &m); err != nil {
			return err
		}
		d.Messages[m.Structure] = m
	}
	return nil
}

// StructureFor resolves a message structure from MSH-9: the structure hint
// (MSH-9.3, e.g. "ADT_A01") when known, otherwise the first structure
// whose trigger list contains the trigger event (MSH-9.2).
func (d *Dictionary) StructureFor(code, trigger, hint string) (MessageDef, bool) {
	if hint != "" {
		if m, ok := d.Messages[hint]; ok {
			return m, true
		}
	}
	for _, m := range d.Messages {
		for _, t := range m.Triggers {
			if t == trigger {
				return m, true
			}
		}
	}
	return MessageDef{}, false
}

func jsonFiles(dir string) ([]string, error) {
	entries, err := fs.ReadDir(dataFS, dir)
	if err != nil {
		return nil, nil // a missing optional dir is not an error
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			out = append(out, path.Join(dir, e.Name()))
		}
	}
	return out, nil
}

func loadSegments(d *Dictionary, dir string) error {
	files, _ := jsonFiles(dir)
	for _, f := range files {
		var s SegmentDef
		if err := readJSON(f, &s); err != nil {
			return err
		}
		d.Segments[s.Name] = s
	}
	return nil
}

func loadDataTypes(d *Dictionary, dir string) error {
	files, _ := jsonFiles(dir)
	for _, f := range files {
		var dts []DataTypeDef
		if err := readJSON(f, &dts); err != nil {
			return err
		}
		for _, dt := range dts {
			d.DataTypes[dt.Name] = dt
		}
	}
	return nil
}

func loadTables(d *Dictionary, dir string) error {
	files, _ := jsonFiles(dir)
	for _, f := range files {
		var ts []Table
		if err := readJSON(f, &ts); err != nil {
			return err
		}
		for _, t := range ts {
			d.Tables[t.ID] = t
		}
	}
	return nil
}

func readJSON(file string, v any) error {
	raw, err := dataFS.ReadFile(file)
	if err != nil {
		return fmt.Errorf("dict: read %s: %w", file, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("dict: parse %s: %w", file, err)
	}
	return nil
}

// Field returns the definition of <segment>-<seq> (1-based), if known.
func (d *Dictionary) Field(segment string, seq int) (FieldDef, bool) {
	s, ok := d.Segments[segment]
	if !ok {
		return FieldDef{}, false
	}
	i := seq - 1
	if i < 0 || i >= len(s.Fields) {
		return FieldDef{}, false
	}
	return s.Fields[i], true
}

// Decode returns the display meaning of a coded value in a table.
func (d *Dictionary) Decode(tableID, code string) (string, bool) {
	t, ok := d.Tables[tableID]
	if !ok {
		return "", false
	}
	v, ok := t.Entries[code]
	return v, ok
}
