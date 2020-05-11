package lxd

import (
	"context"

	_ "github.com/containous/traefik/v2/pkg/log"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/lxc/lxd/shared/api"
)

func (p *Provider) buildConfiguration(ctx context.Context, containers []api.Container) *dynamic.Configuration {
	//configurations := make(map[string]*dynamic.Configuration)

	/* for _, container := range containers {
		//ctxContainer := log.With(ctx, log.Str("container", container.Name))

	} */
	return nil
}
