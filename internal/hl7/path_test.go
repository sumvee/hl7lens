package hl7

import "testing"

func TestGet(t *testing.T) {
	msg := loadSample(t)
	cases := map[string]string{
		"PID-5.1":    "SMITH",
		"PID-5.2":    "JOHN",
		"PID-8":      "M",
		"MSH-9.1":    "ADT",
		"MSH-9.2":    "A01",
		"PID-3[2].1": "371-66-9256",
		"PID-3[1].5": "MR",
		"PV1-2":      "I",
	}
	for path, want := range cases {
		got, err := msg.Get(path)
		if err != nil {
			t.Errorf("Get(%q): %v", path, err)
			continue
		}
		if got != want {
			t.Errorf("Get(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestGetMSH2Literal(t *testing.T) {
	msg := loadSample(t)
	got, err := msg.Get("MSH-2")
	if err != nil {
		t.Fatalf("Get MSH-2: %v", err)
	}
	if got != `^~\&` {
		t.Fatalf("MSH-2 = %q, want %q (encoding chars must stay literal)", got, `^~\&`)
	}
}

func TestGetErrors(t *testing.T) {
	msg := loadSample(t)
	for _, bad := range []string{"nope", "PID5", "ZZZ-1", "PID-99", "PID-3[9].1"} {
		if _, err := msg.Get(bad); err == nil {
			t.Errorf("Get(%q) expected error", bad)
		}
	}
}
