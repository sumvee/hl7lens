# Changelog

All notable changes to hl7lens are documented here.
Format follows Keep a Changelog; versions follow Semantic Versioning.

## [Unreleased]

### Added
- Code tables (CC0 HL7 Terminology): 0003 Event Type (decodes MSH-9
  trigger and EVN-1), 0004 Patient Class, 0007 Admission Type, 0009
  Ambulatory Status, 0063 Relationship, 0069 Hospital Service.
- Code table 0078 Abnormal Flags / Interpretation Codes (decodes OBX-8),
  derived from the Apache-2.0 HL7 v2-to-FHIR ConceptMap.
- Tables carry an `open` flag: user-defined (suggested) tables are
  decode-only, so validate no longer flags legitimate site codes as
  "not in table". Existing 0001/0002 marked open accordingly.

### Fixed
- `go install`-ed builds now report the module version (from build info)
  instead of "dev"; release builds keep their `-ldflags` version.

## [0.1.0]

```
FIRST PUBLIC RELEASE
  v2.5.1 · message types ADT (A01/A04/A08) and ORU (R01)
  commands  explain · get · grep · scrub · validate · view
```

### Added
- `explain`: decode every segment, field, data type, and coded value to
  plain meaning (PID-8 "M" -> "Administrative Sex = Male", table 0001).
- `get`: field-path addressing (PID-5.1, PID-3[2].1, MSH-9.2) over a file
  or stdin.
- `grep`: select segments by type with an optional field predicate
  (3=glucose, 5.1=SMITH).
- `scrub`: dictionary-driven PHI de-identification. Pseudonymizes names,
  addresses, phones, and identifiers (deterministic, keyed, referential);
  hard-masks SSN/account/licence; generalizes dates of birth/death to the
  year. Safe by default: unmapped fields in a strict segment are masked.
- `validate`: conformance checking. T1 structural (required fields,
  cardinality, table membership, numeric format) and T2 grammar (required
  segments, cardinality, order). Exit code 2 on error for CI.
- `view`: interactive terminal tree (segments, fields, components) with
  color, collapse, and scrolling.
- Embedded v2.5.1 dictionary: 8 segments (222 fields), 41 data types
  (157 components), 4 code tables, 2 message grammars. Version is a
  swappable dataset directory.
- Clean data provenance: CC0 HL7 Terminology tables and Apache-2.0
  HL7 v2-to-FHIR cross-check; no copyleft (HAPI) data embedded.
- Single static binary via GoReleaser; Homebrew tap formula.

### Notes
- OBX is scoped to the v2.5.1 fields (1-19).
- A few rare composites (SPS, TQ, NDL) are carried opaque (not yet broken
  into components).
- Table 0125 (Value Type) is not yet embedded, so OBX-2 is not decoded.

[Unreleased]: https://github.com/sumvee/hl7lens/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/sumvee/hl7lens/releases/tag/v0.1.0
