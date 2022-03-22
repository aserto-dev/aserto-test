package svc

import (
	"encoding/json"
	"io"
	"net/url"

	"github.com/aserto-dev/aserto-go/client"
	"github.com/pkg/errors"
)

type ServiceContext struct {
	GRPCAddr string `json:"grpc_addr"`
	HTTPAddr string `json:"http_addr"`
	TenantID string `json:"tenant_id"`
	APIKey   string `json:"api_key"`
	Token    string `json:"token"`
	CACert   string `json:"ca_cert"`
	Insecure bool   `json:"insecure"`
}

func (sc *ServiceContext) ConnectionOpts() []client.ConnectionOption {
	opts := []client.ConnectionOption{}

	if sc.TenantID != "" {
		opts = append(opts, client.WithTenantID(sc.TenantID))
	}
	if sc.APIKey != "" {
		opts = append(opts, client.WithAPIKeyAuth(sc.APIKey))
	}
	if sc.Token != "" {
		opts = append(opts, client.WithTokenAuth(sc.Token))
	}
	if sc.CACert != "" {
		opts = append(opts, client.WithCACertPath(sc.CACert))
	}

	opts = append(opts, client.WithInsecure(sc.Insecure))

	return opts
}

func (sc *ServiceContext) Validate() error {
	if sc == nil {
		return errors.Errorf("service profile not loaded")
	}
	if sc.GRPCAddr == "" && sc.HTTPAddr == "" {
		return errors.Errorf("service profile does not contain a host address")
	}
	if sc.APIKey == "" && sc.Token == "" {
		return errors.Errorf("service profile does not contain a access token")
	}
	if sc.TenantID == "" {
		return errors.Errorf("service profile does not contain a tenant id")
	}
	return nil
}

func (sc *ServiceContext) HTTP() *url.URL {
	if sc.HTTPAddr != "" {
		u, err := url.Parse(sc.HTTPAddr)
		if err != nil {
			return nil
		}
		return u
	}
	return nil
}

func (sc *ServiceContext) GRPC() *url.URL {
	if sc.GRPCAddr != "" {
		u, err := url.Parse(sc.GRPCAddr)
		if err != nil {
			return nil
		}
		return u
	}
	return nil
}

func FromReader(r io.Reader) (*ServiceContext, error) {
	var svc ServiceContext
	dec := json.NewDecoder(r)
	err := dec.Decode(&svc)
	return &svc, err
}
