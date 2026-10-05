package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

func report(t *testing.T, file string) Report {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	msg, err := hl7.Parse(string(raw))
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	d, err := dict.Load("2.5.1")
	if err != nil {
		t.Fatalf("load dict: %v", err)
	}
	return Validate(msg, d)
}

func hasFinding(r Report, lvl Level, substr string) bool {
	for _, f := range r.Findings {
		if f.Level == lvl && strings.Contains(f.Message, substr) {
			return true
		}
	}
	return false
}

func TestCleanADTPasses(t *testing.T) {
	r := report(t, "adt_a01.hl7")
	if r.HasErrors() {
		t.Fatalf("clean ADT reported errors: %+v", r.Findings)
	}
}

func TestCleanORUPasses(t *testing.T) {
	r := report(t, "oru_r01.hl7")
	if len(r.Findings) != 0 {
		t.Fatalf("clean ORU should have no findings, got: %+v", r.Findings)
	}
}

func TestBadADTFindings(t *testing.T) {
	r := report(t, "adt_bad.hl7")
	if !r.HasErrors() {
		t.Fatal("bad ADT should have errors")
	}
	// T2: EVN is a required segment and is missing.
	if !hasFinding(r, Error, "required segment EVN") {
		t.Error("missing EVN not flagged")
	}
	// T2: PV1 must not repeat (two present).
	if !hasFinding(r, Error, "PV1 must not repeat") {
		t.Error("duplicate PV1 not flagged")
	}
	// T1: PID-3 is required but empty.
	if !hasFinding(r, Error, "Patient Identifier List is required") {
		t.Error("empty required PID-3 not flagged")
	}
	// T1: PID-1 Set ID is an SI but holds "X".
	if !hasFinding(r, Warning, "expects a number") {
		t.Error("non-numeric SI not flagged")
	}
}
