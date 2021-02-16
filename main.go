package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type vars map[string]interface{}

type config struct {
	Template string `env:"INPUT_TEMPLATE" envDefault:".kube.yml"`
	Vars     vars   `env:"INPUT_VARS" envDefault:""`
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("echo \"::error::%v\"", err)
		os.Exit(1)
	}
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

	return renderTemplate(c.Template, c.Vars)
}

func renderTemplate(templateFilePath string, vars vars) error {
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

	var result bytes.Buffer
	if err = tmpl.Execute(&result, vars); err != nil {
		return fmt.Errorf("failed to render template %q: %v", templateFilePath, err)
	}

	fmt.Printf("echo \"::set-output name=result::%s\"", result.String())
	return nil
}
