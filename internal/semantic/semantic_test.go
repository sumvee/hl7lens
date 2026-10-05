package semantic

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

func buildSample(t *testing.T) string {
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
		t.Fatalf("load dict: %v", err)
	}
	return Build(msg, d).Text()
}

func TestExplainDecodes(t *testing.T) {
	out := buildSample(t)
	wants := []string{
		"MSH  Message Header",
		"PID  Patient Identification",
		`^~\&`,                     // MSH-2 literal
		"Male (table 0001)",        // PID-8 primitive coded field
		"Single (table 0002)",      // PID-16 coded composite, .1 inherits table
		"Family Name",              // XPN component naming
		"PID-3[1]",                 // repetition addressing
		"PID-3[2]",                 //
		"Trigger Event",            // MSG composite component
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("explain output missing %q\n---\n%s", w, out)
		}
	}
}

func TestExplainSkipsEmpty(t *testing.T) {
	out := buildSample(t)
	// PID-2 is empty in the sample and should not appear as its own line.
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "PID-2 ") {
			t.Errorf("empty field PID-2 should be skipped, got: %q", line)
		}
	}
}
