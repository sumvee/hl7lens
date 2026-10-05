package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <path> [file]",
	Short: "Print one value by path (e.g. PID-5.1)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, err := parseSource(argAt(args, 1))
		if err != nil {
			return err
		}
		val, err := msg.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), val)
		return nil
	},
}

func init() { rootCmd.AddCommand(getCmd) }
