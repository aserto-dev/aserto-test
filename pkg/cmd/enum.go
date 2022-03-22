package cmd

import (
	"context"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/authorizer"
	"github.com/aserto-dev/aserto-test/pkg/cc"
	"github.com/aserto-dev/aserto-test/pkg/printer"
	"github.com/aserto-dev/aserto-test/pkg/tester"
	"github.com/aserto-dev/aserto-test/pkg/x"

	"google.golang.org/grpc"
)

type EnumCmd struct {
	Cmd
}

func (cmd *EnumCmd) Run(c *cc.CommonCtx) error {
	if err := c.Service.Validate(); err != nil {
		return err
	}

	var (
		authzClient *authorizer.Client
		opts        []client.ConnectionOption
		err         error
	)

	opts = append(c.Service.ConnectionOpts(), client.WithURL(c.Service.GRPC()), client.WithDialOptions(grpc.WithBlock()))
	ctx, cancel := context.WithTimeout(c.Context, time.Duration(10)*time.Second)
	defer cancel()

	authzClient, err = authorizer.New(ctx, opts...)
	if err != nil {
		return err
	}

	tm := tester.NewManager("", authzClient)

	result, err := tm.Enum(c)
	if err != nil {
		return err
	}

	switch cmd.Format {
	case x.FormatText:
		printer.NewText(c.OutWriter).Print(result)
	case x.FormatJSON:
		printer.NewJSON(c.OutWriter).Print(result)
	}

	return nil
}
