# Changelog

All notable changes to hl7lens are documented here.
Format follows Keep a Changelog; versions follow Semantic Versioning.

## [Unreleased]

### Added
- MLLP test harness: `send <host:port> [file]` frames a message over TCP
  and prints the ACK (non-zero exit on AE/AR), and `listen [addr]` accepts
  MLLP connections, prints each message, and returns an ACK (with
  `--validate`, failing messages are answered AE). Includes MLLP framing,
  a stream reader, and ACK generation. Dependency-free (net stdlib).
- Message type MDM (medical document management): TXA (Transcription
  Document Header) segment and the MDM_T02 grammar covering triggers
  T01-T11, plus the PPN data type (opaque). Trigger events already decoded
  via table 0003.

## [0.1.2] - 2026-10-06

```
  SIU fully fleshed out (resource segments) + machine-readable --json
```

### Added
- SIU resource segments RGS (Resource Group), AIS (Service), AIG (General
  Resource), AIL (Location Resource), and AIP (Personnel Resource); the
  SIU scheduling message is now fully decoded end to end.
- `--json` output: `explain --json` emits the full annotation tree and
  `get --json` emits the value with its dictionary metadata (name, data
  type, table, decode), for pipelines and other tools.

## [0.1.1] - 2026-10-06

```
  dictionary grows: 15 segments · 46 data types · 11 code tables ·
  3 message types (ADT, ORU, SIU)
```

### Added
- Code tables (CC0 HL7 Terminology): 0003 Event Type (decodes MSH-9
  trigger and EVN-1), 0004 Patient Class, 0007 Admission Type, 0009
  Ambulatory Status, 0063 Relationship, 0069 Hospital Service.
- Code table 0078 Abnormal Flags / Interpretation Codes (decodes OBX-8),
  derived from the Apache-2.0 HL7 v2-to-FHIR ConceptMap.
- Segments AL1 (Patient Allergy Information) and DG1 (Diagnosis), plus the
  CP (Composite Price) data type; both are already allowed by the ADT
  grammar, so explain and validate now handle them fully.
- Segments GT1 (Guarantor) and IN1 (Insurance), plus the AUI (Authorization
  Information) data type; both added to the ADT grammar as optional,
  repeating.
- Segments ORC (Common Order) and SPM (Specimen), plus the CNE (Coded with
  No Exceptions) data type. ORC added to the ORU grammar; SPM is defined
  for future specimen-oriented messages (not part of v2.5.1 ORU_R01).
- Message type SIU (scheduling): SCH (Scheduling Activity Information)
  segment and the SIU_S12 grammar covering triggers S12-S26.
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

[Unreleased]: https://github.com/sumvee/hl7lens/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/sumvee/hl7lens/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/sumvee/hl7lens/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/sumvee/hl7lens/releases/tag/v0.1.0
