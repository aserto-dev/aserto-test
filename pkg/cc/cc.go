package cc

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/aserto-dev/aserto-test/pkg/svc"
)

type CommonCtx struct {
	Context   context.Context
	OutWriter io.Writer
	ErrWriter io.Writer
	Service   *svc.ServiceContext
}

func New(ctx context.Context) *CommonCtx {
	log.SetOutput(io.Discard)
	log.SetPrefix("")
	log.SetFlags(log.LstdFlags)

	service := &svc.ServiceContext{}

	return &CommonCtx{
		Context:   ctx,
		OutWriter: os.Stdout,
		ErrWriter: os.Stderr,
		Service:   service,
	}
}
