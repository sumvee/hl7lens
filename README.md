# hl7lens

Read, validate, de-identify, and move HL7 v2.x messages from the command line.

```
  hl7 explain  adt.hl7           decode every segment, field, and code
  hl7 view     adt.hl7           interactive, color-coded tree (TUI)
  hl7 get      PID-5.1 adt.hl7   pull one value by path
  hl7 grep     OBX -f 3=glucose  filter segments by a field
  hl7 scrub    adt.hl7           PHI-safe copy, structure intact
  hl7 validate adt.hl7           conformance check, CI-friendly exit code
  hl7 send     host:2575 adt.hl7 send over MLLP, print the ACK
  hl7 listen   :2575             accept MLLP connections and ACK
```

hl7lens turns an opaque pipe-and-hat message into something a human can
read. It decodes segments, fields, data types, and HL7 code tables into
plain meaning, de-identifies messages so they are safe to share, checks
conformance against the v2.5.1 structure and message grammar, and sends
or receives messages over MLLP to exercise a real interface.

## Features

```
  explain   PID-8 "M"  ->  Administrative Sex = Male (table 0001)
            MSH-9       ->  Message Type / Trigger Event (table 0003)
            breaks composites into named components; --json for tooling

  view      interactive terminal tree: collapse segments, navigate,
            same decoding as explain

  get/grep  scriptable access by path (PID-5.1, PID-3[2].1, MSH-9.2)
            and segment filtering (OBX -f 3=glucose); get --json adds
            the field's name, data type, table, and decode

  scrub     removes PHI while keeping structure usable as test data:
            names/addresses/phones pseudonymized (deterministic),
            SSN/account hard-masked, dates generalized to the year.
            Safe by default: unknown fields in PHI segments are masked.

  validate  T1 structural (required fields, cardinality, table
            membership) + T2 grammar (required segments, order).
            Exit 0 clean, exit 2 on error.

  send      frame a message in MLLP, send over TCP, print the ACK;
  listen    non-zero exit on AE/AR. listen mocks a receiver and ACKs
            each message (--validate answers AE on a non-conformant one).
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

  # machine-readable output for pipelines
  hl7lens explain --json message.hl7 | jq '.segments[] | select(.name=="PID")'
  hl7lens get --json PID-8 message.hl7 | jq -r .decoded      # -> Male

  # make a sample safe to paste into a ticket
  hl7lens scrub message.hl7 > safe.hl7

  # gate a pipeline on conformance (exit 2 on error)
  hl7lens validate message.hl7 || echo "rejected"

  # exercise an interface over MLLP
  hl7lens send ehr.example.com:2575 message.hl7        # read the ACK
  hl7lens send ehr.example.com:2575 message.hl7 | hl7lens explain
  hl7lens listen :2575 --validate                      # mock a receiver

  # scrub real PHI before sending to a shared endpoint
  hl7lens scrub real.hl7 | hl7lens send test-host:2575
```

The binary is `hl7lens`; `hl7` works as a short alias.

## Supported

```
  version     HL7 v2.5.1
  messages    ADT (A01/A04/A08) · ORU (R01) · SIU (S12-S26) · MDM (T01-T11)
  segments    MSH EVN PID PV1 NK1 OBR OBX NTE AL1 DG1 GT1 IN1
              ORC SPM SCH RGS AIS AIG AIL AIP TXA
  tables      code tables decode coded fields (sex, marital, event type,
              result status, relationship, and more)
  transport   MLLP over TCP (send / listen) with ACK generation
```

The dictionary is a versioned, swappable dataset, so more versions,
segments, and tables are added as data without engine changes.

## Licensing and data provenance

hl7lens is licensed under Apache-2.0. Its embedded dictionary data is
clean and redistributable:

```
  code tables         HL7 Terminology (CC0, public domain); one table
                      (0078) derived from the Apache-2.0 v2-to-FHIR project
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
