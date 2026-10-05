package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/semantic"
	"github.com/sumvee/hl7lens/internal/tui"
)

var viewCmd = &cobra.Command{
	Use:   "view [file]",
	Short: "Open an interactive, semantic TUI tree of a message",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := argAt(args, 0)
		msg, err := parseSource(file)
		if err != nil {
			return err
		}
		d, err := dict.Load(defaultVersion)
		if err != nil {
			return err
		}

		title := "hl7lens — " + sourceLabel(file)
		model := tui.New(semantic.Build(msg, d), title)

		opts := []tea.ProgramOption{tea.WithAltScreen()}
		// When the message arrived on stdin, that stream is not the
		// keyboard; drive the TUI from the controlling terminal instead.
		if file == "" || file == "-" {
			if tty, err := os.Open("/dev/tty"); err == nil {
				opts = append(opts, tea.WithInput(tty))
			}
		}
		_, err = tea.NewProgram(model, opts...).Run()
		return err
	},
}

func sourceLabel(file string) string {
	if file == "" || file == "-" {
		return "stdin"
	}
	return file
}

func init() { rootCmd.AddCommand(viewCmd) }
