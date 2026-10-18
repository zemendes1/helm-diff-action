// Command helm-diff diffs a chart's rendered manifests between BASE_REF and the current checkout.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// options configures a diff. Paths are relative to the repository root.
type options struct {
	ChartsDir string
	BaseRef   string
	Args      []string // extra arguments passed to helm template
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "helm-diff:", err)
		os.Exit(1)
	}
}

func run() error {
	opts := options{
		ChartsDir: os.Getenv("CHARTS_DIR"),
		BaseRef:   os.Getenv("BASE_REF"),
		Args:      strings.Fields(os.Getenv("ARGS")),
	}
	if opts.ChartsDir == "" || opts.BaseRef == "" {
		return errors.New("CHARTS_DIR and BASE_REF must be set")
	}
	postComment := true
	if v := os.Getenv("COMMENT"); v != "" {
		var err error
		if postComment, err = strconv.ParseBool(v); err != nil {
			return fmt.Errorf("COMMENT must be true or false, got %q", v)
		}
	}

	root, err := repoRoot()
	if err != nil {
		return err
	}

	diffs, err := helmDiff(root, opts)
	if err != nil {
		return err
	}

	anyChartChanged := slices.ContainsFunc(diffs, func(d chartDiff) bool { return d.hasChanged })

	prNumber, err := getPrNumber()
	if err != nil {
		return err
	}
	if postComment && prNumber != 0 && anyChartChanged {
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
		slog.Warn("GITHUB_OUTPUT unset, not writing outputs", "changed", anyChartChanged)
		return nil
	}
	return setOutput("changed", fmt.Sprint(anyChartChanged), githubOutput)
}

// helmDiff diffs the rendered manifests of every chart under opts.ChartsDir between opts.BaseRef and
// the checkout at root: one chartDiff per values-*.yaml file, or one for the defaults if there are none.
func helmDiff(root string, opts options) ([]chartDiff, error) {
	tmp, err := os.MkdirTemp("", "helm-diff")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	base := filepath.Join(tmp, "base")
	cleanup, err := checkoutBase(root, opts.BaseRef, base)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	baseCharts, err := findCharts(base, opts.ChartsDir)
	if err != nil {
		return nil, err
	}

	headCharts, err := findCharts(root, opts.ChartsDir)
	if err != nil {
		return nil, err
	}

	var chartDiffs []chartDiff
	baseYAML, headYAML := filepath.Join(tmp, "base.yaml"), filepath.Join(tmp, "head.yaml")

	for _, chart := range merge(baseCharts, headCharts) {
		valuesFiles, err := chartValuesFiles(base, root, chart)
		if err != nil {
			return nil, err
		}
		for _, values := range valuesFiles {
			if err := render(base, chart, values, opts, baseYAML); err != nil {
				return nil, err
			}
			if err := render(root, chart, values, opts, headYAML); err != nil {
				return nil, err
			}
			diff, hasChanged, err := diffFiles(root, opts.BaseRef, baseYAML, headYAML)
			if err != nil {
				return nil, err
			}
			chartDiffs = append(chartDiffs, chartDiff{chartPath: chart, values: values, diff: diff, hasChanged: hasChanged})
		}
	}

	return chartDiffs, nil
}

// diffFiles returns a unified diff of baseFile and headFile, labelling the base side baseLabel,
// and reports whether they differ.
func diffFiles(dir, baseLabel, baseFile, headFile string) (string, bool, error) {
	var diff strings.Builder

	diffCommand := command(dir, "diff", "-u", "--label", baseLabel, "--label", "HEAD", baseFile, headFile)
	diffCommand.Stdout = &diff

	err := diffCommand.Run()
	var exit *exec.ExitError

	switch {
	case err == nil:
		return "", false, nil
	case errors.As(err, &exit) && exit.ExitCode() == 1:
		return diff.String(), true, nil // the diff command printed the differences
	default:
		return "", false, err
	}
}

// render writes the chart's manifests in the checkout at dir to out, using the chart's values file if
// values isn't empty; out is empty if the chart or the values file doesn't exist there.
func render(dir, chart, values string, opts options, out string) error {
	f, err := os.Create(out)
	if err != nil {
		return err
	}
	defer f.Close()

	chartPath := filepath.Join(dir, chart)
	if !exists(filepath.Join(chartPath, "Chart.yaml")) {
		return nil
	}
	args := []string{"template", chartPath, "--no-hooks"}
	if values != "" {
		valuesPath := filepath.Join(chartPath, values)
		if !exists(valuesPath) {
			return nil
		}
		args = append(args, "--values", valuesPath)
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
