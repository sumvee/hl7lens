# Contributing to hl7lens

Thanks for your interest. hl7lens is Apache-2.0 licensed; by contributing
you agree your contribution is licensed under the same terms.

## Dev loop

```
  git clone https://github.com/sumvee/hl7lens && cd hl7lens
        |
     make build        -> ./hl7lens           (version via -ldflags)
     make test         -> go test ./...
     make vet          -> go vet ./...
     make lint         -> vet + test (what CI runs)
        |
     ./hl7lens explain testdata/adt_a01.hl7   try a change end to end
```

Requires Go 1.23+. No CGO, no external services.

## Where things live

```
  internal/hl7        parser (0 deps): message -> addressable tree
  internal/dict       the embedded v2.5.1 dictionary + loader
    data/<version>/   segments/ datatypes/ tables/ messages/  (JSON)
  internal/semantic   joins parse + dict -> annotation tree (explain, view)
  internal/scrub      PHI de-identification engine + profile
  internal/validate   T1 structural + T2 grammar conformance
  internal/tui        Bubble Tea view
  cmd/                 Cobra command surface
```

## The one rule for dictionary data (read before editing data/)

```
  ENCODE VERIFIED FACTS, NEVER GUESSES
    field/component metadata (name, data type, R/O/C, repeats, table)
    must be grounded against the published standard and, where they
    overlap, cross-checked against the Apache-2.0 HL7 v2-to-FHIR maps.

    - unsure of a table binding?        leave it BLANK, do not guess
    - table code meanings?              only from CC0 HL7 Terminology
    - never embed copyleft (HAPI) data or transcribed spec prose
    - the referential-integrity test MUST pass: every data type a field
      or component references has to be defined (go test ./internal/dict)
```

## Pull requests

```
  1 branch from main
  2 add/extend tests (integration-first; see the *_test.go in each pkg)
  3 make lint is green
  4 adding PHI-bearing fields? update internal/scrub profile too, and
    keep the NO-PHI-LEAK test passing
  5 one logical change per PR; describe the behavior, not the diff
```

## Reporting issues

Open a GitHub issue with a minimal, de-identified message
(`hl7lens scrub` your sample first) and the command you ran.
