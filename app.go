package main

import (
	"flag"
	"fmt"
	"log"
	"media-server/src/api"
	"media-server/src/models"
)

type MODE string

const (
	MODE_ENV  MODE = "env"
	MODE_YAML MODE = "yaml"
)

func main() {
	modeFlag := flag.String("mode", "env", "Mode of configs loadings")
	configFlag := flag.String("config", "config.yaml", "Config destination (.yml)")
	flag.Parse()
	var (
		config *models.Config
		err    error
	)
	switch *modeFlag {
	case string(MODE_YAML):
		log.Println("Usings .yaml file for configuration")
		if configFlag == nil {
			log.Panic("Config folder is not setted in arguments!")
		}
		if config, err = models.NewConfigYML(*configFlag); err != nil {
			log.Panic(fmt.Sprintf("Error while configs reading!\nError: %s", err.Error()))
		}
	case string(MODE_ENV):
		log.Println("Loading configuration from ENV")
		if config, err = models.NewConfigENV(); err != nil {
			log.Panic(fmt.Sprintf("Error while configs readings from envs!\nError: %s", err.Error()))
		}
	}
	api, _ := api.NewMediaAPI(&config.MediaAPIConfig)
	api.Run()
}
