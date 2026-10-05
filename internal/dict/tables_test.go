package dict

import "testing"

func TestADTTablesDecode(t *testing.T) {
	d := load(t)
	cases := []struct{ tbl, code, want string }{
		{"0003", "A01", "ADT/ACK - Admit/visit notification"},
		{"0003", "R01", "ORU/ACK - Unsolicited transmission of an observation message"},
		{"0004", "I", "Inpatient"},
		{"0007", "E", "Emergency"},
		{"0009", "A0", "No functional limitations"},
		{"0063", "SPO", "Spouse"},
		{"0069", "SUR", "Surgical Service"},
		{"0078", "N", "Normal"},
		{"0078", "LL", "Critical low"},
	}
	for _, c := range cases {
		if v, ok := d.Decode(c.tbl, c.code); !ok || v != c.want {
			t.Errorf("Decode(%s,%s) = %q (ok=%v), want %q", c.tbl, c.code, v, ok, c.want)
		}
	}
}

func TestTableOpenFlags(t *testing.T) {
	d := load(t)
	// HL7-defined (closed) tables enforce membership.
	for _, id := range []string{"0003", "0078", "0085", "0136"} {
		if d.Tables[id].Open {
			t.Errorf("table %s should be closed", id)
		}
	}
	// User-defined (open) tables are decode-only.
	for _, id := range []string{"0001", "0002", "0004", "0007", "0009", "0063", "0069"} {
		if !d.Tables[id].Open {
			t.Errorf("table %s should be open (user-defined)", id)
		}
	}
}
