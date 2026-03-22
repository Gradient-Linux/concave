package cmd

import (
	"context"

	"github.com/Gradient-Linux/concave/internal/api"
	"github.com/Gradient-Linux/concave/internal/auth"
	"github.com/spf13/cobra"
)

var (
	serveAddr = "127.0.0.1:7777"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the authenticated concave control-plane server",
	RunE:  runServe,
}

func runServe(cmd *cobra.Command, args []string) error {
	if err := ensureWorkspaceLayout(); err != nil {
		return err
	}
	tokenConfig, err := auth.LoadOrCreateTokenConfig(workspaceRoot())
	if err != nil {
		return err
	}

	ctx, stop := systemSignalHandler(context.Background(), nil)
	defer stop()

	server := api.New(api.Config{
		Addr:          serveAddr,
		Version:       displayVersion(),
		WorkspaceRoot: workspaceRoot(),
		Tokens:        tokenConfig,
	})
	return server.ListenAndServe(ctx)
}

func init() {
	serveCmd.Flags().StringVar(&serveAddr, "addr", serveAddr, "bind address for concave serve")
	rootCmd.AddCommand(serveCmd)
}
