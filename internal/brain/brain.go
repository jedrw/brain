package brain

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/anmitsu/go-shlex"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/cockroachdb/cmux"
	"github.com/jedrw/brain/internal/config"
	"github.com/jedrw/brain/internal/server"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
)

type Brain struct {
	config    config.Config
	ctx       context.Context
	tree      *Tree
	updater   chan<- struct{}
	sshServer *ssh.Server
}

var (
	ErrInvalidBrainNode = errors.New("invalid brain node")
	ErrNotExist         = errors.New("node does not exist")
	ErrNodeIsDir        = errors.New("node is directory")
	markdownParser      = goldmark.New(
		goldmark.WithExtensions(
			extension.Typographer,
			meta.New(),
		),
	)
)

func NewBrain(ctx context.Context, config config.Config) (*Brain, error) {
	b := &Brain{
		ctx:    ctx,
		config: config,
		tree:   &Tree{},
	}

	err := os.MkdirAll(b.config.ContentDir, 0770)
	if err != nil {
		return b, err
	}

	b.updater = b.Updater()
	b.updater <- struct{}{}

	return b, nil
}

func (b *Brain) getTree() error {
	b.tree.mu.Lock()
	defer b.tree.mu.Unlock()
	var err error
	b.tree.nodes, err = getNodes(b.config.ContentDir, "")
	if err != nil {
		return err
	}

	return nil
}

func (b *Brain) Updater() chan<- struct{} {
	updateChan := make(chan struct{}, 1)
	go func() {
		select {
		case <-b.ctx.Done():
			return
		default:
			for range updateChan {
				err := b.getTree()
				if err != nil {
					log.Fatal(fmt.Sprintf("could not update brain tree: %s", err))
				}

				for _, fnString := range b.config.UpdateTasks {
					fnArgs, err := shlex.Split(fnString, true)
					if err != nil {
						log.Warnf("failed to parse updater task: %s", err)
						continue
					}

					fn := exec.Command(fnArgs[0], fnArgs[1:]...)
					output, err := fn.CombinedOutput()
					if err != nil {
						log.Warnf("failed to run updater task: %s\n%s", err, string(output))
					}

					log.Info("ran updater task. ")
				}
			}
		}
	}()

	return updateChan
}

func (b *Brain) Serve() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", b.config.Port))
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("opened listener on port: %d", b.config.Port))

	b.sshServer, err = server.NewSSHServer(
		b.config.HostKeyPath,
		b.config.AuthorizedKeys,
		b.sshHandler,
	)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Info("starting ssh server")
		if err := b.sshServer.Serve(listener); err != nil && err != ssh.ErrServerClosed && !errors.Is(err, cmux.ErrListenerClosed) {
			log.Info("error from ssh server")
			log.Fatal(err)
		}
	}()

	return nil
}

func (b *Brain) Shutdown(ctx context.Context) error {
	log.Info("shutting down ssh server")
	err := b.sshServer.Shutdown(ctx)
	if err != nil && err != ssh.ErrServerClosed && !errors.Is(err, net.ErrClosed) {
		log.Info("error from ssh shutdown")
		return err
	}

	return nil
}
