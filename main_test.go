package main

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestEscape(t *testing.T) {
	tests := []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"text", "text"},
		{"%", "%25"},
		{"\n", "%0A"},
		{"\r", "%0D"},
	}

	for _, tt := range tests {
		actual := escape(tt.in)
		if actual != tt.expected {
			t.Errorf("escape(%q) was incorrect, got: %q, want: %q.", tt.in, actual, tt.expected)
		}
	}
}

func TestVarsParser(t *testing.T) {
	tests := []struct {
		in           string
		expectedJSON string
		err          error
	}{
		{
			`
key: value`,
			`{"key":"value"}`,
			nil,
		},
		{
			`
key: |
  value`,
			`{"key":"value"}`,
			nil,
		},
		{
			`
key: |
  line 1
  line 2`,
			`{"key":"line 1\nline 2"}`,
			nil,
		},
		{
			`
key: |
  line 1: val1
line 2: val2
  line 3: val3`,
			``,
			errors.New("unable to parse Vars: yaml: line 5: mapping values are not allowed in this context"),
		},
		{
			`
key: "line 1: val1
line 2: val2
  line 3: val3"`,
			`{"key":"line 1: val1 line 2: val2 line 3: val3"}`,
			nil,
		},
		{
			`{"key": "val1"}`,
			`{"key":"val1"}`,
			nil,
		},
	}

	for _, tt := range tests {
		actual, err := varsParser(tt.in)
		if tt.err != nil {
			if err == nil {
				t.Errorf("varsParser(%q) was incorrect, got: nil, want: %q.", tt.in, tt.err)
			}
			if err.Error() != tt.err.Error() {
				t.Errorf("varsParser(%q) was incorrect, got: %q, want: %q.", tt.in, err, tt.err)
			}
			continue
		}

		actualJSONBytes, err := json.Marshal(actual)
		if err != nil {
			t.Errorf("varsParser (%q) failed to marshal actual %v", tt.in, actual)
		}

		if string(actualJSONBytes) != tt.expectedJSON {
			t.Errorf("varsParser(%q) was incorrect, got: %q, want: %q.", tt.in, string(actualJSONBytes), tt.expectedJSON)
		}
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		templateFilePath string
		vars             vars
		expectedError    error
		expectedOutput   string
	}{
		{
			"./testdata/template.txt",
			map[string]interface{}{
				"name": "world",
			},
			nil,
			"Hello world\n",
		},
		{
			"./testdata/missing.txt",
			map[string]interface{}{},
			errors.New("template file not found (\"./testdata/missing.txt\")"),
			"",
		},
		{
			"./testdata/template.txt",
			map[string]interface{}{},
			errors.New("template: ./testdata/template.txt:1:9: executing \"./testdata/template.txt\" at <.name>: map has no entry for key \"name\""),
			"",
		},
		{
			"./testdata/invalid.txt",
			map[string]interface{}{
				"name": "world",
			},
			errors.New("template: ./testdata/invalid.txt:1: missing value for if"),
			"",
		},
		{
			"./testdata/template.txt",
			map[string]interface{}{
				"name": "text+text",
			},
			nil,
			"Hello text+text\n",
		},
	}

	for _, tt := range tests {
		output, err := renderTemplate(tt.templateFilePath, tt.vars)
		if err != nil {
			if tt.expectedError == nil {
				t.Errorf("renderTemplate(%q, %v) returned an error, but was expected to succeed: %v", tt.templateFilePath, tt.vars, err)
			} else if err.Error() != tt.expectedError.Error() {
				t.Errorf(
					"render(%q, %v) expected error: %q, got: %q",
					tt.templateFilePath,
					tt.vars,
					tt.expectedError,
					err,
				)
			}
		} else if tt.expectedError != nil {
			t.Errorf("renderTemplate(%q, %v) succeeded, but was expected to fail: %v", tt.templateFilePath, tt.vars, err)
		} else if output != tt.expectedOutput {
			t.Errorf(
				"render(%q, %v) expected output: %q, got: %q",
				tt.templateFilePath,
				tt.vars,
				tt.expectedOutput,
				output,
			)
		}
	}
}
