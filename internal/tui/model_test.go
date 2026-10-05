package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
	"github.com/sumvee/hl7lens/internal/semantic"
)

func sampleModel(t *testing.T) Model {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "adt_a01.hl7"))
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	msg, err := hl7.Parse(string(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	d, err := dict.Load("2.5.1")
	if err != nil {
		t.Fatalf("dict: %v", err)
	}
	m := New(semantic.Build(msg, d), "test")
	m.height, m.width = 60, 120
	return m
}

func TestModelBuildsRows(t *testing.T) {
	m := sampleModel(t)
	if len(m.rows) == 0 {
		t.Fatal("no rows built")
	}
	out := m.View()
	for _, want := range []string{"MSH", "Patient Name", "Male"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestCollapseTogglesVisibility(t *testing.T) {
	m := sampleModel(t)
	before := len(m.visible())

	// Cursor starts on the first segment header (MSH). Collapse it.
	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(Model)
	afterCollapse := len(m.visible())
	if afterCollapse >= before {
		t.Fatalf("collapse did not hide rows: before=%d after=%d", before, afterCollapse)
	}

	// Expand again.
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeySpace})
	m = next.(Model)
	if got := len(m.visible()); got != before {
		t.Fatalf("expand did not restore rows: got=%d want=%d", got, before)
	}
}

func TestNavigationClamps(t *testing.T) {
	m := sampleModel(t)
	// Up at the top stays at 0.
	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(Model)
	if m.cursor != 0 {
		t.Errorf("cursor moved above 0: %d", m.cursor)
	}
	// End jumps to last visible row.
	next, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = next.(Model)
	if m.cursor != len(m.visible())-1 {
		t.Errorf("G did not jump to bottom: cursor=%d last=%d", m.cursor, len(m.visible())-1)
	}
}

func TestQuitKey(t *testing.T) {
	m := sampleModel(t)
	_, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return a quit command")
	}
}
