package cmd

import (
	"context"
	"time"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/aserto-dev/aserto-go/client/authorizer"
	"github.com/aserto-dev/aserto-test/pkg/cc"
	"github.com/aserto-dev/aserto-test/pkg/tester"
	"github.com/aserto-dev/aserto-test/pkg/x"
	"github.com/pkg/errors"
)

type RunCmd struct {
	PolicyID string `required:""`
	Cmd
}

func (cmd *RunCmd) Run(c *cc.CommonCtx) error {
	if err := c.Service.Validate(); err != nil {
		return err
	}

	var (
		authzClient *authorizer.Client
		opts        []client.ConnectionOption
		err         error
	)

	opts = append(c.Service.ConnectionOpts(), client.WithURL(c.Service.GRPC()))
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

	result, err := tm.Run(c)
	if err != nil {
		return err
	}

	switch cmd.Format {
	case x.FormatText:
		pr := tester.PrettyReporter{
			Output:      c.OutWriter,
			Verbose:     cmd.Verbose,
			FailureLine: false,
		}
		return pr.Report(result)
	case x.FormatJSON:
		jr := tester.JSONReporter{
			Output: c.OutWriter,
		}
		return jr.Report(result)
	}

	return nil
}
