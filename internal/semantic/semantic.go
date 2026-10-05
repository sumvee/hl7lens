// Package semantic joins a parsed HL7 message (internal/hl7) with the
// dictionary (internal/dict) to produce a human-meaningful annotation
// tree. The tree is rendering-agnostic so it can feed both the `explain`
// text output and the `view` TUI.
package semantic

import (
	"fmt"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

// ComponentView is one annotated component of a composite field value.
type ComponentView struct {
	Path     string // e.g. ".1"
	Name     string // dictionary name, empty if unknown
	Raw      string
	Decoded  string // table meaning, empty if none
	Table    string // table id used for the decode, empty if none
	DataType string
}

// FieldView is one annotated field value (one repetition of a field).
type FieldView struct {
	Path       string // e.g. "PID-5" or "PID-3[2]"
	Name       string
	DataType   string
	Raw        string
	Decoded    string // table meaning for a coded primitive field
	Table      string // table id used for the decode, empty if none
	Components []ComponentView
}

// SegmentView is an annotated segment.
type SegmentView struct {
	Name        string
	Description string
	Known       bool // whether the dictionary defines this segment
	Fields      []FieldView
}

// MessageView is the full annotation tree for a message.
type MessageView struct {
	Version  dict.Version
	Segments []SegmentView
}

// Build annotates a parsed message against a dictionary.
func Build(msg *hl7.Message, d *dict.Dictionary) MessageView {
	mv := MessageView{Version: d.Version}
	for _, seg := range msg.Segments {
		mv.Segments = append(mv.Segments, buildSegment(seg, msg.Delims, d))
	}
	return mv
}

func buildSegment(seg hl7.Segment, delims hl7.Delimiters, d *dict.Dictionary) SegmentView {
	segDef, known := d.Segments[seg.Name]
	sv := SegmentView{Name: seg.Name, Known: known}
	if known {
		sv.Description = segDef.Description
	}

	for i, field := range seg.Fields {
		if field.IsEmpty() {
			continue
		}
		seq := i + 1
		fieldDef, hasDef := d.Field(seg.Name, seq)

		reps := field.Repetitions
		for ri, rep := range reps {
			path := fmt.Sprintf("%s-%d", seg.Name, seq)
			if len(reps) > 1 {
				path = fmt.Sprintf("%s-%d[%d]", seg.Name, seq, ri+1)
			}
			fv := FieldView{Path: path, Raw: rep.Raw(delims)}
			if hasDef {
				fv.Name = fieldDef.Name
				fv.DataType = fieldDef.DataType
			}
			dt, hasDT := d.DataTypes[fv.DataType]

			switch {
			case hasDef && fieldDef.Table != "" && (!hasDT || dt.Primitive):
				// Coded primitive field: decode the whole value.
				if meaning, ok := d.Decode(fieldDef.Table, fv.Raw); ok {
					fv.Decoded = meaning
					fv.Table = fieldDef.Table
				}
			case hasDT && !dt.Primitive && len(dt.Components) > 0:
				// Composite with a modeled layout: break into components. A
				// coded composite (e.g. CE/CWE bound to a user table) carries
				// its code in the first component, so the field's table
				// decodes .1. Composites that are not yet broken down (no
				// components) render opaquely as the field value only.
				fieldTable := ""
				if hasDef {
					fieldTable = fieldDef.Table
				}
				fv.Components = buildComponents(rep, delims, dt, d, fieldTable)
			}
			sv.Fields = append(sv.Fields, fv)
		}
	}
	return sv
}

func buildComponents(rep hl7.Repetition, delims hl7.Delimiters, dt dict.DataTypeDef, d *dict.Dictionary, fieldTable string) []ComponentView {
	var out []ComponentView
	for ci, comp := range rep.Components {
		raw := comp.Raw(delims)
		if raw == "" {
			continue
		}
		cv := ComponentView{Path: fmt.Sprintf(".%d", ci+1), Raw: raw}
		table := ""
		if ci < len(dt.Components) {
			cd := dt.Components[ci]
			cv.Name = cd.Name
			cv.DataType = cd.DataType
			table = cd.Table
		}
		// The first component of a coded composite inherits the field's
		// table binding when it has none of its own.
		if table == "" && ci == 0 {
			table = fieldTable
		}
		if table != "" {
			if meaning, ok := d.Decode(table, raw); ok {
				cv.Decoded = meaning
				cv.Table = table
			}
		}
		out = append(out, cv)
	}
	return out
}
