package global

import (
	"context"

	"github.com/docker/docker/client"
)

var (
	DockerClient *client.Client
	Ctx          context.Context
)
