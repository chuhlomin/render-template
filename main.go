package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"text/template"
	"time"
	_ "time/tzdata"

	"github.com/caarlos0/env/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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
		varsFile, err := os.ReadFile(c.VarsPath)
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

	if err = writeOutput(output); err != nil {
		return err
	}

	if len(c.ResultPath) != 0 {
		err := os.WriteFile(c.ResultPath, []byte(output), 0644)
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

var funcMap = template.FuncMap{
	"date": func(format string, in interface{}) string {
		var t time.Time
		switch v := in.(type) {
		case string:
			var err error
			t, err = time.Parse(time.RFC3339, v)
			if err != nil {
				log.Printf("failed to parse date %q: %v", v, err)
				return v
			}
		case time.Time:
			t = v
		default:
			log.Printf("unsupported type %T for date", in)
			return fmt.Sprintf("%v", in)
		}

		timezone := os.Getenv("INPUT_TIMEZONE")
		if timezone != "" {
			loc, err := time.LoadLocation(timezone)
			if err != nil {
				log.Printf("failed to load timezone %q: %v", timezone, err)
				return in.(string)
			}
			t = t.In(loc)
		}

		return t.Format(format)
	},
	"mdlink": func(text, url string) string {
		return fmt.Sprintf("[%s](%s)", text, url)
	},
	"number": func(in string) string {
		p := message.NewPrinter(language.English)
		d, err := strconv.ParseInt(in, 10, 64)
		if err != nil {
			log.Printf("failed to parse number %q: %v", in, err)
			return in
		}
		return p.Sprintf("%d", d)
	},
	"base64": func(in string) string {
		return base64.StdEncoding.EncodeToString([]byte(in))
	},
}

func renderTemplate(templateFilePath string, vars vars) (string, error) {
	b, err := os.ReadFile(templateFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("template file not found (%q)", templateFilePath)
		}
		if errors.Is(err, os.ErrPermission) {
			return "", fmt.Errorf("have no permissions to read template file (%q)", templateFilePath)
		}
		return "", fmt.Errorf("failed to read template %q: %v", templateFilePath, err)
	}

	tmpl, err := template.
		New(templateFilePath).
		Option("missingkey=error").
		Funcs(funcMap).
		Parse(string(b))
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	if err = tmpl.Execute(&result, vars); err != nil {
		return "", err
	}

	return result.String(), nil
}

func writeOutput(output string) error {
	githubOutput := formatOutput("result", output)
	if githubOutput == "" {
		return nil
	}

	path := os.Getenv("GITHUB_OUTPUT")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf(
			"failed to open result file %q: %v. "+
				"If you are using self-hosted runners "+
				"make sure they are updated to version 2.297.0 or greater",
			path,
			err,
		)
	}
	defer f.Close()

	if _, err = f.WriteString(githubOutput); err != nil {
		return fmt.Errorf("failed to write result to file %q: %v", path, err)
	}

	return nil
}

func formatOutput(name, value string) string {
	if value == "" {
		return ""
	}

	// if value contains new line, use multiline format
	if bytes.ContainsRune([]byte(value), '\n') {
		return fmt.Sprintf("%s<<OUTPUT\n%s\nOUTPUT", name, value)
	}

	return fmt.Sprintf("%s=%s", name, value)
}
