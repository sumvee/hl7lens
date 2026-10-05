package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/validate"
)

// Exit codes are part of the CLI contract (CI integration).
const exitValidationFailed = 2

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Check conformance: structure (T1) and message grammar (T2)",
	Long: `Validate a message against the v2.5.1 dictionary.

  T1 structural : required fields, cardinality, table membership, numeric
                  format
  T2 grammar    : required segments, segment cardinality and order for the
                  message's structure (from MSH-9)

Exit code is 0 when there are no errors (warnings are allowed) and 2 when
there is at least one error, so it can gate CI.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, err := parseSource(argAt(args, 0))
		if err != nil {
			return err
		}
		d, err := dict.Load(defaultVersion)
		if err != nil {
			return err
		}
		report := validate.Validate(msg, d)
		out := cmd.OutOrStdout()
		for _, f := range report.Findings {
			fmt.Fprintf(out, "%-5s  %-10s %s\n", f.Level, f.Path, f.Message)
		}
		errs, warns := report.Counts()
		if len(report.Findings) == 0 {
			fmt.Fprintln(out, "OK: no conformance issues found")
		} else {
			fmt.Fprintf(out, "\n%d error(s), %d warning(s)\n", errs, warns)
		}
		if report.HasErrors() {
			os.Exit(exitValidationFailed)
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(validateCmd) }
