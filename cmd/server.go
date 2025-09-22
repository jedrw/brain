package cmd

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/jedrw/brain/internal/brain"
	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Brainfiles server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := config.New(configPath, cmd.Flags())
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		brain, err := brain.NewBrain(ctx, config)
		if err != nil {
			return err
		}

		err = brain.Serve()
		if err != nil {
			cancel()
			return err
		}

		<-ctx.Done()
		shutdownCtx, shudownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shudownCancel()
		return brain.Shutdown(shutdownCtx)
	},
}

func init() {
	serverCmd.Flags().StringVarP(&contentDir, config.ContentDirFlag, "d", config.ContentDirDefault, "Path to content dir")
	serverCmd.Flags().StringVarP(&hostKeyPath, config.HostKeyPathFlag, "k", config.HostKeyPathDefault, "Path to host key")
	serverCmd.Flags().StringVarP(&authorizedKeys, config.AuthorizedKeysFlag, "z", "", "Authorized keys (comma separated)")
}
