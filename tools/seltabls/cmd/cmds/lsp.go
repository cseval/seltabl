package cmds

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/conneroisu/seltabl/tools/seltabls/pkg/analysis"
	"github.com/conneroisu/seltabl/tools/seltabls/pkg/rpc"
	"github.com/spf13/cobra"
)

// LSPHandler is a struct for the LSP server
type LSPHandler func(ctx context.Context, writer *io.Writer, state *analysis.State, msg []byte) error

// NewLSPCmd creates a new command for the lsp subcommand
func NewLSPCmd(ctx context.Context, writer io.Writer, handle LSPHandler) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp", // the name of the command
		Short: "A command line tooling for package that parsing html tables and elements into structs",
		Long: `
CLI and Language Server for the seltabl package.

Language server provides completions, hovers, and code actions for seltabl defined structs.
	
CLI provides a command line tool for verifying, linting, and reporting on seltabl defined structs.
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(rpc.Split)
			state, err := analysis.NewState()
			if err != nil {
				return fmt.Errorf("failed to create state: %w", err)
			}
			for scanner.Scan() {
				_, cancel := context.WithCancel(ctx)
				defer cancel()
				msg := scanner.Bytes()
				_, _, err := rpc.DecodeMessage(msg)
				if err != nil {
					return fmt.Errorf("failed to decode message: %w", err)
				}
				err = handle(ctx, &writer, &state, msg)
				if err != nil {
					return fmt.Errorf("failed to handle message: %w", err)
				}
				if err := scanner.Err(); err != nil {
					return fmt.Errorf("scanner error: %w", err)
				}
			}
			return nil
		},
	}
	return cmd
}
