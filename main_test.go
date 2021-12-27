package main

import (
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
