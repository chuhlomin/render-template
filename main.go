package main

import (
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type config struct {
	Image              string                 `env:"INPUT_IMAGE,required"`
	Template           string                 `env:"INPUT_TEMPLATE" envDefault:".kube.yml"`
	SkipTemplate       bool                   `env:"INPUT_SKIP_TEMPLATE" envDefault:"false"`
	SecretTemplate     string                 `env:"INPUT_SECRET_TEMPLATE" envDefault:".kube.sec.yml"`
	SkipSecretTemplate bool                   `env:"INPUT_SKIP_SECRET_TEMPLATE" envDefault:"false"`
	WaitDeployments    []string               `env:"INPUT_WAIT_DEPLOYMENTS" envDefault:""`
	WaitDuration       time.Duration          `env:"INPUT_WAIT_DURATION" envDefault:"0"`
	Vars               map[string]interface{} `env:"INPUT_VARS" envDefault:""`
}

func main() {
	log.Printf("Starting...")

	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	log.Printf("Stopped")
}

func run() error {
	var c config
	if err := env.Parse(&c); err != nil {
		return err
	}

	log.Printf("Config: %#v", c)

	return nil
}
