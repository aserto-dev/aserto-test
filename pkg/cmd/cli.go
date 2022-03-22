package cmd

import (
	"fmt"
	"os"

	"github.com/aserto-dev/aserto-test/pkg/cc"
	"github.com/aserto-dev/aserto-test/pkg/version"
	"github.com/aserto-dev/aserto-test/pkg/x"
)

type CLI struct {
	Enum    EnumCmd    `cmd:"" help:"enumerate policy IDs"`
	List    ListCmd    `cmd:"" help:"list tests"`
	Run     RunCmd     `cmd:"" help:"run tests"`
	Service ServiceCmd `cmd:"" help:"service commands"`
	Profile *os.File   `name:"profile" short:"p" help:"service profile file"`
	Version VersionCmd `cmd:"" help:"version information"`
}

type Cmd struct {
	Verbose bool   `name:"verbose" short:"v" help:"verbose output"`
	Format  string `enum:"${formatText},${formatJSON}" default:"${formatText}" help:"output format [${formatText}|${formatJSON}]"`
}

type VersionCmd struct{}

func (cmd *VersionCmd) Run(c *cc.CommonCtx) error {
	fmt.Fprintf(c.OutWriter, "%s - %s\n",
		x.AppName,
		version.GetInfo().String(),
	)
	return nil
}
