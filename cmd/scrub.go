package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/scrub"
)

var scrubCmd = &cobra.Command{
	Use:   "scrub [file]",
	Short: "Emit a PHI-safe copy of a message (safe-by-default)",
	Long: `Remove PHI while preserving structure, so the result is safe to paste
into a ticket and still usable as test data.

Identifiers, names, addresses, and phones are pseudonymized (the same
input maps to the same fake within a run); SSN, account, and licence
numbers are hard-masked. In PHI segments, any field not explicitly passed
through is masked. By default a fresh random key is used so the output is
not reversible; pin --key to make it reproducible across files.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if p, _ := cmd.Flags().GetString("profile"); p != "" {
			return errors.New("scrub: custom --profile files are not supported yet (roadmap); the built-in profile is used")
		}
		msg, err := parseSource(argAt(args, 0))
		if err != nil {
			return err
		}

		var sc *scrub.Scrubber
		if k, _ := cmd.Flags().GetString("key"); k != "" {
			sum := sha256.Sum256([]byte(k))
			sc = scrub.New(scrub.Default(), sum[:])
		} else {
			sc, err = scrub.NewRandom(scrub.Default())
			if err != nil {
				return err
			}
		}
		fmt.Fprintln(cmd.OutOrStdout(), sc.Scrub(msg))
		return nil
	},
}

func init() {
	scrubCmd.Flags().String("profile", "", "PHI profile file (roadmap; not yet supported)")
	scrubCmd.Flags().String("key", "", "pin a pseudonymization key (default: random per run)")
	rootCmd.AddCommand(scrubCmd)
}
