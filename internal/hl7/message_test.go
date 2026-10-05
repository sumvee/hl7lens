package hl7

import (
	"os"
	"path/filepath"
	"testing"
)

func loadSample(t *testing.T) *Message {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "adt_a01.hl7"))
	if err != nil {
		t.Fatalf("read sample: %v", err)
	}
	msg, err := Parse(string(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return msg
}

func TestParseDelimiters(t *testing.T) {
	msg := loadSample(t)
	d := msg.Delims
	if d.Field != '|' || d.Component != '^' || d.Repetition != '~' || d.Escape != '\\' || d.Subcomponent != '&' {
		t.Fatalf("unexpected delimiters: %+v", d)
	}
}

func TestParseSegmentCount(t *testing.T) {
	msg := loadSample(t)
	if got, want := len(msg.Segments), 4; got != want {
		t.Fatalf("segments = %d, want %d", got, want)
	}
	want := []string{"MSH", "EVN", "PID", "PV1"}
	for i, name := range want {
		if msg.Segments[i].Name != name {
			t.Errorf("segment[%d] = %q, want %q", i, msg.Segments[i].Name, name)
		}
	}
}

// MSH-9 is the message type; index 8 since Fields[n] addresses MSH-(n+1).
func TestParseMSHMessageType(t *testing.T) {
	msh := msg(t, "MSH")
	got := msh.Fields[8].Repetitions[0].Components
	if len(got) < 2 || got[0].Subcomponents[0] != "ADT" || got[1].Subcomponents[0] != "A01" {
		t.Fatalf("MSH-9 components = %+v, want ADT / A01", got)
	}
}

// PID-5 (patient name, XPN) sits at Fields index 4.
func TestParsePatientName(t *testing.T) {
	pid := msg(t, "PID")
	comps := pid.Fields[4].Repetitions[0].Components
	if comps[0].Subcomponents[0] != "SMITH" || comps[1].Subcomponents[0] != "JOHN" {
		t.Fatalf("PID-5 = %+v, want SMITH / JOHN", comps)
	}
}

// PID-3 repeats (MR id ~ SS id); check both repetitions survive.
func TestParseRepetition(t *testing.T) {
	pid := msg(t, "PID")
	reps := pid.Fields[2].Repetitions
	if len(reps) != 2 {
		t.Fatalf("PID-3 repetitions = %d, want 2", len(reps))
	}
	if reps[0].Components[0].Subcomponents[0] != "191919" {
		t.Errorf("PID-3[1].1 = %q, want 191919", reps[0].Components[0].Subcomponents[0])
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := Parse(""); err != ErrEmpty {
		t.Errorf("empty: got %v, want ErrEmpty", err)
	}
	if _, err := Parse("PID|1"); err != ErrNoMSH {
		t.Errorf("no MSH: got %v, want ErrNoMSH", err)
	}
}

func msg(t *testing.T, name string) Segment {
	t.Helper()
	m := loadSample(t)
	for _, s := range m.Segments {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("segment %q not found", name)
	return Segment{}
}
