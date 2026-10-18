// Command helm-diff diffs a chart's rendered manifests between BASE_REF and the current checkout.
package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// options configures a diff. Paths are relative to the repository root.
type options struct {
	Chart   string
	BaseRef string
	Values  []string // values files; ones missing from a checkout are skipped there
	Args    []string // extra arguments passed to helm template
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "helm-diff:", err)
		os.Exit(1)
	}
}

func run() error {
	opts := options{
		Chart:   os.Getenv("CHART"),
		BaseRef: os.Getenv("BASE_REF"),
		Values:  strings.Fields(os.Getenv("VALUES")),
		Args:    strings.Fields(os.Getenv("ARGS")),
	}
	if opts.Chart == "" || opts.BaseRef == "" {
		return errors.New("CHART and BASE_REF must be set")
	}
	postComment := true
	if v := os.Getenv("COMMENT"); v != "" {
		var err error
		if postComment, err = strconv.ParseBool(v); err != nil {
			return fmt.Errorf("COMMENT must be true or false, got %q", v)
		}
	}

	revParse := exec.Command("git", "rev-parse", "--show-toplevel")
	revParse.Stderr = os.Stderr
	top, err := revParse.Output()
	if err != nil {
		return err
	}

	var buf strings.Builder
	w := io.MultiWriter(&buf, os.Stdout)
	changed, err := helmDiff(strings.TrimSpace(string(top)), opts, w)
	if err != nil {
		return err
	}

	diffs := []chartDiff{
		{chartPath: opts.Chart,
			diff:       buf.String(),
			hasChanged: changed},
	}

	prNumber, err := getPrNumber()
	if err != nil {
		return err
	}
	if postComment && prNumber != 0 {
		githubAPIURL := os.Getenv("GITHUB_API_URL")
		githubRepository := os.Getenv("GITHUB_REPOSITORY")
		githubToken := os.Getenv("GITHUB_TOKEN")
		if githubAPIURL == "" || githubRepository == "" || githubToken == "" {
			return errors.New("GITHUB_API_URL, GITHUB_REPOSITORY, and GITHUB_TOKEN must be set")
		}
		comments := formatComment(diffs)
		for _, comment := range comments {
			if err := postPRComment(comment, prNumber, githubAPIURL, githubRepository, githubToken); err != nil {
				return err
			}
		}

	}
	githubOutput := os.Getenv("GITHUB_OUTPUT")
	if githubOutput == "" {
		slog.Warn("GITHUB_OUTPUT unset, not writing outputs", "changed", changed)
		return nil
	}
	return setOutput("changed", fmt.Sprint(changed), githubOutput)
}

// helmDiff writes a unified diff of the chart's rendered manifests between opts.BaseRef and the
// checkout at root to w, and reports whether they differ.
func helmDiff(root string, opts options, w io.Writer) (bool, error) {
	tmp, err := os.MkdirTemp("", "helm-diff")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmp)

	base := filepath.Join(tmp, "base")
	cleanup, err := checkoutBase(root, opts.BaseRef, base)
	if err != nil {
		return false, err
	}
	defer cleanup()

	findCharts(base, "testdata")
	baseYAML, headYAML := filepath.Join(tmp, "base.yaml"), filepath.Join(tmp, "head.yaml")
	if err := render(base, opts, baseYAML); err != nil {
		return false, err
	}
	if err := render(root, opts, headYAML); err != nil {
		return false, err
	}

	diff := command(root, "diff", "-u", "--label", opts.BaseRef, "--label", "HEAD", baseYAML, headYAML)
	diff.Stdout = w
	err = diff.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
		fmt.Fprintln(w, "No changes.")
		return false, nil
	case errors.As(err, &exit) && exit.ExitCode() == 1:
		return true, nil
	default:
		return false, err
	}
}

// render writes the chart's manifests in the checkout at dir to out; out is empty if the chart doesn't exist there.
func render(dir string, opts options, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	chart := filepath.Join(dir, opts.Chart)
	if !exists(filepath.Join(chart, "Chart.yaml")) {
		return nil
	}
	args := []string{"template", chart, "--no-hooks"}
	for _, v := range opts.Values {
		if p := filepath.Join(dir, v); exists(p) {
			args = append(args, "--values", p)
		}
	}
	cmd := command(dir, "helm", append(args, opts.Args...)...)
	cmd.Stdout = f
	return cmd.Run()
}

func setOutput(name, value string, githubOutput string) error {
	f, err := os.OpenFile(githubOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=%s\n", name, value)
	return err
}

func command(dir, name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
