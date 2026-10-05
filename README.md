# hl7lens

Read, validate, and safely share HL7 v2.x messages from the command line.

```
  hl7 explain  adt.hl7          decode every segment, field, and code
  hl7 view     adt.hl7          interactive, color-coded tree (TUI)
  hl7 get      PID-5.1 adt.hl7  pull one value by path
  hl7 grep     OBX -f 3=glucose filter segments by a field
  hl7 scrub    adt.hl7          PHI-safe copy, structure intact
  hl7 validate adt.hl7          conformance check, CI-friendly exit code
```

hl7lens turns an opaque pipe-and-hat message into something a human can
read. It decodes segments, fields, data types, and HL7 code tables into
plain meaning, de-identifies messages so they are safe to share, and
checks conformance against the v2.5.1 structure and message grammar.

## Features

```
  explain   PID-8 "M"  ->  Administrative Sex = Male (table 0001)
            MSH-9       ->  Message Type / Trigger Event A01
            breaks composites into named components

  view      interactive terminal tree: collapse segments, navigate,
            same decoding as explain

  get/grep  scriptable access by path (PID-5.1, PID-3[2].1, MSH-9.2)
            and segment filtering (OBX -f 3=glucose)

  scrub     removes PHI while keeping structure usable as test data:
            names/addresses/phones pseudonymized (deterministic),
            SSN/account hard-masked, dates generalized to the year.
            Safe by default: unknown fields in PHI segments are masked.

  validate  T1 structural (required fields, cardinality, table
            membership) + T2 grammar (required segments, order).
            Exit 0 clean, exit 2 on error.
```

## Install

```
  brew install sumvee/tap/hl7lens
  go install github.com/sumvee/hl7lens@latest
```

Prebuilt binaries for macOS, Linux, and Windows are attached to each
release. Building from source needs Go 1.23+:

```
  git clone https://github.com/sumvee/hl7lens && cd hl7lens
  make build        # -> ./hl7lens
```

## Usage

```
  # read a file, or pipe on stdin
  hl7lens explain message.hl7
  cat message.hl7 | hl7lens get PID-5.1

  # make a sample safe to paste into a ticket
  hl7lens scrub message.hl7 > safe.hl7

  # gate a pipeline on conformance
  hl7lens validate message.hl7 || echo "rejected"
```

The binary is `hl7lens`; `hl7` works as a short alias.

## Supported

```
  version    HL7 v2.5.1
  messages   ADT (A01/A04/A08) · ORU (R01)
  segments   MSH EVN PID PV1 NK1 OBR OBX NTE
```

## Licensing and data provenance

hl7lens is licensed under Apache-2.0. Its embedded dictionary data is
clean and redistributable:

```
  code tables         HL7 Terminology (CC0, public domain)
  segment/structure   verified against the v2.5.1 standard, cross-checked
                      against the Apache-2.0 HL7 v2-to-FHIR project
```

It embeds no copyleft (HAPI) data and no copy of the HL7 v2
specification text. See `NOTICE` for details. "HL7" is used nominatively;
this project is not affiliated with or endorsed by HL7 International.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and pull requests welcome;
please de-identify any sample message (`hl7lens scrub`) before sharing it.

## License

Apache-2.0. Copyright Sumit Vig.
