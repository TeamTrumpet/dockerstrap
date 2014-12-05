package main

import (
	"github.com/TeamTrumpet/dockerstrap"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type ConfigFile struct {
	Containers []*dockerstrap.Container `yaml:"containers"`
}

func readConfig(filename string) ConfigFile {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var config ConfigFile

	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return config
}

func main() {

	app := cli.NewApp()
	app.Name = "dockerstrap"
	app.Usage = "Bootstrap a testing environment using docker containers."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "testenv-file, f",
			Value:  "testenv.yaml",
			EnvVar: "TEST_ENV_FILE",
			Usage:  "file containing the test environment configuration",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "start",
			ShortName: "s",
			Usage:     "start the listed docker containers",
			Action: func(c *cli.Context) {
				config := readConfig(c.GlobalString("f"))
				dockerstrap.StartContainers(config.Containers)
			},
		},
		{
			Name:      "clean",
			ShortName: "c",
			Usage:     "removes the listed docker containers",
			Action: func(c *cli.Context) {
				config := readConfig(c.GlobalString("f"))
				dockerstrap.TeardownContainers(config.Containers)
			},
		},
		{
			Name:      "refresh",
			ShortName: "r",
			Usage:     "removes running containers and starts them up",
			Action: func(c *cli.Context) {
				config := readConfig(c.GlobalString("f"))
				dockerstrap.RefreshContainers(config.Containers)
			},
		},
	}

	app.Run(os.Args)

}
