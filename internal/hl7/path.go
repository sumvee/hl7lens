package hl7

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// pathRe matches addresses like:
//   PID-5            field
//   PID-5.1          component
//   PID-5.1.2        subcomponent
//   PID-3[2].1       repetition then component
//   MSH-9.2
var pathRe = regexp.MustCompile(`^([A-Z0-9]{3})-(\d+)(?:\[(\d+)\])?(?:\.(\d+))?(?:\.(\d+))?$`)

// Address is a parsed field path.
type Address struct {
	Segment    string
	Field      int // 1-based
	Repetition int // 1-based; 0 means unspecified (whole field / first rep)
	Component  int // 1-based; 0 means unspecified
	Subcompo   int // 1-based; 0 means unspecified
}

// ParsePath parses a textual path such as "PID-5.1" into an Address.
func ParsePath(path string) (Address, error) {
	m := pathRe.FindStringSubmatch(strings.TrimSpace(path))
	if m == nil {
		return Address{}, fmt.Errorf("invalid path %q (want e.g. PID-5, PID-5.1, PID-3[2].1)", path)
	}
	a := Address{Segment: m[1]}
	a.Field, _ = strconv.Atoi(m[2])
	a.Repetition = atoiOr0(m[3])
	a.Component = atoiOr0(m[4])
	a.Subcompo = atoiOr0(m[5])
	return a, nil
}

func atoiOr0(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// Segment returns the first segment with the given name.
func (m *Message) Segment(name string) (Segment, bool) {
	for _, s := range m.Segments {
		if s.Name == name {
			return s, true
		}
	}
	return Segment{}, false
}

// Get resolves a path to a raw value. When the path stops above the leaf
// (e.g. a field path on a composite value), the remaining structure is
// reassembled with the message delimiters.
func (m *Message) Get(path string) (string, error) {
	a, err := ParsePath(path)
	if err != nil {
		return "", err
	}
	seg, ok := m.Segment(a.Segment)
	if !ok {
		return "", fmt.Errorf("segment %s not present", a.Segment)
	}
	field, ok := seg.Field(a.Field)
	if !ok {
		return "", fmt.Errorf("%s-%d not present", a.Segment, a.Field)
	}

	if a.Repetition == 0 && a.Component == 0 {
		return field.Raw(m.Delims), nil
	}

	repIdx := 0
	if a.Repetition > 0 {
		repIdx = a.Repetition - 1
	}
	if repIdx < 0 || repIdx >= len(field.Repetitions) {
		return "", fmt.Errorf("%s repetition %d not present", path, a.Repetition)
	}
	rep := field.Repetitions[repIdx]

	if a.Component == 0 {
		return rep.Raw(m.Delims), nil
	}
	ci := a.Component - 1
	if ci < 0 || ci >= len(rep.Components) {
		return "", fmt.Errorf("%s component %d not present", path, a.Component)
	}
	comp := rep.Components[ci]

	if a.Subcompo == 0 {
		return comp.Raw(m.Delims), nil
	}
	si := a.Subcompo - 1
	if si < 0 || si >= len(comp.Subcomponents) {
		return "", fmt.Errorf("%s subcomponent %d not present", path, a.Subcompo)
	}
	return comp.Subcomponents[si], nil
}
