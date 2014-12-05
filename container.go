package dockerstrap

import (
	"github.com/fsouza/go-dockerclient"
	"log"
	"strings"
	"time"
)

type Container struct {
	Name  string   `yaml:"name"`
	Image string   `yaml:"image"`
	Ports []string `yaml:"ports"`
}

func (c *Container) isUp(client *docker.Client) (error, bool) {
	containers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
	})

	if err != nil {
		return err, false
	}

	// Check the running containers to see if the postgres container is running
	for _, container := range containers {
		for _, containerName := range container.Names {
			if strings.Contains(containerName, c.Name) && container.Image == c.Image && strings.HasPrefix(container.Status, "Up") {
				return nil, true
			}
		}
	}

	return nil, false
}

func (c *Container) Start(client *docker.Client) error {
	// Map of port to port bindings
	portBindings := make(map[docker.Port][]docker.PortBinding)

	for _, port := range c.Ports {
		tcpPort := docker.Port(port + "/tcp")

		postgresPortSA := []docker.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: port,
			},
		}
		portBindings[tcpPort] = postgresPortSA
	}

	err := client.StartContainer(c.Name, &docker.HostConfig{
		PortBindings: portBindings,
	})

	if err != nil {
		return err
	}

	// Sleep to allow processes to start
	// TODO: Fix
	time.Sleep(3 * time.Second)

	return nil
}

func (c *Container) Create(client *docker.Client) error {
	port_mappings := make(map[docker.Port]struct{})
	var blank_struct struct{}

	for _, port := range c.Ports {
		postgresPort := docker.Port(port + "/tcp")
		port_mappings[postgresPort] = blank_struct
	}

	// Create the container
	_, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: c.Name,
		Config: &docker.Config{
			Image:        c.Image,
			ExposedPorts: port_mappings,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Container) Setup() error {
	// Get the client for the docker daemon
	client := getDockerClient()

	err, running := c.isUp(client)

	if err != nil {
		return err
	}

	if running {
		return nil
	}

	err = c.Start(client)

	if _, ok := err.(*docker.ContainerAlreadyRunning); ok {
		// If it is already running, ok...
		return nil
	} else if _, ok := err.(*docker.NoSuchContainer); ok {
		// Docker container not created, must create
		err = c.Create(client)
		if err == docker.ErrNoSuchImage {
			// Pull the image
			c.Pull(client)

			// Create the container
			err = c.Create(client)
			if err != nil {
				return err
			}

			// Call again to actually start it
			return c.Start(client)
		} else if err != nil {
			return err
		}

		// Call again to actually start it
		return c.Start(client)
	} else if err != nil {
		return err
	}

	// All was good
	return nil
}

func (c *Container) Pull(client *docker.Client) error {
	log.Printf("Downloading image for %v, this may take a few minutes\n", c)
	err := client.PullImage(
		docker.PullImageOptions{
			Repository: c.Image,
		},
		docker.AuthConfiguration{},
	)

	if err != nil {
		log.Printf("Download failed for %v\n", c)
		return err
	}

	log.Printf("Download complete for %v\n", c)

	return nil
}

func (c *Container) Teardown() error {
	client := getDockerClient()

	err := client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    c.Name,
		Force: true,
	})

	if err != nil {
		return err
	}

	return nil
}
