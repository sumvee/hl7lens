package cmd

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/hl7"
	"github.com/sumvee/hl7lens/internal/mllp"
)

var sendCmd = &cobra.Command{
	Use:   "send <host:port> [file]",
	Short: "Send a message over MLLP and print the ACK",
	Long: `Frame a message in MLLP, send it over TCP to host:port, and print the
ACK that comes back. Exit code is non-zero if the ACK is not an accept
(MSA-1 other than AA/CA).`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := readSource(argAt(args, 1))
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}
		timeout, _ := cmd.Flags().GetDuration("timeout")

		conn, err := net.DialTimeout("tcp", args[0], timeout)
		if err != nil {
			return fmt.Errorf("dial %s: %w", args[0], err)
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(timeout))

		if _, err := conn.Write(mllp.Frame([]byte(toWire(string(raw))))); err != nil {
			return fmt.Errorf("send: %w", err)
		}
		ackBytes, err := mllp.NewReader(conn).ReadMessage()
		if err != nil {
			return fmt.Errorf("read ACK: %w", err)
		}
		ack := string(ackBytes)
		fmt.Fprintln(cmd.OutOrStdout(), strings.ReplaceAll(ack, "\r", "\n"))

		if am, err := hl7.Parse(ack); err == nil {
			if code, _ := am.Get("MSA-1"); code != "" && code != "AA" && code != "CA" {
				return fmt.Errorf("ACK not accepted (MSA-1 = %s)", code)
			}
		}
		return nil
	},
}

// toWire normalizes line endings to CR, as HL7 requires on the wire.
func toWire(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", "\r")
	return strings.Trim(s, "\r")
}

func init() {
	sendCmd.Flags().Duration("timeout", 10*time.Second, "dial/read timeout")
	rootCmd.AddCommand(sendCmd)
}
