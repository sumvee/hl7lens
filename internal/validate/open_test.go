package validate

import (
	"testing"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

// A bogus code in a CLOSED table (0003) must warn; a bogus code in an OPEN
// (user-defined) table (0001) must not, since its entries are suggestions.
func TestOpenVsClosedTableMembership(t *testing.T) {
	d, err := dict.Load("2.5.1")
	if err != nil {
		t.Fatal(err)
	}
	raw := "MSH|^~\\&|A|B|C|D|200101061015||ADT^A01^ADT_A01|1|P|2.5.1\r" +
		"EVN|ZZ|200101061015\r" + // EVN-1 bound to closed 0003, bogus value
		"PID|1||5^^^X^MR||DOE^JOHN||19700101|Q\r" + // PID-8 bound to open 0001, bogus value
		"PV1|1|I\r"
	msg, err := hl7.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	r := Validate(msg, d)

	if !hasFinding(r, Warning, "not in table 0003") {
		t.Errorf("closed table 0003 bogus code should warn; findings=%+v", r.Findings)
	}
	if hasFinding(r, Warning, "not in table 0001") {
		t.Errorf("open table 0001 bogus code must NOT warn; findings=%+v", r.Findings)
	}
}
