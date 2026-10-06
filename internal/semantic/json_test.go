package semantic

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

func TestJSONMarshal(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "adt_a01.hl7"))
	if err != nil {
		t.Fatal(err)
	}
	msg, err := hl7.Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	d, err := dict.Load("2.5.1")
	if err != nil {
		t.Fatal(err)
	}
	view := Build(msg, d)

	b, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Valid JSON that round-trips.
	var back MessageView
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("round-trip: %v", err)
	}
	if string(back.Version) != "2.5.1" || len(back.Segments) == 0 {
		t.Fatalf("round-trip lost data: %+v", back)
	}

	s := string(b)
	for _, want := range []string{
		`"version":"2.5.1"`,
		`"path":"PID-8"`,
		`"decoded":"Male"`,
		`"table":"0001"`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("JSON missing %q", want)
		}
	}
	// omitempty: an empty component list must not appear.
	if strings.Contains(s, `"components":[]`) {
		t.Error("empty components should be omitted")
	}
}
