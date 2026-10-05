package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/hl7"
)

var grepCmd = &cobra.Command{
	Use:   "grep <segment> [file]",
	Short: "Print segments of a type, optionally filtered by a field predicate",
	Long: `Print every segment of the named type. With -f, keep only segments
whose field matches a substring, e.g.:

  hl7 grep OBX -f 3=glucose msg.hl7     # OBX where field 3 contains "glucose"
  hl7 grep PID -f 5.1=SMITH msg.hl7     # PID where component 5.1 contains "SMITH"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		segName := strings.ToUpper(args[0])
		msg, err := parseSource(argAt(args, 1))
		if err != nil {
			return err
		}
		pred, _ := cmd.Flags().GetString("field")
		field, comp, want, err := parsePredicate(pred)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		matched := 0
		for _, seg := range msg.Segments {
			if seg.Name != segName {
				continue
			}
			if pred == "" || segmentMatches(seg, msg.Delims, field, comp, want) {
				fmt.Fprintln(out, seg.Raw(msg.Delims))
				matched++
			}
		}
		if matched == 0 {
			return fmt.Errorf("no %s segment matched", segName)
		}
		return nil
	},
}

// parsePredicate parses "N=val" or "N.C=val" into field, component, want.
func parsePredicate(p string) (field, comp int, want string, err error) {
	if p == "" {
		return 0, 0, "", nil
	}
	eq := strings.IndexByte(p, '=')
	if eq < 0 {
		return 0, 0, "", fmt.Errorf("predicate %q must be FIELD=VALUE (e.g. 3=glucose)", p)
	}
	lhs, want := p[:eq], p[eq+1:]
	parts := strings.SplitN(lhs, ".", 2)
	field, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, "", fmt.Errorf("predicate field %q is not a number", parts[0])
	}
	if len(parts) == 2 {
		comp, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, "", fmt.Errorf("predicate component %q is not a number", parts[1])
		}
	}
	return field, comp, want, nil
}

// segmentMatches reports whether any repetition of the segment's field
// (or component) contains want as a substring.
func segmentMatches(seg hl7.Segment, d hl7.Delimiters, field, comp int, want string) bool {
	f, ok := seg.Field(field)
	if !ok {
		return false
	}
	for _, rep := range f.Repetitions {
		var val string
		if comp == 0 {
			val = rep.Raw(d)
		} else if comp-1 < len(rep.Components) {
			val = rep.Components[comp-1].Raw(d)
		}
		if strings.Contains(val, want) {
			return true
		}
	}
	return false
}

func init() {
	grepCmd.Flags().StringP("field", "f", "", "field predicate, e.g. 3=glucose or 5.1=SMITH")
	rootCmd.AddCommand(grepCmd)
}
