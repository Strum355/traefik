package lxd

import (
	"context"
	"fmt"
	"net/url"
	"text/template"
	"time"

	"github.com/lxc/lxd/shared/api"

	"github.com/containous/traefik/v2/pkg/version"

	lxd "github.com/lxc/lxd/client"

	"github.com/cenkalti/backoff/v4"
	"github.com/containous/traefik/v2/pkg/job"
	"github.com/containous/traefik/v2/pkg/log"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
)

const DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

type Provider struct {
	Endpoint         string `description:"LXD server endpoint. Can be a HTTP or a unix socket endpoint." json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	ExposedByDefault bool   `description:"Expose containers by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	DefaultRule      string `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`
	defaultRuleTpl   *template.Template
}

var _ provider.Provider = (*Provider)(nil)

func (p *Provider) SetDefaults() {
	p.Endpoint = "unix:///var/lib/lxd/unix.socket"
	p.ExposedByDefault = true
	p.DefaultRule = DefaultTemplateRule
}

func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %v", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return nil
}

type lxdData struct {
	Name    string
	Labels  map[string]string
	Network api.Network
}

func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		ctxLog := log.With(routineCtx, log.Str(log.ProviderName, "lxd"))
		logger := log.FromContext(ctxLog)

		operation := func() error {
			ctx, cancel := context.WithCancel(ctxLog)
			defer cancel()

			ctx = log.With(ctx, log.Str(log.ProviderName, "lxd"))

			lxdClient, err := p.createClient()
			if err != nil {
				return err
			}

			p.listContainers(ctx, lxdClient)

			return nil
		}

		notify := func(err error, time time.Duration) {
			logger.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}

		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), ctxLog), notify)
		if err != nil {
			logger.Error("Cannot connect to LXD server %+v", err)
		}
	})
	return nil
}

func (p *Provider) createClient() (lxd.InstanceServer, error) {
	u, err := url.Parse(p.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing endpoint url: %w", err)
	}

	if u.Scheme == "unix" {
		return lxd.ConnectLXDUnix(u.Path, &lxd.ConnectionArgs{
			UserAgent: "Traefik " + version.Version,
		})
	}

	return nil, fmt.Errorf("unsupported endpoint type %s", u.Scheme)
}

func (p *Provider) listContainers(ctx context.Context, client lxd.InstanceServer) ([]lxdData, error) {
	containers, err := client.GetContainersFull()
	if err != nil {
		return nil, err
	}

	var inspectedContainers []lxdData
	for _, container := range containers {
		lData := p.parseContainerData(container)

		p.getConfiguration(lData)

		inspectedContainers = append(inspectedContainers, lData)
	}
	return inspectedContainers, nil
}

func (p *Provider) parseContainerData(container api.ContainerFull) lxdData {

}
