package dict

import "testing"

func load(t *testing.T) *Dictionary {
	t.Helper()
	d, err := Load("2.5.1")
	if err != nil {
		t.Fatalf("load 2.5.1: %v", err)
	}
	return d
}

func TestLoadCounts(t *testing.T) {
	d := load(t)
	if len(d.Segments) == 0 || len(d.DataTypes) == 0 || len(d.Tables) == 0 {
		t.Fatalf("empty dataset: seg=%d dt=%d tbl=%d",
			len(d.Segments), len(d.DataTypes), len(d.Tables))
	}
}

func TestLoadMissingVersion(t *testing.T) {
	if _, err := Load("9.9.9"); err == nil {
		t.Fatal("expected error for unknown version")
	}
}

func TestFieldResolution(t *testing.T) {
	d := load(t)

	f, ok := d.Field("PID", 5)
	if !ok {
		t.Fatal("PID-5 not found")
	}
	if f.Name != "Patient Name" || f.DataType != "XPN" || f.Optionality != "R" || !f.Repeats {
		t.Fatalf("PID-5 = %+v", f)
	}

	f8, ok := d.Field("PID", 8)
	if !ok || f8.Table != "0001" {
		t.Fatalf("PID-8 table = %q (ok=%v), want 0001", f8.Table, ok)
	}

	if _, ok := d.Field("PID", 99); ok {
		t.Error("PID-99 should not resolve")
	}
	if _, ok := d.Field("ZZZ", 1); ok {
		t.Error("unknown segment should not resolve")
	}
}

func TestTableDecode(t *testing.T) {
	d := load(t)

	if v, ok := d.Decode("0001", "M"); !ok || v != "Male" {
		t.Errorf("0001/M = %q (ok=%v), want Male", v, ok)
	}
	if v, ok := d.Decode("0002", "M"); !ok || v != "Married" {
		t.Errorf("0002/M = %q (ok=%v), want Married", v, ok)
	}
	if _, ok := d.Decode("0001", "ZZ"); ok {
		t.Error("unknown code should not decode")
	}
}

func TestDataTypeComponents(t *testing.T) {
	d := load(t)

	xpn, ok := d.DataTypes["XPN"]
	if !ok || xpn.Primitive {
		t.Fatalf("XPN not a composite: %+v", xpn)
	}
	if xpn.Components[0].Name != "Family Name" || xpn.Components[0].DataType != "FN" {
		t.Fatalf("XPN.1 = %+v, want Family Name/FN", xpn.Components[0])
	}

	st, ok := d.DataTypes["ST"]
	if !ok || !st.Primitive {
		t.Fatalf("ST should be primitive: %+v", st)
	}
}
