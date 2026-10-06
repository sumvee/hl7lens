package cmd

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/spf13/cobra"

	"github.com/sumvee/hl7lens/internal/dict"
	"github.com/sumvee/hl7lens/internal/hl7"
	"github.com/sumvee/hl7lens/internal/mllp"
	"github.com/sumvee/hl7lens/internal/validate"
)

var listenCmd = &cobra.Command{
	Use:   "listen [addr]",
	Short: "Accept MLLP connections, print each message, and ACK",
	Long: `Listen for MLLP connections (default :2575). For each message received,
print a summary line and return an ACK. With --validate, each message is
checked and a failing one is answered with an application error (AE).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr := argAt(args, 0)
		if addr == "" {
			addr = ":2575"
		}
		doValidate, _ := cmd.Flags().GetBool("validate")

		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("listen %s: %w", addr, err)
		}
		defer ln.Close()

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "hl7lens listening on %s (MLLP); Ctrl-C to stop\n", ln.Addr())

		var d *dict.Dictionary
		if doValidate {
			if d, err = dict.Load(defaultVersion); err != nil {
				return err
			}
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				return err
			}
			go serveConn(out, conn, d)
		}
	},
}

func serveConn(out io.Writer, conn net.Conn, d *dict.Dictionary) {
	defer conn.Close()
	r := mllp.NewReader(conn)
	for {
		payload, err := r.ReadMessage()
		if err != nil {
			return
		}
		msg, perr := hl7.Parse(string(payload))
		if perr != nil {
			fmt.Fprintf(out, "recv: unparseable message (%v)\n", perr)
			_, _ = conn.Write(mllp.Frame([]byte(nak())))
			continue
		}
		mt, _ := msg.Get("MSH-9")
		ctrl, _ := msg.Get("MSH-10")
		fmt.Fprintf(out, "recv: %s ctrl=%s\n", mt, ctrl)

		code := "AA"
		if d != nil {
			rep := validate.Validate(msg, d)
			if rep.HasErrors() {
				errs, _ := rep.Counts()
				fmt.Fprintf(out, "  validation: %d error(s) -> AE\n", errs)
				code = "AE"
			}
		}
		_, _ = conn.Write(mllp.Frame([]byte(mllp.BuildACK(msg, code))))
	}
}

// nak is a generic reject ACK for a message that could not be parsed.
func nak() string {
	ts := time.Now().Format("20060102150405")
	return "MSH|^~\\&|HL7LENS|||||" + ts + "||ACK|1|P|2.5.1\rMSA|AR|"
}

func init() {
	listenCmd.Flags().Bool("validate", false, "validate each message and reject failures with AE")
	rootCmd.AddCommand(listenCmd)
}
