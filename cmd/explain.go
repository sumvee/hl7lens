package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/semantic"
)

var explainCmd = &cobra.Command{
	Use:   "explain [file]",
	Short: "Annotate every segment, field, and coded value (non-TUI)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, err := parseSource(argAt(args, 0))
		if err != nil {
			return err
		}
		d, err := dict.Load(defaultVersion)
		if err != nil {
			return err
		}
		view := semantic.Build(msg, d)
		fmt.Fprint(cmd.OutOrStdout(), view.Text())
		return nil
	},
}

func init() { rootCmd.AddCommand(explainCmd) }
