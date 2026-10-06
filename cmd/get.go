package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
)

type getResult struct {
	Path     string `json:"path"`
	Value    string `json:"value"`
	Name     string `json:"name,omitempty"`
	DataType string `json:"datatype,omitempty"`
	Table    string `json:"table,omitempty"`
	Decoded  string `json:"decoded,omitempty"`
}

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

		if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
			return emitGetJSON(cmd, args[0], val)
		}
		fmt.Fprintln(cmd.OutOrStdout(), val)
		return nil
	},
}

// emitGetJSON prints the value plus, for a field-level path, its dictionary
// metadata (name, data type, table) and any table decode.
func emitGetJSON(cmd *cobra.Command, path, val string) error {
	res := getResult{Path: path, Value: val}
	if a, err := hl7.ParsePath(path); err == nil && a.Component == 0 {
		if d, err := dict.Load(defaultVersion); err == nil {
			if fd, ok := d.Field(a.Segment, a.Field); ok {
				res.Name, res.DataType, res.Table = fd.Name, fd.DataType, fd.Table
				dt, hasDT := d.DataTypes[fd.DataType]
				if fd.Table != "" && (!hasDT || dt.Primitive) {
					if meaning, ok := d.Decode(fd.Table, val); ok {
						res.Decoded = meaning
					}
				}
			}
		}
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetEscapeHTML(false)
	return enc.Encode(res)
}

func init() {
	getCmd.Flags().Bool("json", false, "emit the value with its dictionary metadata as JSON")
	rootCmd.AddCommand(getCmd)
}
