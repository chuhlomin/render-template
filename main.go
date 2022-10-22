package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

type vars map[string]interface{}

type config struct {
	Template   string `env:"INPUT_TEMPLATE" envDefault:".kube.yml"`
	Vars       vars   `env:"INPUT_VARS" envDefault:""`
	VarsPath   string `env:"INPUT_VARS_PATH" envDefault:""`
	ResultPath string `env:"INPUT_RESULT_PATH" envDefault:""`
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("::error::%v", err)
		os.Exit(1)
	}
}

func run() error {
	var c config
	parsers := map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(vars{}): varsParser,
	}
	if err := env.ParseWithFuncs(&c, parsers); err != nil {
		return err
	}

	if c.VarsPath != "" {
		varsFile, err := ioutil.ReadFile(c.VarsPath)
		if err != nil {
			return fmt.Errorf("failed to read vars file %q: %v", c.VarsPath, err)
		}
		var varsFromFile vars
		if err = yaml.Unmarshal(varsFile, &varsFromFile); err != nil {
			return fmt.Errorf("failed to parse vars file %q: %v", c.VarsPath, err)
		}
		c.Vars = mergeVars(c.Vars, varsFromFile)
	}

	output, err := renderTemplate(c.Template, c.Vars)
	if err != nil {
		return fmt.Errorf("failed to render template: %v", err)
	}

	githubOutput := fmt.Sprintf("%s=%s", "result", escape(output))
	if os.Getenv("GITHUB_OUTPUT") != "" {
		githubOutput = fmt.Sprintln(os.Getenv("GITHUB_OUTPUT")) + githubOutput
	}

	err = os.Setenv("GITHUB_OUTPUT", githubOutput)
	if err != nil {
		return fmt.Errorf("failed to set GITHUB_OUTPUT: %v", err)
	}

	fmt.Printf("::debug::%s", os.Getenv("GITHUB_OUTPUT"))

	if len(c.ResultPath) != 0 {
		err := ioutil.WriteFile(c.ResultPath, []byte(output), 0644)
		if err != nil {
			return fmt.Errorf("failed to write file %q: %v", c.ResultPath, err)
		}
	}

	return nil
}

func varsParser(v string) (interface{}, error) {
	m := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(v), &m)
	if err != nil {
		return nil, fmt.Errorf("unable to parse Vars: %v", err)
	}
	return m, nil
}

func mergeVars(a, b vars) vars {
	if a == nil {
		return b
	}

	for k, v := range b {
		if _, ok := a[k]; ok {
			continue
		}
		a[k] = v
	}
	return a
}

func renderTemplate(templateFilePath string, vars vars) (string, error) {
	b, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("template file not found (%q)", templateFilePath)
		}
		if errors.Is(err, os.ErrPermission) {
			return "", fmt.Errorf("have no permissions to read template file (%q)", templateFilePath)
		}
		return "", fmt.Errorf("failed to read template %q: %v", templateFilePath, err)
	}

	tmpl, err := template.New(templateFilePath).Option("missingkey=error").Parse(string(b))
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	if err = tmpl.Execute(&result, vars); err != nil {
		return "", err
	}

	return result.String(), nil
}

func escape(str string) string {
	/*
		set-output truncates multiline strings.
		% and \n and \r can be escaped, the runner will unescape in reverse:
		https://github.community/t/set-output-truncates-multiline-strings/16852
	*/

	return strings.NewReplacer(
		"%", "%25",
		"\n", "%0A",
		"\r", "%0D",
	).Replace(str)
}
