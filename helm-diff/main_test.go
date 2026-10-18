package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelmDiff(t *testing.T) {
	tests := []struct {
		name        string
		opts        options
		edit        func(t *testing.T, dir string) // changes the head checkout before diffing
		wantChanged bool
		wantOutput  []string // substrings of the printed diff
		wantErr     bool
	}{
		{
			name:       "unchanged chart",
			opts:       options{ChartsDir: "chart"},
			wantOutput: []string{"No changes."},
		},
		{
			name:        "modified values",
			opts:        options{ChartsDir: "chart"},
			edit:        writeFile("chart/values.yaml", "message: changed\n"),
			wantChanged: true,
			wantOutput:  []string{"--- main\n+++ HEAD\n", `-  message: "hello"`, `+  message: "changed"`},
		},
		{
			name: "modified template",
			opts: options{ChartsDir: "chart"},
			edit: writeFile("chart/templates/service.yaml",
				"apiVersion: v1\nkind: Service\nmetadata:\n  name: {{ .Release.Name }}\n"),
			wantChanged: true,
			wantOutput:  []string{"+kind: Service"},
		},
		{
			name:        "chart missing on base",
			opts:        options{ChartsDir: "new-chart"},
			edit:        copyDir("chart", "new-chart"),
			wantChanged: true,
			wantOutput:  []string{"@@ -0,0 +1,", "+kind: ConfigMap"},
		},
		{
			name:        "chart removed on head",
			opts:        options{ChartsDir: "chart"},
			edit:        removeAll("chart"),
			wantChanged: true,
			wantOutput:  []string{"+0,0 @@", "-kind: ConfigMap"},
		},
		{
			name:       "args apply to both sides",
			opts:       options{ChartsDir: "chart", Args: []string{"--set", "message=override"}},
			edit:       writeFile("chart/values.yaml", "message: changed\n"),
			wantOutput: []string{"No changes."},
		},
		{
			name:    "unknown base ref",
			opts:    options{ChartsDir: "chart", BaseRef: "does-not-exist"},
			wantErr: true,
		},
		{
			name:    "invalid template",
			opts:    options{ChartsDir: "chart"},
			edit:    writeFile("chart/templates/configmap.yaml", "{{ .Values.message"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := newRepo(t)
			if tt.edit != nil {
				tt.edit(t, dir)
			}
			if tt.opts.BaseRef == "" {
				tt.opts.BaseRef = "main"
			}

			var out bytes.Buffer
			changed, err := helmDiff(dir, tt.opts, &out)
			if (err != nil) != tt.wantErr {
				t.Fatalf("helmDiff() error = %v, wantErr %v", err, tt.wantErr)
			}
			if changed != tt.wantChanged {
				t.Errorf("helmDiff() changed = %v, want %v", changed, tt.wantChanged)
			}
			for _, want := range tt.wantOutput {
				if !strings.Contains(out.String(), want) {
					t.Errorf("output missing %q:\n%s", want, out.String())
				}
			}
			if worktrees := strings.Count(git(t, dir, "worktree", "list", "--porcelain"), "worktree "); worktrees != 1 {
				t.Errorf("base worktree not removed: %d worktrees", worktrees)
			}
		})
	}
}

func TestRun(t *testing.T) {
	dir := newRepo(t)
	output := filepath.Join(t.TempDir(), "output")
	event := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(event, []byte(`{"ref": "refs/heads/main"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(dir, "chart"))
	t.Setenv("CHARTS_DIR", "chart")
	t.Setenv("BASE_REF", "main")
	t.Setenv("ARGS", "--namespace prod")
	t.Setenv("GITHUB_OUTPUT", output)
	t.Setenv("GITHUB_EVENT_PATH", event) // a push event, so run() doesn't comment
	t.Setenv("GITHUB_API_URL", "https://api.github.invalid")

	if err := run(); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "changed=false\n" {
		t.Errorf("GITHUB_OUTPUT = %q, want %q", got, "changed=false\n")
	}
}

func TestRunLocally(t *testing.T) {
	dir := newRepo(t)
	t.Chdir(filepath.Join(dir, "chart"))
	t.Setenv("CHARTS_DIR", "chart")
	t.Setenv("BASE_REF", "main")
	t.Setenv("GITHUB_OUTPUT", "")
	t.Setenv("GITHUB_EVENT_PATH", "")

	if err := run(); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestRunRequiresInputs(t *testing.T) {
	t.Setenv("CHARTS_DIR", "")
	t.Setenv("BASE_REF", "main")
	if err := run(); err == nil {
		t.Error("run() succeeded without CHARTS_DIR")
	}
}

// newRepo commits testdata to the main branch of an origin repository and returns a clone of it.
func newRepo(t *testing.T) string {
	t.Helper()
	origin, clone := t.TempDir(), t.TempDir()
	if err := os.CopyFS(origin, os.DirFS("testdata")); err != nil {
		t.Fatal(err)
	}
	git(t, origin, "init", "--quiet", "--initial-branch=main")
	git(t, origin, "add", ".")
	git(t, origin, "commit", "--quiet", "--message=Initial commit")
	git(t, clone, "clone", "--quiet", origin, ".")
	return clone
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func writeFile(name, content string) func(*testing.T, string) {
	return func(t *testing.T, dir string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func removeAll(name string) func(*testing.T, string) {
	return func(t *testing.T, dir string) {
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
}

func copyDir(src, dst string) func(*testing.T, string) {
	return func(t *testing.T, dir string) {
		if err := os.CopyFS(filepath.Join(dir, dst), os.DirFS(filepath.Join(dir, src))); err != nil {
			t.Fatal(err)
		}
	}
}
