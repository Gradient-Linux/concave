package cmd

import (
	"context"

	"github.com/Gradient-Linux/concave/internal/suite"
)

func composeServiceStatus(ctx context.Context, s suite.Suite, service string) (string, error) {
	return dockerComposeServiceStatus(ctx, dockerComposePath(s.Name), service)
}

func composePrimaryStatus(ctx context.Context, s suite.Suite) (string, error) {
	container := primaryContainer(s)
	if container == "" {
		return "not found", nil
	}
	return composeServiceStatus(ctx, s, container)
}
