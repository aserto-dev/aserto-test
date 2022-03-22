package main

import (
	"context"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aserto-dev/aserto-test/pkg/cc"
	"github.com/aserto-dev/aserto-test/pkg/cmd"
	"github.com/aserto-dev/aserto-test/pkg/svc"
	"github.com/aserto-dev/aserto-test/pkg/x"
)

func main() {
	ct1, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	c := cc.New(ct1)

	cli := cmd.CLI{}
	ctx := kong.Parse(&cli,
		kong.Name(x.AppName),
		kong.Description(x.AppDescription),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary:        false,
			Summary:             true,
			Compact:             true,
			Tree:                false,
			FlagsLast:           true,
			Indenter:            kong.SpaceIndenter,
			NoExpandSubcommands: false,
		}),
		kong.Bind(&cli),
		kong.Bind(c),
		kong.Vars{
			"formatText": x.FormatText,
			"formatJSON": x.FormatJSON,
		},
	)

	if cli.Profile != nil {
		serviceContext, err := svc.FromReader(cli.Profile)
		if err != nil {
			ctx.FatalIfErrorf(err)
		}
		c.Service = serviceContext
	}

	err := ctx.Run(c)
	ctx.FatalIfErrorf(err)
}
