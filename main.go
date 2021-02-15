package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

const outputManifestPath = "/tmp/kube.yml"

type Vars map[string]interface{}

type config struct {
	Template string `env:"INPUT_TEMPLATE" envDefault:".kube.yml"`
	Vars     Vars   `env:"INPUT_VARS" envDefault:""`
}

func main() {
	log.Printf("Starting...")

	if err := run(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	log.Printf("Finished")
}

func run() error {
	var c config
	parsers := map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(Vars{}): func(v string) (interface{}, error) {
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

	if err := renderTemplate(outputManifestPath, c.Template, c.Vars); err != nil {
		return err
	}

	if err := applyManifest(outputManifestPath); err != nil {
		return err
	}

	return nil
}

func renderTemplate(outputFilePath string, templateFilePath string, vars Vars) error {
	_, err := os.Stat(templateFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("template file not found (%q)", templateFilePath)
	}

	b, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return fmt.Errorf("failed to read template %q: %v", templateFilePath, err)
	}

	tmpl, err := template.New(".kube.yml").Option("missingkey=error").Parse(string(b))
	if err != nil {
		return fmt.Errorf("failed to parse template %q: %s", templateFilePath, err)
	}

	f, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to open output file %q: %v", outputFilePath, err)
	}

	if err = tmpl.Execute(f, vars); err != nil {
		return fmt.Errorf("failed to render manifest from template %q: %v", templateFilePath, err)
	}

	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close file %q: %v", outputFilePath, err)
	}

	return nil
}

func applyManifest(manifestFilePath string) error {
	return runCommand(
		cmdKubectl,
		[]string{
			"apply",
			"-f",
			manifestFilePath,
		},
	)
}
