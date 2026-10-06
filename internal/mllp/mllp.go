// Package mllp implements the HL7 Minimal Lower Layer Protocol: the TCP
// framing used to carry v2.x messages between systems, plus ACK building.
//
// On the wire each message is wrapped as:
//
//	<VT> message-bytes <FS><CR>
//	0x0B    ...        0x1C 0x0D
package mllp

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/sumvee/hl7lens/internal/hl7"
)

// Framing bytes.
const (
	VT = 0x0b // start block
	FS = 0x1c // end block
	CR = 0x0d // carriage return following the end block
)

// Frame wraps a message payload in MLLP framing.
func Frame(payload []byte) []byte {
	out := make([]byte, 0, len(payload)+3)
	out = append(out, VT)
	out = append(out, payload...)
	out = append(out, FS, CR)
	return out
}

// Reader reads MLLP-framed messages from a stream, tolerating junk before
// a start block and an optional CR after the end block.
type Reader struct{ r *bufio.Reader }

// NewReader wraps r.
func NewReader(r io.Reader) *Reader { return &Reader{r: bufio.NewReader(r)} }

// ReadMessage returns the next unframed message, or io.EOF when the stream
// ends with no further message.
func (m *Reader) ReadMessage() ([]byte, error) {
	// Discard everything up to and including the start block.
	if _, err := m.r.ReadBytes(VT); err != nil {
		return nil, err
	}
	// Read the payload up to the end block.
	payload, err := m.r.ReadBytes(FS)
	if err != nil {
		return nil, err
	}
	payload = payload[:len(payload)-1] // strip FS
	// Consume the trailing CR if present; otherwise leave the byte.
	if b, err := m.r.ReadByte(); err == nil && b != CR {
		_ = m.r.UnreadByte()
	}
	return payload, nil
}

// BuildACK constructs an acknowledgement for a received message with the
// given code (AA accept, AE error, AR reject). Sending and receiving
// application/facility are swapped, and the original control id is echoed
// in MSA-2.
func BuildACK(orig *hl7.Message, code string) string {
	get := func(path string) string { v, _ := orig.Get(path); return v }

	sendApp, sendFac := get("MSH-3"), get("MSH-4")
	recvApp, recvFac := get("MSH-5"), get("MSH-6")
	trigger := get("MSH-9.2")
	control := get("MSH-10")
	version := get("MSH-12")
	if version == "" {
		version = "2.5.1"
	}
	ts := time.Now().Format("20060102150405")

	msgType := "ACK"
	if trigger != "" {
		msgType = "ACK^" + trigger + "^ACK"
	}
	msh := fmt.Sprintf("MSH|^~\\&|%s|%s|%s|%s|%s||%s|%s|P|%s",
		recvApp, recvFac, sendApp, sendFac, ts, msgType, control, version)
	msa := fmt.Sprintf("MSA|%s|%s", code, control)
	return msh + "\r" + msa
}
