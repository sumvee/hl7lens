package hl7

import "strings"

// Raw reassembles a component from its subcomponents using the message
// delimiters.
func (c Component) Raw(d Delimiters) string {
	return strings.Join(c.Subcomponents, string(d.Subcomponent))
}

// Raw reassembles a repetition from its components.
func (r Repetition) Raw(d Delimiters) string {
	parts := make([]string, len(r.Components))
	for i, c := range r.Components {
		parts[i] = c.Raw(d)
	}
	return strings.Join(parts, string(d.Component))
}

// Raw reassembles a field (all repetitions) using the message delimiters.
func (f Field) Raw(d Delimiters) string {
	parts := make([]string, len(f.Repetitions))
	for i, r := range f.Repetitions {
		parts[i] = r.Raw(d)
	}
	return strings.Join(parts, string(d.Repetition))
}

// IsEmpty reports whether a field carries no data.
func (f Field) IsEmpty() bool {
	for _, r := range f.Repetitions {
		for _, c := range r.Components {
			for _, s := range c.Subcomponents {
				if s != "" {
					return false
				}
			}
		}
	}
	return true
}

// Raw reassembles a whole segment line using the message delimiters.
// MSH is handled specially: MSH-1 is the field separator and MSH-2 the
// encoding characters.
func (s Segment) Raw(d Delimiters) string {
	var b strings.Builder
	if s.Name == "MSH" {
		b.WriteString("MSH")
		b.WriteByte(d.Field)
		if len(s.Fields) > 1 {
			b.WriteString(s.Fields[1].Raw(d)) // encoding characters, literal
		}
		for i := 2; i < len(s.Fields); i++ {
			b.WriteByte(d.Field)
			b.WriteString(s.Fields[i].Raw(d))
		}
		return b.String()
	}
	b.WriteString(s.Name)
	for _, f := range s.Fields {
		b.WriteByte(d.Field)
		b.WriteString(f.Raw(d))
	}
	return b.String()
}

// Field returns segment field n (1-based) if present.
func (s Segment) Field(n int) (Field, bool) {
	i := n - 1
	if i < 0 || i >= len(s.Fields) {
		return Field{}, false
	}
	return s.Fields[i], true
}
