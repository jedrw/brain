package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/cockroachdb/cmux"
	"github.com/jedrw/brain/server"
)

// TODO: read from flag
const port = 8080

var hostKeyPath string

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	log.Info(fmt.Sprintf("opened listener on port: %d", port))
	mux := cmux.New(listener)
	sshListener := mux.Match(cmux.PrefixMatcher("SSH-"))
	httpListener := mux.Match(cmux.Any())

	// TODO make content dir for httpserver dynamic
	httpServer := server.NewHttpServer()
	sshServer, err := server.NewSSHServer(hostKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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
	err = sshServer.Close()
	if err != nil {
		log.Error(err)
	}

	err = httpServer.Shutdown(context.Background())
	if err != nil && err != http.ErrServerClosed {
		log.Error(err)
	}

}
