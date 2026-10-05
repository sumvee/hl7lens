package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/sumvee/hl7lens/internal/hl7"
)

// defaultVersion is the dictionary version used until message-driven
// version selection (from MSH-12) lands with multi-version datasets.
const defaultVersion = "2.5.1"

// readSource returns the raw message bytes from a file, or from stdin when
// file is empty or "-".
func readSource(file string) ([]byte, error) {
	if file == "" || file == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(file)
}

// parseSource reads and parses a message from a file or stdin.
func parseSource(file string) (*hl7.Message, error) {
	raw, err := readSource(file)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}
	msg, err := hl7.Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return msg, nil
}

// argAt returns args[i] or "" when absent.
func argAt(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}
