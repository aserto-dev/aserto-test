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
	"github.com/pkg/errors"

	"google.golang.org/grpc"
)

type ListCmd struct {
	PolicyID string `required:""`
	Cmd
}

func (cmd *ListCmd) Run(c *cc.CommonCtx) error {
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

	tm := tester.NewManager(cmd.PolicyID, authzClient)
	if !tm.Ping(c) {
		return errors.Errorf("ping to authorizer failed")
	}

	result, err := tm.List(c)
	if err != nil {
		return err
	}

	switch cmd.Format {
	case x.FormatText:
		printer.NewText(c.OutWriter).Print(p(result))
	case x.FormatJSON:
		printer.NewJSON(c.OutWriter).Print(result)
	}

	return nil
}

func p(r []*tester.TestRule) []string {
	s := []string{}
	for _, x := range r {
		s = append(s, x.String())
	}
	return s
}
