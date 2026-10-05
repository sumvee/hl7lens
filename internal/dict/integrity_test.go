package dict

import (
	"fmt"
	"testing"
)

// TestLaunchSegmentsPresent confirms the ADT + ORU launch segment set.
func TestLaunchSegmentsPresent(t *testing.T) {
	d := load(t)
	for _, name := range []string{"MSH", "EVN", "PID", "PV1", "NK1", "OBR", "OBX", "NTE"} {
		if _, ok := d.Segments[name]; !ok {
			t.Errorf("segment %s missing", name)
		}
	}
}

// TestMessagesPresent confirms both launch message grammars loaded.
func TestMessagesPresent(t *testing.T) {
	d := load(t)
	for _, s := range []string{"ADT_A01", "ORU_R01"} {
		if _, ok := d.Messages[s]; !ok {
			t.Errorf("message structure %s missing", s)
		}
	}
}

// TestMSHMessageType checks a representative composite binding.
func TestMSHMessageType(t *testing.T) {
	d := load(t)
	f, ok := d.Field("MSH", 9)
	if !ok || f.DataType != "MSG" {
		t.Fatalf("MSH-9 = %+v, want datatype MSG", f)
	}
	msg := d.DataTypes["MSG"]
	if msg.Components[1].Name != "Trigger Event" || msg.Components[1].Table != "0003" {
		t.Fatalf("MSG.2 = %+v, want Trigger Event / table 0003", msg.Components[1])
	}
}

// TestReferentialIntegrity is the key self-validation: every data type
// referenced by a segment field or by a component must be defined. This
// catches typos and missing definitions in the hand-curated dataset.
func TestReferentialIntegrity(t *testing.T) {
	d := load(t)

	check := func(where, dt string) {
		if dt == "" {
			t.Errorf("%s: empty datatype", where)
			return
		}
		if _, ok := d.DataTypes[dt]; !ok {
			t.Errorf("%s references undefined datatype %q", where, dt)
		}
	}

	for name, seg := range d.Segments {
		for _, f := range seg.Fields {
			check(fmt.Sprintf("%s-%d (%s)", name, f.Seq, f.Name), f.DataType)
		}
	}
	for name, dt := range d.DataTypes {
		for _, c := range dt.Components {
			check(fmt.Sprintf("%s.%d (%s)", name, c.Seq, c.Name), c.DataType)
		}
	}
}

// TestTableBindingsReference flags fields/components bound to a table id
// that is not loaded. This is informational, not fatal: many tables are
// not embedded yet, so it is logged (t.Log) rather than failed.
func TestTableBindingsReference(t *testing.T) {
	d := load(t)
	missing := map[string]bool{}
	for _, seg := range d.Segments {
		for _, f := range seg.Fields {
			if f.Table != "" {
				if _, ok := d.Tables[f.Table]; !ok {
					missing[f.Table] = true
				}
			}
		}
	}
	if len(missing) > 0 {
		t.Logf("tables bound but not yet embedded (expected, roadmap): %d distinct", len(missing))
	}
}
