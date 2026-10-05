package semantic

import (
	"fmt"
	"strings"
)

const (
	fieldPathWidth = 10
	nameWidth      = 36
	typeWidth      = 5
	compPathWidth  = 6
	compNameWidth  = 32
)

// Text renders the annotation tree as aligned, plain (non-ANSI) text
// suitable for `explain` output, piping, and golden tests.
func (mv MessageView) Text() string {
	var b strings.Builder
	for si, seg := range mv.Segments {
		if si > 0 {
			b.WriteByte('\n')
		}
		writeSegmentHeader(&b, seg)
		for _, f := range seg.Fields {
			writeField(&b, f)
			for _, c := range f.Components {
				writeComponent(&b, c)
			}
		}
	}
	return b.String()
}

func writeSegmentHeader(b *strings.Builder, seg SegmentView) {
	if seg.Known {
		fmt.Fprintf(b, "%s  %s\n", seg.Name, seg.Description)
	} else {
		fmt.Fprintf(b, "%s  (segment not in dictionary)\n", seg.Name)
	}
}

func writeField(b *strings.Builder, f FieldView) {
	name := f.Name
	if name == "" {
		name = "(unknown field)"
	}
	fmt.Fprintf(b, "  %-*s %-*s %-*s %s",
		fieldPathWidth, f.Path,
		nameWidth, truncate(name, nameWidth),
		typeWidth, f.DataType,
		f.Raw,
	)
	if f.Decoded != "" {
		fmt.Fprintf(b, "  => %s%s", f.Decoded, tableSuffix(f.Table))
	}
	b.WriteByte('\n')
}

func writeComponent(b *strings.Builder, c ComponentView) {
	name := c.Name
	if name == "" {
		name = "(unknown component)"
	}
	fmt.Fprintf(b, "    %-*s %-*s %s",
		compPathWidth, c.Path,
		compNameWidth, truncate(name, compNameWidth),
		c.Raw,
	)
	if c.Decoded != "" {
		fmt.Fprintf(b, "  => %s%s", c.Decoded, tableSuffix(c.Table))
	}
	b.WriteByte('\n')
}

func tableSuffix(table string) string {
	if table == "" {
		return ""
	}
	return fmt.Sprintf(" (table %s)", table)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}
