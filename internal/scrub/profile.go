// Package scrub de-identifies HL7 v2.x messages: it removes PHI while
// preserving message structure so the result is safe to share (tickets,
// bug reports) and still usable as test data.
//
// Design (locked):
//   - PHI is identified by a dictionary-driven profile (per segment, per
//     field), not by guessing from values.
//   - Replacement is hybrid: identifiers/names/addresses/phones are
//     pseudonymized deterministically (same input -> same fake within a
//     run, so referential integrity survives); the highest-risk
//     identifiers (SSN, account, licence) are hard-masked irreversibly.
//   - Safe by default: in a segment marked strict (predominantly PHI, e.g.
//     PID), any field not explicitly passed through is masked. A field we
//     have never seen can never leak.
package scrub

// Treatment is what scrub does to a field's values.
type Treatment int

const (
	// Pass leaves the value unchanged (explicitly judged non-identifying).
	Pass Treatment = iota
	// Mask replaces alphanumerics irreversibly, preserving layout.
	Mask
	// Pseudonymize replaces with a deterministic fake, keyed per run.
	Pseudonymize
	// GeneralizeDate reduces a date to its year (YYYY0101).
	GeneralizeDate
)

// SegmentProfile is the treatment policy for one segment type.
type SegmentProfile struct {
	// Default applies to any field without an explicit entry. For a strict
	// (predominantly-PHI) segment this is Mask; for a segment with only a
	// few PHI fields it is Pass.
	Default Treatment
	Fields  map[int]Treatment // 1-based field seq -> treatment
}

// Profile is the full de-identification policy. A segment not present is
// passed through unchanged (not a PHI-bearing segment).
type Profile struct {
	Segments map[string]SegmentProfile
}

// Policy returns the treatment for <segment>-<seq>, and whether the
// segment is a PHI-bearing segment at all.
func (p Profile) Policy(segment string, seq int) (Treatment, bool) {
	sp, ok := p.Segments[segment]
	if !ok {
		return Pass, false
	}
	if t, ok := sp.Fields[seq]; ok {
		return t, true
	}
	return sp.Default, true
}

// Default is the built-in conservative profile for the launch segment set.
// PID is strict (unmapped fields are masked). PV1 carries only specific
// provider-name and visit-id PHI, so its default is Pass.
func Default() Profile {
	return Profile{Segments: map[string]SegmentProfile{
		"PID": {
			Default: Mask,
			Fields: map[int]Treatment{
				1:  Pass,           // Set ID
				2:  Pseudonymize,   // Patient ID
				3:  Pseudonymize,   // Patient Identifier List
				4:  Pseudonymize,   // Alternate Patient ID
				5:  Pseudonymize,   // Patient Name
				6:  Pseudonymize,   // Mother's Maiden Name
				7:  GeneralizeDate, // Date/Time of Birth
				8:  Pass,           // Administrative Sex
				9:  Pseudonymize,   // Patient Alias
				10: Pass,           // Race
				11: Pseudonymize,   // Patient Address
				12: Pass,           // County Code
				13: Pseudonymize,   // Phone - Home
				14: Pseudonymize,   // Phone - Business
				15: Pass,           // Primary Language
				16: Pass,           // Marital Status
				17: Pass,           // Religion
				18: Mask,           // Patient Account Number
				19: Mask,           // SSN
				20: Mask,           // Driver's License
				21: Pseudonymize,   // Mother's Identifier
				22: Pass,           // Ethnic Group
				23: Pseudonymize,   // Birth Place
				24: Pass,           // Multiple Birth Indicator
				25: Pass,           // Birth Order
				26: Pass,           // Citizenship
				27: Pass,           // Veterans Military Status
				28: Pass,           // Nationality
				29: GeneralizeDate, // Death Date/Time
				30: Pass,           // Death Indicator
			},
		},
		"PV1": {
			Default: Pass,
			Fields: map[int]Treatment{
				7:  Pseudonymize, // Attending Doctor
				8:  Pseudonymize, // Referring Doctor
				9:  Pseudonymize, // Consulting Doctor
				17: Pseudonymize, // Admitting Doctor
				19: Pseudonymize, // Visit Number
				50: Pseudonymize, // Alternate Visit ID
				52: Pseudonymize, // Other Healthcare Provider
			},
		},
		"NK1": {
			Default: Pass,
			Fields: map[int]Treatment{
				2:  Pseudonymize,   // Name
				4:  Pseudonymize,   // Address
				5:  Pseudonymize,   // Phone
				6:  Pseudonymize,   // Business Phone
				12: Pseudonymize,   // Employee Number
				16: GeneralizeDate, // Date/Time of Birth
				26: Pseudonymize,   // Mother's Maiden Name
				30: Pseudonymize,   // Contact Person's Name
				31: Pseudonymize,   // Contact Person's Telephone
				32: Pseudonymize,   // Contact Person's Address
				33: Pseudonymize,   // Identifiers
				37: Mask,           // Social Security Number
				38: Pseudonymize,   // Birth Place
			},
		},
	}}
}
