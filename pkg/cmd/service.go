package cmd

import (
	"encoding/json"
	"io"

	"github.com/aserto-dev/aserto-test/pkg/cc"
	"github.com/aserto-dev/aserto-test/pkg/svc"
)

type ServiceCmd struct {
	Info     bool `name:"info" help:"display service info"`
	Ping     bool `name:"ping" help:"ping service instance"`
	Template bool `name:"template" alias:"t" help:"output service template"`
}

func (cmd *ServiceCmd) Run(c *cc.CommonCtx) error {
	if cmd.Template {
		return jsonOut(c.OutWriter, &svc.ServiceContext{})
	}

	if cmd.Info {
		return jsonOut(c.OutWriter, c.Service)
	}

	return nil
}

func jsonOut(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	enc.SetIndent("", "  ")
	return enc.Encode(&v)
}
