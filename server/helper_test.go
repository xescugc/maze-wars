package server_test

import (
	"io"

	"github.com/sagikazarmark/slog-shim"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/server"
)

func newEmptyLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func initStore() (*server.ActionDispatcher, *server.Store) {
	d := flux.NewDispatcher()
	l := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := server.NewStore(d, newEmptyLogger())
	return server.NewActionDispatcher(d, l, s), s
}
