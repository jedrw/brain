package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var address string
var port int

var rootCmd = &cobra.Command{
	Use:   "brain [command]",
	Short: "Brain store",
	Args:  cobra.NoArgs,
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
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	rootCmd.Flags().StringVarP(&address, "address", "a", "", "Brain host address")
	rootCmd.AddCommand(serverCmd)
	cobra.EnableCommandSorting = false
}
