package dockerstrap

import (
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"sync"
)

func getDockerTLSClient() *docker.Client {
	endpoint := os.Getenv("DOCKER_HOST")
	cert := os.Getenv("DOCKER_CERT_PATH") + "/cert.pem"
	key := os.Getenv("DOCKER_CERT_PATH") + "/key.pem"
	ca := os.Getenv("DOCKER_CERT_PATH") + "/ca.pem"

	if endpoint == "" {
		log.Fatal("Docker not started!")
	}

	client, err := docker.NewTLSClient(endpoint, cert, key, ca)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func getDockerClient() *docker.Client {
	endpoint := os.Getenv("DOCKER_HOST")

	// Check if docker is in TLS mode
	tls_verify := os.Getenv("DOCKER_TLS_VERIFY")

	if endpoint == "" {
		log.Fatal("Docker not started!")
	}

	if tls_verify == "1" {
		return getDockerTLSClient()
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func StartContainers(containers []*Container) {
	var wg sync.WaitGroup

	// Add all the container names
	wg.Add(len(containers))

	for _, container := range containers {
		log.Printf("Now starting container %v\n", container)

		go setupContainer(&wg, container)
	}

	wg.Wait()
}

func setupContainer(wg *sync.WaitGroup, container *Container) {
	// Call the setup method
	err := container.Setup()
	if err != nil {
		log.Fatal(err)
	}

	// Log some stuff
	log.Printf("Container \"%v\" has been started.\n", container)

	// Say we're all done
	wg.Done()
}

// Stop and remove all the containers listed
func TeardownContainers(containers []*Container) {
	var wg sync.WaitGroup

	// Add all the container names
	wg.Add(len(containers))

	for _, container := range containers {
		log.Printf("Now tearing down container %v\n", container)

		go teardownContainer(&wg, container)
	}

	wg.Wait()
}

func teardownContainer(wg *sync.WaitGroup, container *Container) {
	// Teardown the container
	err := container.Teardown()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Container \"%v\" has been torn down.\n", container)

	wg.Done()
}

func RefreshContainers(containers []*Container) {
	var wg sync.WaitGroup

	// Add all the container names
	wg.Add(len(containers))

	for _, container := range containers {
		log.Printf("Now refreshing container %v\n", container)

		go refreshContainer(&wg, container)
	}

	wg.Wait()
}

func refreshContainer(wg *sync.WaitGroup, container *Container) {
	err, isUp := container.isUp(getDockerClient())
	if err != nil {
		log.Fatal(err)
	}

	if isUp {
		log.Printf("Container %v is up\n", container)
		log.Printf("Tearing down %v\n", container)
		// Teardown the container
		err = container.Teardown()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Container %v is not up\n", container)
	}

	log.Printf("Setting up %v\n", container)

	// Call the setup method
	err = container.Setup()
	if err != nil {
		log.Fatal(err)
	}

	wg.Done()
}
