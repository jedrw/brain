package brain

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/cockroachdb/cmux"
	"github.com/jedrw/brain/internal/config"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
)

type Node struct {
	Title    string
	Tags     []string
	Raw      []byte
	Content  []byte
	Path     string
	IsDir    bool
	Children []*Node
}

type Tree struct {
	mu    sync.RWMutex
	nodes []*Node
}

type Brain struct {
	config     config.Config
	ctx        context.Context
	tree       *Tree
	updater    chan<- struct{}
	httpServer *http.Server
	sshServer  *ssh.Server
}

var (
	ErrInvalidBrainNode = errors.New("invalid brain node")

	markdownParser = goldmark.New(
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
	mux := cmux.New(listener)
	sshListener := mux.Match(cmux.PrefixMatcher("SSH-"))
	httpListener := mux.Match(cmux.Any())
	go func() {
		if err := mux.Serve(); err != nil && !errors.Is(err, cmux.ErrListenerClosed) && !errors.Is(err, net.ErrClosed) {
			log.Fatal(err)
		}

	}()

	b.httpServer = NewHttpServer(b)
	go func() {
		log.Info("starting http server")
		if err := b.httpServer.Serve(httpListener); err != nil && err != http.ErrServerClosed {
			log.Info("error from http serve")
			log.Fatal(err)
		}
	}()

	if !b.config.NoSSH {
		b.sshServer, err = NewSSHServer(b)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			log.Info("starting ssh server")
			if err := b.sshServer.Serve(sshListener); err != nil && err != ssh.ErrServerClosed && !errors.Is(err, cmux.ErrListenerClosed) {
				log.Info("error from ssh serve")
				log.Fatal(err)
			}
		}()
	}

	return nil
}

func (b *Brain) Shutdown(ctx context.Context) error {
	log.Info("shutting down http server")
	err := b.httpServer.Shutdown(ctx)
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	if !b.config.NoSSH && b.sshServer != nil {
		log.Info("shutting down ssh server")
		err = b.sshServer.Shutdown(ctx)
		if err != nil && err != ssh.ErrServerClosed && !errors.Is(err, net.ErrClosed) {
			log.Info("error from ssh shutdown")
			return err
		}
	}

	return nil
}
