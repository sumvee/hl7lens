// Package hl7 is the dependency-free core: it parses an HL7 v2.x message
// into an addressable tree. It carries no opinion about meaning; the
// dictionary layer (internal/dict) supplies that.
package hl7

import (
	"errors"
	"strings"
)

// Delimiters are the encoding characters a message declares in MSH-1/MSH-2.
// Defaults follow the HL7 recommendation: | ^ ~ \ &
type Delimiters struct {
	Field        byte // MSH-1, default '|'
	Component    byte // MSH-2 pos 1, default '^'
	Repetition   byte // MSH-2 pos 2, default '~'
	Escape       byte // MSH-2 pos 3, default '\'
	Subcomponent byte // MSH-2 pos 4, default '&'
}

// DefaultDelimiters returns the HL7-recommended encoding characters.
func DefaultDelimiters() Delimiters {
	return Delimiters{Field: '|', Component: '^', Repetition: '~', Escape: '\\', Subcomponent: '&'}
}

// Message is an ordered list of segments plus the delimiters in force.
type Message struct {
	Delims   Delimiters
	Segments []Segment
}

// Segment is a 3-char name (e.g. "PID") and its fields, in order.
// For MSH, Fields[0] is MSH-1 (the field separator itself is not stored
// as data; Fields are indexed so that Fields[i] == <SEG>-(i+1)).
type Segment struct {
	Name   string
	Fields []Field
}

// Field holds one or more repetitions ( typically one).
type Field struct {
	Repetitions []Repetition
}

// Repetition is one value, split into components and subcomponents.
type Repetition struct {
	Components []Component
}

// Component is split into subcomponents (usually one).
type Component struct {
	Subcomponents []string
}

var (
	// ErrEmpty is returned when there is nothing to parse.
	ErrEmpty = errors.New("empty message")
	// ErrNoMSH is returned when the message does not begin with MSH.
	ErrNoMSH = errors.New("message does not start with an MSH segment")
)

// Parse reads a raw HL7 v2.x message. It derives the delimiters from the
// MSH segment, tolerates \r, \n, and \r\n line endings, and skips blank
// trailing lines.
func Parse(raw string) (*Message, error) {
	raw = strings.TrimRight(raw, "\r\n ")
	if raw == "" {
		return nil, ErrEmpty
	}
	if !strings.HasPrefix(raw, "MSH") {
		return nil, ErrNoMSH
	}
	if len(raw) < 8 {
		return nil, ErrNoMSH
	}

	d := Delimiters{
		Field:        raw[3],
		Component:    raw[4],
		Repetition:   raw[5],
		Escape:       raw[6],
		Subcomponent: raw[7],
	}

	lines := splitLines(raw)
	msg := &Message{Delims: d, Segments: make([]Segment, 0, len(lines))}
	for _, line := range lines {
		if line == "" {
			continue
		}
		msg.Segments = append(msg.Segments, parseSegment(line, d))
	}
	return msg, nil
}

func splitLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\r")
	raw = strings.ReplaceAll(raw, "\n", "\r")
	return strings.Split(raw, "\r")
}

func parseSegment(line string, d Delimiters) Segment {
	name := line
	if len(line) >= 3 {
		name = line[:3]
	}
	seg := Segment{Name: name}

	rawFields := strings.Split(line, string(d.Field))

	// MSH is special: MSH-1 is the field separator and MSH-2 is the
	// encoding characters, both of which must be kept literal (MSH-2
	// contains the component/repetition/subcomponent delimiters and must
	// not be split on them). rawFields[0] is "MSH"; we re-seat so that
	// Fields index aligns with <SEG>-n addressing.
	if name == "MSH" {
		seg.Fields = append(seg.Fields, literalField(string(d.Field)))
		if len(rawFields) > 1 {
			seg.Fields = append(seg.Fields, literalField(rawFields[1]))
		}
		for _, rf := range rawFields[2:] {
			seg.Fields = append(seg.Fields, parseField(rf, d))
		}
		return seg
	}

	for _, rf := range rawFields[1:] {
		seg.Fields = append(seg.Fields, parseField(rf, d))
	}
	return seg
}

// literalField wraps a raw string as a single-value field with no further
// splitting (used for MSH-1 and MSH-2).
func literalField(raw string) Field {
	return Field{Repetitions: []Repetition{{Components: []Component{{Subcomponents: []string{raw}}}}}}
}

func parseField(raw string, d Delimiters) Field {
	reps := strings.Split(raw, string(d.Repetition))
	f := Field{Repetitions: make([]Repetition, 0, len(reps))}
	for _, rep := range reps {
		f.Repetitions = append(f.Repetitions, parseRepetition(rep, d))
	}
	return f
}

func parseRepetition(raw string, d Delimiters) Repetition {
	comps := strings.Split(raw, string(d.Component))
	r := Repetition{Components: make([]Component, 0, len(comps))}
	for _, c := range comps {
		subs := strings.Split(c, string(d.Subcomponent))
		r.Components = append(r.Components, Component{Subcomponents: subs})
	}
	return r
}
