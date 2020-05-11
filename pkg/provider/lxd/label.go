package lxd

import (
	"github.com/containous/traefik/v2/pkg/config/label"
)

type configuration struct {
	Enable bool
}

func (p *Provider) getConfiguration(container lxdData) (configuration, error) {
	conf := configuration{
		Enable: p.ExposedByDefault,
	}

	err := label.Decode(container.Labels, &conf, "user.traefik.enable")
	if err != nil {
		return configuration{}, err
	}

	return conf, nil
}
