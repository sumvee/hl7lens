// Package cmd wires the hl7lens command surface (Cobra).
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is overwritten at build time via -ldflags.
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "hl7lens",
	Short: "Read, validate, and safely share HL7 v2.x messages",
	Long: `hl7lens turns an opaque HL7 v2.x message into something a human can
read, validate, and share.

  view      open an interactive, semantic TUI tree
  explain   annotate every segment/field/table (non-TUI)
  get       print one value by path (e.g. PID-5.1)
  grep      filter segments by a field predicate
  scrub     emit a PHI-safe copy of a message
  validate  check conformance (structure + message grammar)

Reads a file argument or stdin; most verbs are pipeable.`,
	Version:       version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "hl7lens:", err)
		os.Exit(1)
	}
}
