package tscaddy

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"
)

type TailscaleCaddyTransport struct {
	logger *zap.Logger
	server *tsnetServerDestructor
}

func (t *TailscaleCaddyTransport) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

func (t *TailscaleCaddyTransport) Provision(context caddy.Context) error {
	t.logger = context.Logger()
	fmt.Fprintln(os.Stderr, "Provision")

	s, err := getServer("", "caddy-tsnet-client:80")
	if err != nil {
		return err
	}

	s.Ephemeral = true
	s.Logf = func(format string, args ...any) {
		t.logger.Debug(fmt.Sprintf(format, args))
	}

	if err := s.Start(); err != nil {
		return err
	}
	t.server = s

	return nil
}

func (t *TailscaleCaddyTransport) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.reverse_proxy.transport.tailscale",
		New: func() caddy.Module {
			return new(TailscaleCaddyTransport)
		},
	}
}

func (t *TailscaleCaddyTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Scheme == "" {
		request.URL.Scheme = "http"
	}
	return t.server.HTTPClient().Transport.RoundTrip(request)
}

func (t *TailscaleCaddyTransport) Cleanup() error {
	t.server.Logf("cleaning up tsnet")
	fmt.Fprintln(os.Stderr, "called")
	t.server.Destruct()
	return errors.New("Force this error")
}

var (
	_ http.RoundTripper     = (*TailscaleCaddyTransport)(nil)
	_ caddy.Provisioner     = (*TailscaleCaddyTransport)(nil)
	_ caddy.CleanerUpper    = (*TailscaleCaddyTransport)(nil)
	_ caddyfile.Unmarshaler = (*TailscaleCaddyTransport)(nil)
)
