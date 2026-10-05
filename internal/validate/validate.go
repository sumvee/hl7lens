// Package validate checks an HL7 v2.x message for conformance:
//   T1 structural : required fields, cardinality, table membership, and
//                   basic data-type format, driven by the dictionary.
//   T2 grammar    : required segments, segment cardinality, and order,
//                   driven by the abstract message definition.
package validate

import (
	"fmt"
	"strings"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

// Level is the severity of a finding.
type Level int

const (
	// Warning does not fail validation (exit 0).
	Warning Level = iota
	// Error fails validation (non-zero exit).
	Error
)

func (l Level) String() string {
	if l == Error {
		return "ERROR"
	}
	return "WARN"
}

// Finding is one conformance problem.
type Finding struct {
	Level   Level
	Path    string // segment or field path, e.g. "PID-5"
	Message string
}

// Report is the result of validation.
type Report struct {
	Findings []Finding
}

func (r *Report) err(path, msg string)  { r.Findings = append(r.Findings, Finding{Error, path, msg}) }
func (r *Report) warn(path, msg string) { r.Findings = append(r.Findings, Finding{Warning, path, msg}) }

// HasErrors reports whether any finding is an error.
func (r Report) HasErrors() bool {
	for _, f := range r.Findings {
		if f.Level == Error {
			return true
		}
	}
	return false
}

// Counts returns the number of errors and warnings.
func (r Report) Counts() (errors, warnings int) {
	for _, f := range r.Findings {
		if f.Level == Error {
			errors++
		} else {
			warnings++
		}
	}
	return
}

// Validate runs T1 and T2 against a message.
func Validate(msg *hl7.Message, d *dict.Dictionary) Report {
	var r Report
	validateStructure(&r, msg, d) // T1
	validateGrammar(&r, msg, d)   // T2
	return r
}

// ---- T1 structural ----

func validateStructure(r *Report, msg *hl7.Message, d *dict.Dictionary) {
	for _, seg := range msg.Segments {
		segDef, known := d.Segments[seg.Name]
		if !known {
			continue // cannot structurally validate an unknown segment
		}
		for _, fd := range segDef.Fields {
			path := fmt.Sprintf("%s-%d", seg.Name, fd.Seq)
			field, present := seg.Field(fd.Seq)
			empty := !present || field.IsEmpty()

			if fd.Optionality == "R" && empty {
				r.err(path, fmt.Sprintf("%s is required but empty", fd.Name))
				continue
			}
			if empty {
				continue
			}
			if !fd.Repeats && len(field.Repetitions) > 1 {
				r.err(path, fmt.Sprintf("%s must not repeat (found %d repetitions)", fd.Name, len(field.Repetitions)))
			}
			checkFieldValue(r, d, path, fd, field, msg.Delims)
		}
	}
}

func checkFieldValue(r *Report, d *dict.Dictionary, path string, fd dict.FieldDef, field hl7.Field, delims hl7.Delimiters) {
	dt, hasDT := d.DataTypes[fd.DataType]
	primitive := !hasDT || dt.Primitive

	// Table membership for coded primitive fields, when the table is loaded.
	if primitive && fd.Table != "" {
		if _, loaded := d.Tables[fd.Table]; loaded {
			val := field.Repetitions[0].Raw(delims)
			if _, ok := d.Decode(fd.Table, val); !ok && val != "" {
				r.warn(path, fmt.Sprintf("value %q not in table %s (%s)", val, fd.Table, fd.Name))
			}
		}
	}

	// Numeric format for NM/SI.
	if fd.DataType == "NM" || fd.DataType == "SI" {
		val := field.Repetitions[0].Raw(delims)
		if val != "" && !isNumeric(val) {
			r.warn(path, fmt.Sprintf("%s expects a number, got %q", fd.DataType, val))
		}
	}
}

// ---- T2 grammar ----

func validateGrammar(r *Report, msg *hl7.Message, d *dict.Dictionary) {
	if len(msg.Segments) == 0 || msg.Segments[0].Name != "MSH" {
		r.err("MSH", "message must begin with an MSH segment")
	}

	code, _ := msg.Get("MSH-9.1")
	trigger, _ := msg.Get("MSH-9.2")
	hint, _ := msg.Get("MSH-9.3")

	def, ok := d.StructureFor(code, trigger, hint)
	if !ok {
		label := strings.TrimRight(code+"^"+trigger, "^")
		r.warn("MSH-9", fmt.Sprintf("no grammar loaded for message type %q; grammar checks skipped", label))
		return
	}

	counts := map[string]int{}
	for _, seg := range msg.Segments {
		counts[seg.Name]++
	}
	ruleIdx := map[string]int{}
	for i, rule := range def.Segments {
		ruleIdx[rule.Segment] = i
		switch {
		case rule.Optionality == "R" && counts[rule.Segment] == 0:
			r.err(rule.Segment, fmt.Sprintf("required segment %s is missing (message %s)", rule.Segment, def.Structure))
		case !rule.Repeats && counts[rule.Segment] > 1:
			r.err(rule.Segment, fmt.Sprintf("segment %s must not repeat (found %d)", rule.Segment, counts[rule.Segment]))
		}
	}

	// Order: segments that the grammar knows must appear in grammar order.
	// Only checked for flat message structures; grouped messages (ORU)
	// legitimately recur segments across groups.
	if !def.Ordered {
		return
	}
	lastIdx := -1
	for _, seg := range msg.Segments {
		idx, known := ruleIdx[seg.Name]
		if !known {
			continue
		}
		if idx < lastIdx {
			r.warn(seg.Name, fmt.Sprintf("segment %s appears out of order for %s", seg.Name, def.Structure))
		}
		if idx > lastIdx {
			lastIdx = idx
		}
	}
}

func isNumeric(s string) bool {
	dot := false
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r == '-' && i == 0:
		case r == '.' && !dot:
			dot = true
		default:
			return false
		}
	}
	return s != "" && s != "-"
}
