package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Exercise the actual composite-action script without downloading or executing binaries.
func TestBinaryActionDownload(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required to test the composite action")
	}
	data, err := os.ReadFile("binary/action.yml")
	if err != nil {
		t.Fatal(err)
	}
	var action struct {
		Runs struct {
			Steps []struct{ Run string }
		}
	}
	if err := yaml.Unmarshal(data, &action); err != nil {
		t.Fatal(err)
	}
	script := strings.NewReplacer("${{ runner.os }}", "Linux", "${{ runner.arch }}", "X64").Replace(action.Runs.Steps[0].Run)
	const sha = "a7c644e341797d050d1cb24332d83791b9a46dae"
	tests := []struct {
		name, ref, scenario, want string
		wantError                 bool
	}{
		{"release tag", "v1.12", "tag", "Downloading render-template_linux_amd64", false},
		{"floating tag", "v1", "floating", "Resolved to v1.12", false},
		{"released SHA", sha, "sha", "Resolved commit " + sha + " to release v1.12", false},
		{"uppercase SHA", strings.ToUpper(sha), "sha", "Resolved commit " + sha + " to release v1.12", false},
		{"no matching tag", sha, "no-tags", "No release tag points to commit", true},
		{"tag without release", sha, "no-release", "No release tag points to commit", true},
		{"API failure", sha, "api-error", "API unavailable", true},
		{"pinned asset unavailable", sha, "asset-error", "Failed to download", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			mock := `#!/usr/bin/env bash
set -eu
echo "$*" >> "$CALL_LOG"
case "$1 $2" in
  "api --paginate")
    [ "$3" = "repos/chuhlomin/render-template/tags?per_page=100" ]
    [ "$4" = "--jq" ]
    [ "$5" = '.[] | select(.commit.sha == "a7c644e341797d050d1cb24332d83791b9a46dae") | .name' ]
    case "$SCENARIO" in
      api-error) echo 'API unavailable' >&2; exit 1 ;;
      no-tags) exit 0 ;;
      *) printf 'v1\nv1.12\n' ;;
    esac ;;
  "release view")
    [ "$SCENARIO" != "no-release" ] && [ "$3" = "v1.12" ] ;;
  "release list")
    [ "$SCENARIO" = "floating" ]
    echo v1.12 ;;
  "release download")
    [ "$SCENARIO" != "asset-error" ]
    [ "$3" = "v1.12" ]
    touch "$RUNNER_TEMP/render-template/render-template_linux_amd64" ;;
  *) exit 1 ;;
esac
`
			if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(mock), 0o755); err != nil {
				t.Fatal(err)
			}
			logPath := filepath.Join(dir, "calls")
			outputPath := filepath.Join(dir, "github-env")
			cmd := exec.Command(bash, "--noprofile", "--norc", "-e", "-o", "pipefail", "-c", script)
			cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"ACTION_REPO=chuhlomin/render-template", "ACTION_REF="+tt.ref,
				"RUNNER_TEMP="+dir, "GITHUB_ENV="+outputPath, "SCENARIO="+tt.scenario, "CALL_LOG="+logPath)
			output, err := cmd.CombinedOutput()
			if (err != nil) != tt.wantError || !strings.Contains(string(output), tt.want) {
				t.Fatalf("error = %v, wantError = %v; output:\n%s", err, tt.wantError, output)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if len(tt.ref) == 40 && strings.Contains(string(calls), "release list") {
				t.Fatal("SHA pin must never fall back to a floating release")
			}
			env, err := os.ReadFile(outputPath)
			if tt.wantError {
				if !os.IsNotExist(err) {
					t.Fatal("failed download must not export a binary path")
				}
			} else if err != nil || !strings.Contains(string(env), "RENDER_TEMPLATE_BIN="+dir+"/render-template/render-template_linux_amd64") {
				t.Fatalf("unexpected binary path: %s, error: %v", env, err)
			}
		})
	}
}
