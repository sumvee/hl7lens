package mllp

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/sumvee/hl7lens/internal/hl7"
)

const sample = "MSH|^~\\&|SEND|SFAC|RECV|RFAC|20200101||ADT^A01^ADT_A01|CTRL99|P|2.5.1\rPID|1||5^^^X^MR||DOE^JOHN"

func TestFrameRoundTrip(t *testing.T) {
	framed := Frame([]byte(sample))
	if framed[0] != VT || framed[len(framed)-2] != FS || framed[len(framed)-1] != CR {
		t.Fatalf("framing bytes wrong: % x ... % x", framed[0], framed[len(framed)-2:])
	}
	got, err := NewReader(bytes.NewReader(framed)).ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != sample {
		t.Fatalf("round-trip mismatch:\n got %q\nwant %q", got, sample)
	}
}

func TestReaderSkipsJunkAndReadsTwo(t *testing.T) {
	// leading junk, two framed messages back to back
	var buf bytes.Buffer
	buf.WriteString("garbage")
	buf.Write(Frame([]byte("MSH|^~\\&|A|B|C|D|1||ADT^A01|one|P|2.5.1")))
	buf.Write(Frame([]byte("MSH|^~\\&|A|B|C|D|1||ADT^A01|two|P|2.5.1")))
	r := NewReader(&buf)
	for _, want := range []string{"one", "two"} {
		m, err := r.ReadMessage()
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if msg, _ := hl7.Parse(string(m)); func() bool { v, _ := msg.Get("MSH-10"); return v == want }() == false {
			t.Errorf("control id != %s", want)
		}
	}
}

func TestBuildACK(t *testing.T) {
	msg, _ := hl7.Parse(sample)
	ack := BuildACK(msg, "AA")
	am, err := hl7.Parse(ack)
	if err != nil {
		t.Fatalf("ack does not parse: %v", err)
	}
	// MSA-1 code, MSA-2 echoes the original control id.
	if c, _ := am.Get("MSA-1"); c != "AA" {
		t.Errorf("MSA-1 = %q, want AA", c)
	}
	if c, _ := am.Get("MSA-2"); c != "CTRL99" {
		t.Errorf("MSA-2 = %q, want CTRL99 (echoed)", c)
	}
	// Sending/receiving swapped: ACK MSH-3 is the original MSH-5.
	if a, _ := am.Get("MSH-3"); a != "RECV" {
		t.Errorf("ACK MSH-3 = %q, want RECV (swapped)", a)
	}
}

// TestLoopback drives the real network path: a listener reads a framed
// message and returns an ACK; a client sends and reads it back.
func TestLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		m, err := NewReader(conn).ReadMessage()
		if err != nil {
			return
		}
		msg, _ := hl7.Parse(string(m))
		_, _ = conn.Write(Frame([]byte(BuildACK(msg, "AA"))))
	}()

	conn, err := net.DialTimeout("tcp", ln.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

	if _, err := conn.Write(Frame([]byte(sample))); err != nil {
		t.Fatal(err)
	}
	ackBytes, err := NewReader(conn).ReadMessage()
	if err != nil {
		t.Fatalf("read ack: %v", err)
	}
	am, _ := hl7.Parse(string(ackBytes))
	if c, _ := am.Get("MSA-2"); c != "CTRL99" {
		t.Fatalf("loopback ACK MSA-2 = %q, want CTRL99", c)
	}
}
