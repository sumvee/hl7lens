// Package tui renders a parsed, annotated HL7 message as an interactive
// tree (Bubble Tea). It consumes the same semantic.MessageView that the
// `explain` text output uses, so the two never drift apart.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sumvee/hl7lens/internal/semantic"
)

type rowKind int

const (
	rowSeg rowKind = iota
	rowField
	rowComp
)

type row struct {
	kind rowKind
	seg  int    // index of the owning segment
	text string // styled content (without cursor treatment)
}

// Model is the Bubble Tea model for `view`.
type Model struct {
	title     string
	rows      []row
	collapsed map[int]bool
	cursor    int // index into the visible rows
	offset    int // scroll offset into the visible rows
	height    int
	width     int
}

var (
	segStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	pathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	nameStyle    = lipgloss.NewStyle()
	decodeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cursorStyle  = lipgloss.NewStyle().Reverse(true)
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	headingStyle = lipgloss.NewStyle().Bold(true)
)

// New builds a model from an annotated message view.
func New(mv semantic.MessageView, title string) Model {
	m := Model{
		title:     title,
		collapsed: map[int]bool{},
		height:    24,
		width:     80,
	}
	for si, seg := range mv.Segments {
		head := segStyle.Render(seg.Name)
		if seg.Description != "" {
			head += "  " + seg.Description
		}
		m.rows = append(m.rows, row{kind: rowSeg, seg: si, text: head})
		for _, f := range seg.Fields {
			m.rows = append(m.rows, row{kind: rowField, seg: si, text: renderField(f)})
			for _, c := range f.Components {
				m.rows = append(m.rows, row{kind: rowComp, seg: si, text: renderComp(c)})
			}
		}
	}
	return m
}

func renderField(f semantic.FieldView) string {
	name := f.Name
	if name == "" {
		name = "(unknown)"
	}
	s := fmt.Sprintf("  %s %s %s %s",
		pathStyle.Render(pad(f.Path, 10)),
		nameStyle.Render(pad(name, 34)),
		pathStyle.Render(pad(f.DataType, 4)),
		f.Raw,
	)
	if f.Decoded != "" {
		s += "  " + decodeStyle.Render("=> "+f.Decoded+tableSuffix(f.Table))
	}
	return s
}

func renderComp(c semantic.ComponentView) string {
	name := c.Name
	if name == "" {
		name = "(unknown)"
	}
	s := fmt.Sprintf("      %s %s %s",
		pathStyle.Render(pad(c.Path, 6)),
		pathStyle.Render(pad(name, 30)),
		c.Raw,
	)
	if c.Decoded != "" {
		s += "  " + decodeStyle.Render("=> "+c.Decoded+tableSuffix(c.Table))
	}
	return s
}

func tableSuffix(t string) string {
	if t == "" {
		return ""
	}
	return " (table " + t + ")"
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// visible returns the indices of rows currently shown (respecting
// collapsed segments).
func (m Model) visible() []int {
	var out []int
	for i, r := range m.rows {
		if r.kind != rowSeg && m.collapsed[r.seg] {
			continue
		}
		out = append(out, i)
	}
	return out
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vis := m.visible()
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "down", "j":
		if m.cursor < len(vis)-1 {
			m.cursor++
		}
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		m.cursor = len(vis) - 1
	case " ", "space", "enter":
		if len(vis) > 0 {
			r := m.rows[vis[m.cursor]]
			m.collapsed[r.seg] = !m.collapsed[r.seg]
			// keep the cursor on the segment header after toggling
			m.cursor = segHeaderVisibleIndex(m, r.seg)
		}
	}
	m.clampScroll()
	return m, nil
}

func segHeaderVisibleIndex(m Model, seg int) int {
	for i, idx := range m.visible() {
		if m.rows[idx].kind == rowSeg && m.rows[idx].seg == seg {
			return i
		}
	}
	return 0
}

func (m *Model) clampScroll() {
	body := m.bodyHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+body {
		m.offset = m.cursor - body + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) bodyHeight() int {
	h := m.height - 3 // title + blank + footer
	if h < 1 {
		return 1
	}
	return h
}

// View implements tea.Model.
func (m Model) View() string {
	vis := m.visible()
	var b strings.Builder
	b.WriteString(headingStyle.Render(m.title))
	b.WriteByte('\n')

	body := m.bodyHeight()
	end := m.offset + body
	if end > len(vis) {
		end = len(vis)
	}
	for i := m.offset; i < end; i++ {
		line := m.rows[vis[i]].text
		if i == m.cursor {
			line = cursorStyle.Render(stripToWidth(line, m.width))
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	b.WriteString(hintStyle.Render("↑/↓ move · space collapse segment · g/G top/bottom · q quit"))
	return b.String()
}

// stripToWidth trims a styled line's visible length so the cursor
// highlight does not wrap. It is a best-effort trim on the raw string.
func stripToWidth(s string, w int) string {
	if w <= 0 || len(s) <= w {
		return s
	}
	return s[:w]
}
