package cmd

import (
	"errors"

	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var (
	brainConfig    config.Config
	configPath     string
	address        string
	port           int
	contentDir     string
	hostKeyPath    string
	authorizedKeys string
	keyPath        string
)

var rootCmd = &cobra.Command{
	Use:   "brain [command]",
	Short: "Brainfiles",
	Args:  cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		brainConfig, err = config.New(configPath, cmd.Flags())

		return err
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		if brainConfig.Address == "" {
			return errors.New("set host address with -a flag or configure in config")
		}

		return cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", "", "Path to config file")
	rootCmd.PersistentFlags().IntVarP(&port, config.PortFlag, "p", config.PortDefault, "Port to use to listen/connect")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(moveCmd)
	rootCmd.AddCommand(deleteCmd)
	cobra.EnableCommandSorting = false
}
