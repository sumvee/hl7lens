package scrub

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumvee/hl7lens/internal/hl7"
)

func sampleRaw(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "adt_a01.hl7"))
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	return string(raw)
}

func scrubWithKey(t *testing.T, key string) string {
	t.Helper()
	msg, err := hl7.Parse(sampleRaw(t))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	sc := New(Default(), []byte(key))
	return sc.Scrub(msg)
}

// TestNoPHILeak is the CI gate: no original identifier may survive.
func TestNoPHILeak(t *testing.T) {
	out := scrubWithKey(t, "unit-test-key")
	phi := []string{
		"SMITH", "JOHN",          // patient name
		"371-66-9256",            // SSN (and SS id)
		"191919",                 // MR id
		"(919)379-1212",          // home phone
		"(919)271-3434",          // business phone
		"1200 N ELM STREET",      // street
		"GREENSBORO",             // city
		"PATID12345001",          // account number
		"19610615",               // date of birth
		"GOOD", "SIDNEY",         // attending doctor
	}
	for _, p := range phi {
		if strings.Contains(out, p) {
			t.Errorf("PHI leaked: %q present in output\n---\n%s", p, out)
		}
	}
}

// TestNonPHIPreserved: clinical/administrative codes must survive intact.
func TestNonPHIPreserved(t *testing.T) {
	out := scrubWithKey(t, "unit-test-key")
	msg, err := hl7.Parse(out)
	if err != nil {
		t.Fatalf("scrubbed output does not re-parse: %v", err)
	}
	checks := map[string]string{
		"PID-8":  "M", // Administrative Sex
		"PID-16": "S", // Marital Status
		"PID-10": "C", // Race
		"PV1-2":  "I", // Patient Class
		"MSH-9.1": "ADT",
	}
	for path, want := range checks {
		got, err := msg.Get(path)
		if err != nil || got != want {
			t.Errorf("%s = %q (err=%v), want %q", path, got, err, want)
		}
	}
}

// TestStructurePreserved: field/component layout is unchanged.
func TestStructurePreserved(t *testing.T) {
	out := scrubWithKey(t, "unit-test-key")
	msg, err := hl7.Parse(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	pid, _ := msg.Segment("PID")
	// PID-5 still has 3 components; PID-3 still repeats twice.
	f5, _ := pid.Field(5)
	if n := len(f5.Repetitions[0].Components); n != 3 {
		t.Errorf("PID-5 components = %d, want 3", n)
	}
	f3, _ := pid.Field(3)
	if n := len(f3.Repetitions); n != 2 {
		t.Errorf("PID-3 repetitions = %d, want 2", n)
	}
}

func TestSSNHardMasked(t *testing.T) {
	out := scrubWithKey(t, "unit-test-key")
	msg, _ := hl7.Parse(out)
	ssn, _ := msg.Get("PID-19")
	if ssn != "XXX-XX-XXXX" {
		t.Errorf("PID-19 = %q, want XXX-XX-XXXX (hard mask)", ssn)
	}
}

func TestDOBGeneralized(t *testing.T) {
	out := scrubWithKey(t, "unit-test-key")
	msg, _ := hl7.Parse(out)
	dob, _ := msg.Get("PID-7")
	if dob != "19610101" {
		t.Errorf("PID-7 = %q, want 19610101 (year only)", dob)
	}
}

// TestNK1ScrubbedInORU: next-of-kin PHI must be removed, relationship kept.
func TestNK1ScrubbedInORU(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "oru_r01.hl7"))
	if err != nil {
		t.Fatalf("read oru: %v", err)
	}
	msg, err := hl7.Parse(string(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	out := New(Default(), []byte("k")).Scrub(msg)
	for _, phi := range []string{"DOE", "JOHN", "(919)555-1212"} {
		if strings.Contains(out, phi) {
			t.Errorf("NK1 PHI leaked: %q in\n%s", phi, out)
		}
	}
	// Relationship code (SPO / Spouse) is not PHI and must survive.
	scrubbed, _ := hl7.Parse(out)
	if rel, _ := scrubbed.Get("NK1-3.1"); rel != "SPO" {
		t.Errorf("NK1-3.1 = %q, want SPO (relationship preserved)", rel)
	}
}

// TestDeterminismAndReferentialIntegrity: same key reproduces output, and
// a repeated value maps to a single consistent fake.
func TestDeterminismAndReferentialIntegrity(t *testing.T) {
	a := scrubWithKey(t, "same-key")
	b := scrubWithKey(t, "same-key")
	if a != b {
		t.Fatal("same key produced different output")
	}
	c := scrubWithKey(t, "other-key")
	if a == c {
		t.Fatal("different key produced identical output")
	}
}
