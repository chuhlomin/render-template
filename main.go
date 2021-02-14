package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type vars map[string]interface{}

type config struct {
	Image              string        `env:"INPUT_IMAGE,required"`
	Template           string        `env:"INPUT_TEMPLATE" envDefault:".kube.yml"`
	SkipTemplate       bool          `env:"INPUT_SKIP_TEMPLATE" envDefault:"false"`
	SecretTemplate     string        `env:"INPUT_SECRET_TEMPLATE" envDefault:".kube.sec.yml"`
	SkipSecretTemplate bool          `env:"INPUT_SKIP_SECRET_TEMPLATE" envDefault:"false"`
	WaitDeployments    []string      `env:"INPUT_WAIT_DEPLOYMENTS" envDefault:""`
	WaitDuration       time.Duration `env:"INPUT_WAIT_DURATION" envDefault:"0"`
	Vars               vars          `env:"INPUT_VARS" envDefault:""`
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
	parsers := map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(vars{}): func(v string) (interface{}, error) {
			m := map[string]interface{}{}
			err := yaml.Unmarshal([]byte(v), &m)
			if err != nil {
				return nil, fmt.Errorf("unable to parse Vars: %v", err)
			}
			return m, nil
		},
	}
	if err := env.ParseWithFuncs(&c, parsers); err != nil {
		return err
	}

	return nil
}
