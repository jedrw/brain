package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/cockroachdb/cmux"
	"github.com/jedrw/brain/server"
	"github.com/spf13/cobra"
)

var contentDir string
var hostKeyPath string
var authorizedKeys []string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Brain store server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("opened listener on port: %d", port))
		mux := cmux.New(listener)
		sshListener := mux.Match(cmux.PrefixMatcher("SSH-"))
		httpListener := mux.Match(cmux.Any())
		httpServer := server.NewHttpServer(contentDir)

		// TODO sort authorized keys config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		authorizedKey, err := os.ReadFile(path.Join(homeDir, ".ssh", "id_rsa.pub"))
		if err != nil {
			return err
		}

		sshServer, err := server.NewSSHServer(contentDir, hostKeyPath, []string{string(authorizedKey)})
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		go func() {
			err = mux.Serve()
			if err != nil && err != cmux.ErrListenerClosed && !errors.Is(err, net.ErrClosed) {
				log.Fatal(err)
			}
		}()

		go sshServer.Serve(sshListener)
		go httpServer.Serve(httpListener)

		<-ctx.Done()
		log.Info("shutting down")
		err = sshServer.Shutdown(cmd.Context())
		if err != nil {
			return err
		}

		err = httpServer.Shutdown(cmd.Context())
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	},
}

func init() {
	serverCmd.Flags().StringVarP(&contentDir, "content-dir", "c", "./content", "Path to content dir")
	serverCmd.Flags().StringVarP(&hostKeyPath, "host-key-path", "k", "./id_ed25519", "Path to host key")
}
