package cmd

import (
	"errors"

	"github.com/jedrw/brain/internal/config"
	"github.com/spf13/cobra"
)

var address string
var port int

var rootCmd = &cobra.Command{
	Use:   "brain [command]",
	Short: "Brain store",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// TODO read config for host
		if address == "" {
			return errors.New("set host address with -a flag or configure in config")
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config-path", "c", "", "Path to config file")
	rootCmd.PersistentFlags().IntVarP(&port, config.PortFlag, "p", 8080, "Port to listen on")
	rootCmd.PersistentFlags().StringVarP(&address, config.AddressFlag, "a", "", "Brain host address")
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(listCmd)
	cobra.EnableCommandSorting = false
}
