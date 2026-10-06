package cmd

import (
	"encoding/json"
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

		if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			enc.SetEscapeHTML(false)
			return enc.Encode(view)
		}
		fmt.Fprint(cmd.OutOrStdout(), view.Text())
		return nil
	},
}

func init() {
	explainCmd.Flags().Bool("json", false, "emit the annotation tree as JSON")
	rootCmd.AddCommand(explainCmd)
}
