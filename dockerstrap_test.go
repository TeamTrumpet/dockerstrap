package dockerstrap_test

import (
	"github.com/TeamTrumpet/trumpet/helpers/docker"
	"testing"
)

func TestStartContainers(t *testing.T) {
	docker.StartContainers(&docker.Container{
		Name:  "postgres",
		Image: "postgres:latest",
		Ports: []string{
			"5432",
		},
	})
}
