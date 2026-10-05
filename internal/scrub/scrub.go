package scrub

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"unicode"

	"github.com/sumvee/hl7lens/internal/hl7"
)

// Scrubber applies a Profile to a message using a pseudonymization key.
type Scrubber struct {
	profile Profile
	key     []byte
}

// New returns a scrubber with an explicit key. A fixed key makes
// pseudonymization reproducible across runs and files; callers that want
// irreversibility should use NewRandom.
func New(p Profile, key []byte) *Scrubber {
	return &Scrubber{profile: p, key: key}
}

// NewRandom returns a scrubber with a fresh random key, so the output is
// not reversible and differs run to run.
func NewRandom(p Profile) (*Scrubber, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return &Scrubber{profile: p, key: key}, nil
}

// Scrub returns a de-identified copy of the message. Segments that are not
// PHI-bearing pass through unchanged; structure is preserved throughout.
func (s *Scrubber) Scrub(msg *hl7.Message) string {
	lines := make([]string, 0, len(msg.Segments))
	for _, seg := range msg.Segments {
		if _, isPHI := s.profile.Segments[seg.Name]; !isPHI {
			lines = append(lines, seg.Raw(msg.Delims))
			continue
		}
		lines = append(lines, s.transformSegment(seg, msg.Delims).Raw(msg.Delims))
	}
	return strings.Join(lines, "\n")
}

func (s *Scrubber) transformSegment(seg hl7.Segment, d hl7.Delimiters) hl7.Segment {
	out := hl7.Segment{Name: seg.Name, Fields: make([]hl7.Field, len(seg.Fields))}
	for i, f := range seg.Fields {
		seq := i + 1
		t, _ := s.profile.Policy(seg.Name, seq)
		out.Fields[i] = s.applyField(f, t)
	}
	return out
}

func (s *Scrubber) applyField(f hl7.Field, t Treatment) hl7.Field {
	if t == Pass {
		return f
	}
	out := hl7.Field{Repetitions: make([]hl7.Repetition, len(f.Repetitions))}
	for ri, rep := range f.Repetitions {
		nr := hl7.Repetition{Components: make([]hl7.Component, len(rep.Components))}
		for ci, comp := range rep.Components {
			ns := make([]string, len(comp.Subcomponents))
			for si, sub := range comp.Subcomponents {
				ns[si] = s.transform(sub, t)
			}
			nr.Components[ci] = hl7.Component{Subcomponents: ns}
		}
		out.Repetitions[ri] = nr
	}
	return out
}

func (s *Scrubber) transform(value string, t Treatment) string {
	if value == "" {
		return ""
	}
	switch t {
	case Mask:
		return maskHard(value)
	case GeneralizeDate:
		return generalizeDate(value)
	case Pseudonymize:
		return s.pseudo(value)
	default:
		return value
	}
}

// pseudo replaces a value with a deterministic fake, token by token, so
// that referential integrity is preserved within a run (the same input
// always maps to the same output).
func (s *Scrubber) pseudo(value string) string {
	fields := splitKeepSpaces(value)
	for i, tok := range fields {
		if tok == "" || isSpace(tok) {
			continue
		}
		h := s.mac(tok)
		if isNumericish(tok) {
			fields[i] = digitRemap(tok, h)
		} else {
			fields[i] = wordFor(h)
		}
	}
	return strings.Join(fields, "")
}

func (s *Scrubber) mac(tok string) []byte {
	m := hmac.New(sha256.New, s.key)
	m.Write([]byte(tok))
	return m.Sum(nil)
}

// ---- transforms ----

func maskHard(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return 'X'
		}
		return r
	}, s)
}

func generalizeDate(s string) string {
	if len(s) >= 4 && allDigits(s[:4]) {
		return s[:4] + "0101"
	}
	return maskHard(s)
}

func digitRemap(s string, h []byte) string {
	i := 0
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			d := h[i%len(h)] % 10
			i++
			return rune('0' + d)
		}
		return r
	}, s)
}

var fakeWords = []string{
	"KESTREL", "MERIDIAN", "HARLOW", "VANCE", "OSWALD", "BRECKEN", "MARLOWE",
	"FENWICK", "ALDRIDGE", "CORVO", "DELACROIX", "ENNIS", "FABLES", "GRANGER",
	"HOLLIS", "IVERSON", "JARRAH", "KELVIN", "LOWELL", "MERCER", "NAVARRE",
	"ORTEGA", "PENROSE", "QUINCY", "RADLEY", "SABLE", "THORNE", "UPTON",
	"VESPER", "WREN", "XANTHE", "YARROW", "ZEPHYR", " ASHBY", "BLYTHE",
	"CADE", "DRAVEN", "ELORA", "FINCH", "GALE", "HEATH", "INGA", "JUNIPER",
	"KANE", "LARK", "MABRY", "NOLAN", "OAKES",
}

func wordFor(h []byte) string {
	idx := binary.BigEndian.Uint32(h[:4]) % uint32(len(fakeWords))
	return strings.TrimSpace(fakeWords[idx])
}

// ---- helpers ----

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// isNumericish reports whether a token has digits and no letters (so it is
// an identifier/phone/number rather than a name/word).
func isNumericish(tok string) bool {
	hasDigit := false
	for _, r := range tok {
		if unicode.IsLetter(r) {
			return false
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasDigit
}

func isSpace(tok string) bool {
	for _, r := range tok {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return len(tok) > 0
}

// splitKeepSpaces splits on runs of spaces but keeps the space runs as
// their own elements, so Join reproduces the original spacing.
func splitKeepSpaces(s string) []string {
	var out []string
	var cur strings.Builder
	inSpace := false
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		sp := unicode.IsSpace(r)
		if sp != inSpace {
			flush()
			inSpace = sp
		}
		cur.WriteRune(r)
	}
	flush()
	return out
}
